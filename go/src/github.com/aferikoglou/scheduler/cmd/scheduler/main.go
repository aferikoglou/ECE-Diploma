package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"time"
	"sort"
	"errors"
	"strconv"
	"container/list"
	"github.com/jupp0r/go-priority-queue"
	"github.com/promqueriesutil"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	listersv1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/apimachinery/pkg/util/wait"
)

const schedulerName = "custom-scheduler"

var AVAIL_GPU_MEM int = 32

var podGPUMemoryRequestMap map[string]int

var runningPodsList *list.List
var tmpList *list.List

var FOUND_SCHEDULABLE_POD bool = true

const PROMETHEUS_ENDPOINT string = "http://192.168.1.145:30000"
const d int = 50

const k int = 1

type predicateFunc func(node *v1.Node, pod *v1.Pod) bool
type priorityFunc func(node *v1.Node, pod *v1.Pod) int

type Scheduler struct {
	clientset  *kubernetes.Clientset
	podPriorityQueue pq.PriorityQueue
	nodeLister listersv1.NodeLister
	predicates []predicateFunc
	priorities []priorityFunc
}

func NewScheduler(podPriorityQueue pq.PriorityQueue, quit chan struct{}) Scheduler {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatal(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	return Scheduler{
		clientset:  clientset,
		podPriorityQueue: podPriorityQueue,
		nodeLister: initInformers(clientset, podPriorityQueue, quit),
		predicates: []predicateFunc{
			containsGPUPredicate,
		},
		priorities: []priorityFunc{
			randomPriority,
		},
	}
}

func initInformers(clientset *kubernetes.Clientset, podPriorityQueue pq.PriorityQueue, quit chan struct{}) listersv1.NodeLister {
	factory := informers.NewSharedInformerFactory(clientset, 0)

	nodeInformer := factory.Core().V1().Nodes()

	nodeInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			node, ok := obj.(*v1.Node)
			if !ok {
				log.Println("This is not a node.")
				return
			}
			log.Printf("New Node Added to Store: %s", node.GetName())
		},
	})

	podInformer := factory.Core().V1().Pods()

	podInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			pod, ok := obj.(*v1.Pod)
			if !ok {
				log.Println("This is not a pod.")
				return
			}

			if pod.Spec.NodeName == "" && pod.Spec.SchedulerName == schedulerName {

				gpuMemory, _ := strconv.Atoi(pod.Labels["GPU_MEM_REQ"])

				podGPUMemoryRequestMap[pod.Name] = gpuMemory

				priority := float64(gpuMemory)

				fmt.Print("Found a pod to schedule=", pod.Namespace, "/", pod.Name)
				fmt.Println(" Assigned Priority=" + fmt.Sprintf("%f", priority))

				podPriorityQueue.Insert(pod, priority)

			}
		},
	})

	factory.Start(quit)

	return nodeInformer.Lister()
}

func main() {
	fmt.Println("Custom Kubernetes GPU Scheduler - Kube-Knots")

	podGPUMemoryRequestMap = make(map[string]int)

	podPriorityQueue := pq.New()

	runningPodsList = list.New()
	tmpList = list.New()

	quit := make(chan struct{})
	defer close(quit)

	scheduler := NewScheduler(podPriorityQueue, quit)

	scheduler.Run(quit)
}

func (s *Scheduler) Run(quit chan struct{}) {
	wait.Until(s.ScheduleOne, 0, quit)
}

func (s *Scheduler) ScheduleOne() {
	pod, _ := s.podPriorityQueue.Pop()

	for pod == nil {
		pod, _ = s.podPriorityQueue.Pop()

		s.removeRunListPods()

		getState()

		time.Sleep(1 * time.Second)
	}

	p := pod.(*v1.Pod)

	FOUND_SCHEDULABLE_POD = true
	tmpList = list.New()

	// Check running GPU pod list if some of these pods are finished.
	s.removeRunListPods()

	// Get GPU Timeseries
	freeGPUMemTS := getGPUMetricTS(PROMETHEUS_ENDPOINT, "dcgm_fb_free", d)
	usedGPUMemTS := getGPUMetricTS(PROMETHEUS_ENDPOINT, "dcgm_fb_used", d)

	// If the pod requests more GPU memory than the available add the pod to tempList
	// and get the next pod.
	for !Can_Be_Scheduled(freeGPUMemTS, usedGPUMemTS, p) {
		// Save pod to tmpList.
		tmpList.PushBack(p)
		// Get next pod.
		np, _ := s.podPriorityQueue.Pop()
		// Check if the Priority Queue is empty.
		if np == nil {
			FOUND_SCHEDULABLE_POD = false
			break
		}
		p = np.(*v1.Pod)
	}

	getState()

	// Add pods back to priority queue.
	for e := tmpList.Front(); e != nil; e = e.Next() {
		pod_ := e.Value.(*v1.Pod)
		s.podPriorityQueue.Insert(pod_, float64(podGPUMemoryRequestMap[pod_.Name]))
	}

	time.Sleep(1 * time.Second)

	if !FOUND_SCHEDULABLE_POD {
		return
	} else {
		// Try to schedule the selected pod.
		node, err := s.findFit(p)
		if err != nil {
			log.Println("Cannot find node that fits pod.", err.Error())

			// In case of FAILURE add pod back to the Priority Queue.
			s.podPriorityQueue.Insert(p, float64(podGPUMemoryRequestMap[p.Name]))

			return
		}

		err = s.bindPod(p, node)
		if err != nil {
			log.Println("Failed to bind pod.", err.Error())

			// In case of FAILURE add pod back to the Priority Queue.
			s.podPriorityQueue.Insert(p, float64(podGPUMemoryRequestMap[p.Name]))

			return
		}

		message := fmt.Sprintf("Placed pod [%s/%s] on %s\n", p.Namespace, p.Name, node)

		err = s.emitEvent(p, message)
		if err != nil {
			log.Println("Failed to emit scheduled event.", err.Error())

			// In case of FAILURE add pod back to the Priority Queue.
			s.podPriorityQueue.Insert(p, float64(podGPUMemoryRequestMap[p.Name]))

			return
		}

		fmt.Println(message)

		// If the pod is scheduled successfully add it to the running pods list and
		// decrease the available GPU memory.
		runningPodsList.PushBack(p)
		AVAIL_GPU_MEM -= podGPUMemoryRequestMap[p.Name]
	}
}

func (s *Scheduler) findFit(pod *v1.Pod) (string, error) {
	nodes, err := s.nodeLister.List(labels.Everything())
	if err != nil {
		return "", err
	}

	filteredNodes := s.runPredicates(nodes, pod)
	if len(filteredNodes) == 0 {
		return "", errors.New("Failed to find node that fits pod.")
	}

	priorities := s.prioritize(filteredNodes, pod)

	return s.findBestNode(priorities), nil
}

func (s *Scheduler) bindPod(p *v1.Pod, node string) error {
	return s.clientset.CoreV1().Pods(p.Namespace).Bind(&v1.Binding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      p.Name,
			Namespace: p.Namespace,
		},
		Target: v1.ObjectReference{
			APIVersion: "v1",
			Kind:       "Node",
			Name:       node,
		},
	})
}

func (s *Scheduler) emitEvent(p *v1.Pod, message string) error {
	timestamp := time.Now().UTC()
	_, err := s.clientset.CoreV1().Events(p.Namespace).Create(&v1.Event{
		Count:          1,
		Message:        message,
		Reason:         "Scheduled",
		LastTimestamp:  metav1.NewTime(timestamp),
		FirstTimestamp: metav1.NewTime(timestamp),
		Type:           "Normal",
		Source: v1.EventSource{
			Component: schedulerName,
		},
		InvolvedObject: v1.ObjectReference{
			Kind:      "Pod",
			Name:      p.Name,
			Namespace: p.Namespace,
			UID:       p.UID,
		},
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: p.Name + "-",
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *Scheduler) runPredicates(nodes []*v1.Node, pod *v1.Pod) []*v1.Node {
	filteredNodes := make([]*v1.Node, 0)
	for _, node := range nodes {
		if s.predicatesApply(node, pod) {
			filteredNodes = append(filteredNodes, node)
		}
	}

	log.Println("Nodes that fit:")
	for _, n := range filteredNodes {
		log.Println(n.Name)
	}
	return filteredNodes
}

func (s *Scheduler) predicatesApply(node *v1.Node, pod *v1.Pod) bool {
	for _, predicate := range s.predicates {
		if !predicate(node, pod) {
			return false
		}
	}
	return true
}

func containsGPUPredicate(node *v1.Node, pod *v1.Pod) bool {
	return (node.Name == "kube-gpu-1")
}

func (s *Scheduler) prioritize(nodes []*v1.Node, pod *v1.Pod) map[string]int {
	priorities := make(map[string]int)

	for _, node := range nodes {
		for _, priority := range s.priorities {
			priorities[node.Name] += priority(node, pod)
		}
	}
	log.Println("Calculated priorities:", priorities)
	return priorities
}

func (s *Scheduler) findBestNode(priorities map[string]int) string {
	var maxP int
	var bestNode string
	for node, p := range priorities {
		if p > maxP {
			maxP = p
			bestNode = node
		}
	}
	return bestNode
}

func randomPriority(node *v1.Node, pod *v1.Pod) int {
	return rand.Intn(100)
}

func getState() {
	fmt.Println(" ")
	fmt.Print("Available GPU Memory = ")
	fmt.Println(AVAIL_GPU_MEM)

        fmt.Print("Running pods         = [ ")
        for e := runningPodsList.Front(); e != nil; e = e.Next() {
		pod := e.Value.(*v1.Pod)
		fmt.Print(pod.Name + " ")
	}
	fmt.Println("]")

	fmt.Print("Temp List pods       = [ ")
	for e := tmpList.Front(); e != nil; e = e.Next() {
		pod := e.Value.(*v1.Pod)
		fmt.Print(pod.Name + " ")
	}
	fmt.Println("]")
	fmt.Println(" ")
}

func (s *Scheduler) removeRunListPods() {
	for e := runningPodsList.Front(); e != nil; e = e.Next() {
		pod_ := e.Value.(*v1.Pod)

		// resp, _ := s.clientset.CoreV1().Pods("default").Get(pod_.Name, metav1.GetOptions{})
		resp, er := s.clientset.CoreV1().Pods("default").Get(pod_.Name, metav1.GetOptions{})

		// f_ := resp.Status.Phase == "Succeeded"
		f_ := er != nil || resp.Status.Phase == "Succeeded"
		if f_ {
			gpuMemoryReq, exists := podGPUMemoryRequestMap[pod_.Name]
			if exists {
				AVAIL_GPU_MEM += gpuMemoryReq
				delete(podGPUMemoryRequestMap, pod_.Name);
			}

			runningPodsList.Remove(e)
		}
	}
}

func getGPUMetricTS(PROMETHEUS_ENDPOINT string, GPU_METRIC string, d int) map[int]float64 {
	const STEP int = 1

	endTimestamp := int32(time.Now().Unix())

	startTimestamp := endTimestamp - int32(d)

	return promqueriesutil.RangeQuery(PROMETHEUS_ENDPOINT, GPU_METRIC, int64(startTimestamp), int64(endTimestamp), STEP)
}

func Can_Colocate(freeGPUMemTS map[int]float64) bool {
	// Get map data sorted
	keys := make([]int, 0)
        for k, _ := range freeGPUMemTS {
		keys = append(keys, k)
	}

        sort.Ints(keys)

	var N int = d + 1
	var sum float64 = 0.0   // x_i summary
	var sum_2 float64 = 0.0 // x_i * x_i summary

	for _, timestamp := range keys {
		val := freeGPUMemTS[timestamp]

		sum += val
		sum_2 += val * val
	}

	fmt.Println(" ")
	fmt.Println("Can_Colocate Calculations ...")

	var mean float64 = sum / float64(N)
	fmt.Print("Mean               = ")
        fmt.Println(mean)

	mean_val_squared := sum_2 / float64(N)
	var variance float64 = mean_val_squared - mean * mean
	fmt.Print("Variance           = ")
        fmt.Println(variance)

	var standard_deviation float64 = math.Sqrt(variance)
	fmt.Print("Standard Deviation = ")
	fmt.Println(standard_deviation)

	var COV float64 = standard_deviation / mean
	fmt.Print("COV                = ")
	fmt.Println(COV)
	fmt.Println(" ")

	return (COV < 0.5)
}

func Harvest_Resource(usedGPUMemTS map[int]float64) int {
	// Calculate the 80%-ile GPU memory consumption
	values := make([]float64, 0)
        for _, val := range usedGPUMemTS {
		values = append(values, val)
	}

	fmt.Println(" ")
	fmt.Println("Harvest_Resource Calculations ...")

	sort.Float64s(values)
	fmt.Print("Sorted Values      = ")
	fmt.Println(values)

	var index int = int(float64(d) * 0.8)
	fmt.Print("Index              = ")
	fmt.Println(index)

	var percentile float64 = values[index]
	fmt.Print("80%-ile            = ")
	fmt.Println(percentile)

	var temp float64 = math.Round(percentile * 0.001)
	var usedGPUMemory int = int(temp)
	fmt.Print("Used GPU Memory    = ")
	fmt.Println(usedGPUMemory)
	fmt.Println(" ")

	return usedGPUMemory
}

func Can_Be_Scheduled(freeGPUMemTS map[int]float64, usedGPUMemTS map[int]float64, p *v1.Pod) bool {
	if podGPUMemoryRequestMap[p.Name] <= AVAIL_GPU_MEM {
		fmt.Println("The pod GPU memory request CAN be satisfied.")
		return true
	}

	fmt.Println("The pod GPU memory request CANNOT be satisfied.")

	if Can_Colocate(freeGPUMemTS) {
		harvestedAvailGPUMemory := 32 - Harvest_Resource(usedGPUMemTS)

		fmt.Print("Harvested GPU Available Memory = ")
		fmt.Println(harvestedAvailGPUMemory)

		if podGPUMemoryRequestMap[p.Name] <= harvestedAvailGPUMemory {
			fmt.Println("The pod GPU memory request CAN be satisfied through GPU memory HARVESTING.")
			return true
		}
	}

	fmt.Println("The pod GPU memory request CANNOT be satisfied through GPU memory HARVESTING.")

	/*
	autoCor := AutoCorrelation(freeGPUMemTS, k)
	if autoCor > 0 {
		const steps int = 1

		freeGPUMemPred := AR_1(freeGPUMemTS, autoCor, steps)

		fmt.Print("Free GPU Memory Prediction in ")
		fmt.Print(steps)
		fmt.Print(" sec = ")
                fmt.Println(freeGPUMemPred)

		if podGPUMemoryRequestMap[p.Name] <= freeGPUMemPred {
			fmt.Println("The pod GPU memory request CAN be satisfied through PEAK PREDICTION.")
			return true
		}
	}

	fmt.Println("The pod GPU memory request CANNOT be satisfied through PEAK PREDICTION.")
	*/

	return false
}

func AutoCorrelation(freeGPUMemTS map[int]float64, k int) float64 {
	// Get map data sorted
	keys := make([]int, 0)
	for k, _ := range freeGPUMemTS {
		keys = append(keys, k)
	}

	sort.Ints(keys)

	var N int = d + 1
	var sum float64 = 0.0  // x_i summary

	values := make([]float64, 0)
	for _, timestamp := range keys {
		val := freeGPUMemTS[timestamp]

		values = append(values, val)
		sum += val
	}

	fmt.Println(" ")
	fmt.Println("AutoCorrelation Calculations ...")

	fmt.Print("Values          = ")
	fmt.Println(values)

	var mean float64 = sum / float64(N)

	// Calculate autocorrelation
	// Numerator calculation
	var numerator float64 = 0.0
	for i := 0; i < N - k; i++ {
		numerator += (values[i] - mean) * (values[i + k] - mean)
	}

	// Denominator calculation
        var denominator float64 = 0.0
	for i := 0; i < N; i++ {
		denominator += (values[i] - mean) * (values[i] - mean)
	}

	var autocorrelation float64 = numerator / denominator

	fmt.Print("Autocorrelation of order ")
	fmt.Print(k)
	fmt.Print(" = ")
	fmt.Println(autocorrelation)
	fmt.Println(" ")

	return autocorrelation
}

func AR_1(freeGPUMemTS map[int]float64, autocorrelation float64, steps int) int {
	// Get map data sorted
	keys := make([]int, 0)
	for k, _ := range freeGPUMemTS {
		keys = append(keys, k)
	}

        sort.Ints(keys)

	var N int = d + 1
	var sum float64 = 0.0   // x_i summary
	var sum_2 float64 = 0.0 // x_i * x_i summary

        values := make([]float64, 0)
        for _, timestamp := range keys {
		val := freeGPUMemTS[timestamp]

		values = append(values, val)
		sum += val
		sum_2 += val * val
	}

	fmt.Println(" ")
	fmt.Println("AR(1) Calculations ...")

	fmt.Print("Values                   = ")
	fmt.Println(values)

	var mean float64 = sum / float64(N)

	mean_val_squared := sum_2 / float64(N)
	var variance float64 = mean_val_squared - mean * mean

	// Calculate model parameters
	var phi_1 float64 = autocorrelation
	var phi_0 float64 = mean * (1 - phi_1)
	var error_variance float64 = variance * (1 - phi_1 * phi_1)
	var error_standard_deviation float64 = math.Sqrt(error_variance)

	fmt.Print("φ_0                      = ")
	fmt.Println(phi_0)
	fmt.Print("φ_1                      = ")
	fmt.Println(phi_1)
	fmt.Print("Error Standard Deviation = ")
	fmt.Println(error_standard_deviation)
	fmt.Print("MODEL : ")
	fmt.Print("y_i = ")
	fmt.Print(phi_0)
	fmt.Print(" + ")
	fmt.Print(phi_1)
	fmt.Print(" * ")
	fmt.Println("y_i-1")
	fmt.Println(" ")

	var val float64 = values[N-1]
	for i:=1; i < steps + 1; i++ {
		pred := phi_0 + phi_1 * val

		// fmt.Print("Prediction at step ")
		// fmt.Print(i)
		// fmt.Print(" = ")
		// fmt.Println(pred)

		val = pred
	}

	var temp float64 = math.Round(val * 0.001)
	var freeGPUMemoryPred int = int(temp)

	return freeGPUMemoryPred
}
