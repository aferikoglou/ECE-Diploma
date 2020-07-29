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

const PROMETHEUS_ENDPOINT string = "http://192.168.1.146:30000"

const schedDelay int = 10 // sec

// Resource Harvesting Constants
const covThr float64 = 0.5
const resHarvPerc float64 = 0.8 // 80%-ile

// Peak Prediction Constants
const k int = 10 // 1
const steps int = 1 // 10
const autocorThr float64 = 0.3


var AVAIL_GPU_MEM int = 32510 // in MiB

var podGPUMemoryRequestMap map[string]int

var runningPodsList *list.List
var tmpList *list.List

var lastSchedulingTimestamp int

// LR Model Evaluation
var predTimestampList *list.List
var autocorValMap map[int32]float64
var predValMap map[int32]int

var FOUND_SCHEDULABLE_POD bool = true


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
			fmt.Print(time.Now().Format("01-02-2006 15:04:05.000000") + "\t")
			fmt.Printf("New Node Added to Store : %s\n", node.GetName())
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

				fmt.Print(time.Now().Format("01-02-2006 15:04:05.000000") + "\t")
				fmt.Printf("Found a pod to schedule : [%s/%s] / Assigned Priority : %f\n", pod.Namespace, pod.Name, priority)

				podPriorityQueue.Insert(pod, priority)
			}
		},
	})

	factory.Start(quit)

	return nodeInformer.Lister()
}

func main() {
	fmt.Print(time.Now().Format("01-02-2006 15:04:05.000000") + "\t")
	fmt.Printf("Improved custom Kube-Knots like GPU Scheduler\n")

	podGPUMemoryRequestMap = make(map[string]int)

	podPriorityQueue := pq.New()

	runningPodsList = list.New()
	tmpList = list.New()

	lastSchedulingTimestamp = int(time.Now().Unix())

	predTimestampList = list.New()
	autocorValMap = make(map[int32]float64)
	predValMap = make(map[int32]int)

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

	// If the pod priority queue is empty wait
	for pod == nil {
		pod, _ = s.podPriorityQueue.Pop()

		s.removeRunListPods()

		getState()

		time.Sleep(500 * time.Millisecond)
	}

	p := pod.(*v1.Pod)

	FOUND_SCHEDULABLE_POD = true
	tmpList = list.New()

	s.removeRunListPods()

	// Get a pod from the priority queue and check whether it can be scheduled.
	// If it cannot be scheduled save the pod in tmpList and try the next one.
	for !Can_Be_Scheduled(p) {
		tmpList.PushBack(p)
		np, _ := s.podPriorityQueue.Pop()

		if np == nil {
			FOUND_SCHEDULABLE_POD = false
			break
		}
		p = np.(*v1.Pod)
	}

	getState()

	for e := tmpList.Front(); e != nil; e = e.Next() {
		pod_ := e.Value.(*v1.Pod)
		s.podPriorityQueue.Insert(pod_, float64(podGPUMemoryRequestMap[pod_.Name]))
	}

	time.Sleep(500 * time.Millisecond)

	// LR Model Evaluation
	getRealGPUFreeMemory()

	if !FOUND_SCHEDULABLE_POD {
		return
	} else {
		// Try to schedule the pod
		node, err := s.findFit(p)
		if err != nil {
			log.Println("Cannot find node that fits pod.", err.Error())

			s.podPriorityQueue.Insert(p, float64(podGPUMemoryRequestMap[p.Name]))

			return
		}

		err = s.bindPod(p, node)
		if err != nil {
			log.Println("Failed to bind pod.", err.Error())

			s.podPriorityQueue.Insert(p, float64(podGPUMemoryRequestMap[p.Name]))

			return
		}

		fmt.Print(time.Now().Format("01-02-2006 15:04:05.000000") + "\t")
		message := fmt.Sprintf("Placed pod [%s/%s] on %s node", p.Namespace, p.Name, node)

		err = s.emitEvent(p, message)
		if err != nil {
			log.Println("Failed to emit scheduled event.", err.Error())

			s.podPriorityQueue.Insert(p, float64(podGPUMemoryRequestMap[p.Name]))

			return
		}

		fmt.Println(message)

		// If the pod is scheduled successfully add it to the running pods list and
		// decrease the available GPU memory.
		runningPodsList.PushBack(p)
		AVAIL_GPU_MEM -= podGPUMemoryRequestMap[p.Name]

		// Save the scheduling timestamp for d calculation.
		lastSchedulingTimestamp = int(time.Now().Unix())

		time.Sleep(time.Duration(schedDelay) * time.Second)
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

	fmt.Print(time.Now().Format("01-02-2006 15:04:05.000000") + "\t")
	fmt.Print("Nodes that fit : [ ")
	for _, n := range filteredNodes {
		fmt.Print(n.Name + " ")
	}
	fmt.Println("]")
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
	fmt.Print(time.Now().Format("01-02-2006 15:04:05.000000") + "\t")
	fmt.Printf("Available GPU Memory : %d MiB\n", AVAIL_GPU_MEM)

	fmt.Print(time.Now().Format("01-02-2006 15:04:05.000000") + "\t")
        fmt.Printf("Running pods         : [ ")
        for e := runningPodsList.Front(); e != nil; e = e.Next() {
		pod := e.Value.(*v1.Pod)
		fmt.Printf(pod.Name + " ")
	}
	fmt.Printf("]\n")

	fmt.Print(time.Now().Format("01-02-2006 15:04:05.000000") + "\t")
	fmt.Printf("Temp List pods       : [ ")
	for e := tmpList.Front(); e != nil; e = e.Next() {
		pod := e.Value.(*v1.Pod)
		fmt.Printf(pod.Name + " ")
	}
	fmt.Printf("]\n")
}

func (s *Scheduler) removeRunListPods() {
	for e := runningPodsList.Front(); e != nil; e = e.Next() {
		pod_ := e.Value.(*v1.Pod)

		resp, er := s.clientset.CoreV1().Pods("default").Get(pod_.Name, metav1.GetOptions{})

		f_ := er != nil || resp.Status.Phase == "Succeeded"
		if f_ {
			gpuMemoryReq, exists := podGPUMemoryRequestMap[pod_.Name]
			if exists {
				AVAIL_GPU_MEM += gpuMemoryReq
				delete(podGPUMemoryRequestMap, pod_.Name)
			}

			runningPodsList.Remove(e)
		}
	}
}

func getRealGPUFreeMemory() {
	currentTimestamp := int32(time.Now().Unix())

        for e := predTimestampList.Front(); e != nil; e = e.Next() {
		predictionTimestamp_ := e.Value.(int32)

                if currentTimestamp - predictionTimestamp_ >= 10 {
			realGPUMemoryMap := promqueriesutil.Query(PROMETHEUS_ENDPOINT, "dcgm_fb_free")

			realGPUMemory := 0
			for _, val := range realGPUMemoryMap {
				realGPUMemory = int(val)
			}

			autocorrelationVal := autocorValMap[predictionTimestamp_]
			predictedGPUMemory := predValMap[predictionTimestamp_]

			fmt.Printf("Timestamp = %d / Predicted GPU free memory = %d MiB / Steps = %d sec / Autocorrelation order = %d / Autocorrelation value = %f / Autocorrelation threshold = %f \n", predictionTimestamp_, predictedGPUMemory, steps, k, autocorrelationVal, autocorThr)

			fmt.Printf("Timestamp = %d / Real GPU free memory = %d MiB \n", currentTimestamp, realGPUMemory)

			// Delete
			delete(autocorValMap, predictionTimestamp_)
			delete(predValMap, predictionTimestamp_)
			predTimestampList.Remove(e)
		}
        }
}

func startTimestampCalculation(endTimestamp int32) int32{
	const MIN_D int = 20

        val := int(endTimestamp - int32(lastSchedulingTimestamp))
        if val < MIN_D {
		return (endTimestamp - int32(20))
	} else {
                return int32(lastSchedulingTimestamp)
        }
}

func getGPUMetricTS(PROMETHEUS_ENDPOINT string, GPU_METRIC string, startTimestamp int32, endTimestamp int32) map[int]float64 {
	const STEP int = 1

        return promqueriesutil.RangeQuery(PROMETHEUS_ENDPOINT, GPU_METRIC, int64(startTimestamp), int64(endTimestamp), STEP)
}

func Can_Be_Scheduled(p *v1.Pod) bool {
	// For LR Model evaluation
        getRealGPUFreeMemory()

	fmt.Print(time.Now().Format("01-02-2006 15:04:05.000000") + "\t")
	fmt.Print("Check whether pod ")
	fmt.Print(p.Name)
	fmt.Println(" can be scheduled.")

	if podGPUMemoryRequestMap[p.Name] <= AVAIL_GPU_MEM {
		fmt.Print(time.Now().Format("01-02-2006 15:04:05.000000") + "\t")
		fmt.Println("The pod GPU memory request CAN be satisfied.")
		return true
	}

	fmt.Print(time.Now().Format("01-02-2006 15:04:05.000000") + "\t")
	fmt.Println("The pod GPU memory request CANNOT be satisfied.")

        // Get dcgm_fb_free and dcgm_fb_used GPU metrics timeseries
	endTimestamp := int32(time.Now().Unix())
	startTimestamp := startTimestampCalculation(endTimestamp)
	freeGPUMemTS := getGPUMetricTS(PROMETHEUS_ENDPOINT, "dcgm_fb_free", startTimestamp, endTimestamp)
	usedGPUMemTS := getGPUMetricTS(PROMETHEUS_ENDPOINT, "dcgm_fb_used", startTimestamp, endTimestamp)

	if len(freeGPUMemTS) < 20 || len(usedGPUMemTS) < 20 {
		fmt.Print(time.Now().Format("01-02-2006 15:04:05.000000") + "\t")
		fmt.Println("Not enough values for RESOURCE HARVESTING and PEAK PREDICTION calculations.")
		return false
	}

	// For LR Model evaluation
	getRealGPUFreeMemory()

	if Can_Colocate(freeGPUMemTS) {
		harvestedAvailGPUMemory := 32510 - Harvest_Resource(usedGPUMemTS)

		fmt.Print(time.Now().Format("01-02-2006 15:04:05.000000") + "\t")
		fmt.Printf("Harvested GPU Available Memory = %d MiB\n", harvestedAvailGPUMemory)

		if podGPUMemoryRequestMap[p.Name] <= harvestedAvailGPUMemory {
			fmt.Print(time.Now().Format("01-02-2006 15:04:05.000000") + "\t")
			fmt.Println("The pod GPU memory request CAN be satisfied through GPU memory HARVESTING.")
			return true
		}
	}

	fmt.Print(time.Now().Format("01-02-2006 15:04:05.000000") + "\t")
	fmt.Println("The pod GPU memory request CANNOT be satisfied through GPU memory HARVESTING.")

	// For LR Model evaluation
	getRealGPUFreeMemory()

	autocorrelationVal := AutoCorrelation(freeGPUMemTS, k)
	if autocorrelationVal > autocorThr {
		predictionTimestamp := int32(time.Now().Unix())

		freeGPUMemPred := AR_1(freeGPUMemTS, k, steps)

		// Save prediction timestamp, autocorrelation value and predicted GPU memory usage value
		PRED_TIMESTAMP_IS_FOUND := false
		for e := predTimestampList.Front(); e != nil; e = e.Next() {
			predictionTimestamp_ := e.Value.(int32)

			if predictionTimestamp_ == predictionTimestamp {
				PRED_TIMESTAMP_IS_FOUND = true
				break
			}
		}

		if !PRED_TIMESTAMP_IS_FOUND {
			predTimestampList.PushBack(predictionTimestamp)
			autocorValMap[predictionTimestamp] = autocorrelationVal
			predValMap[predictionTimestamp] = freeGPUMemPred
		}

		fmt.Print(time.Now().Format("01-02-2006 15:04:05.000000") + "\t")
		fmt.Printf("Free GPU Memory Prediction in %d sec = %d MiB\n", (steps * k), freeGPUMemPred)

		if podGPUMemoryRequestMap[p.Name] <= freeGPUMemPred {
			fmt.Print(time.Now().Format("01-02-2006 15:04:05.000000") + "\t")
			fmt.Println("The pod GPU memory request CAN be satisfied through PEAK PREDICTION.")
			return true
		}
	}

	fmt.Print(time.Now().Format("01-02-2006 15:04:05.000000") + "\t")
	fmt.Println("The pod GPU memory request CANNOT be satisfied through PEAK PREDICTION.")

	// For LR Model evaluation
        getRealGPUFreeMemory()

	return false
}

func Can_Colocate(freeGPUMemTS map[int]float64) bool {
	keys := make([]int, 0)
	for k, _ := range freeGPUMemTS {
		keys = append(keys, k)
	}

        sort.Ints(keys)

        N := len(keys)

        var sum float64 = 0.0   // x_i sum
        var sum_2 float64 = 0.0 // x_i * x_i sum

        for _, timestamp := range keys {
		val := freeGPUMemTS[timestamp]

                sum += val
                sum_2 += val * val
        }

        fmt.Print(time.Now().Format("01-02-2006 15:04:05.000000") + "\t")
        fmt.Printf("Can_Colocate Calculations ...\n")

        var mean float64 = sum / float64(N)
        fmt.Print(time.Now().Format("01-02-2006 15:04:05.000000") + "\t")
        fmt.Printf("Mean               = %f\n", mean)

        mean_val_squared := sum_2 / float64(N)
        var variance float64 = mean_val_squared - mean * mean

        var standard_deviation float64 = math.Sqrt(variance)
        fmt.Print(time.Now().Format("01-02-2006 15:04:05.000000") + "\t")
        fmt.Printf("Standard Deviation = %f\n", standard_deviation)

        var COV float64 = standard_deviation / mean
        fmt.Print(time.Now().Format("01-02-2006 15:04:05.000000") + "\t")
        fmt.Printf("COV                = %f\n", COV)

        return (COV < covThr)
}

func Harvest_Resource(usedGPUMemTS map[int]float64) int {
	values := make([]float64, 0)
	for _, val := range usedGPUMemTS {
		values = append(values, val)
	}

        fmt.Print(time.Now().Format("01-02-2006 15:04:05.000000") + "\t")
        fmt.Printf("Harvest_Resource Calculations ...\n")

        sort.Float64s(values)
        fmt.Print(time.Now().Format("01-02-2006 15:04:05.000000") + "\t")
        fmt.Printf("Sorted Values   : ")
        fmt.Println(values)

        var index int = int(float64(len(values)) * resHarvPerc)
        fmt.Print(time.Now().Format("01-02-2006 15:04:05.000000") + "\t")
        fmt.Printf("Index           = %d\n", index)

        var percentile float64 = values[index]
        fmt.Print(time.Now().Format("01-02-2006 15:04:05.000000") + "\t")
        fmt.Printf("80perc-ile      = %f\n", percentile)

        var usedGPUMemory int = int(percentile)
        fmt.Print(time.Now().Format("01-02-2006 15:04:05.000000") + "\t")
        fmt.Printf("Used GPU Memory = %d MiB\n", usedGPUMemory)

        return usedGPUMemory
}

func AutoCorrelation(freeGPUMemTS map[int]float64, k int) float64 {
	keys := make([]int, 0)
	for k, _ := range freeGPUMemTS {
		keys = append(keys, k)
	}

	sort.Ints(keys)

	var sum float64 = 0.0  // x_i sum

	values := make([]float64, 0)
	for _, timestamp := range keys {
		val := freeGPUMemTS[timestamp]

		values = append(values, val)
		sum += val
	}

	fmt.Print(time.Now().Format("01-02-2006 15:04:05.000000") + "\t")
	fmt.Println("AutoCorrelation Calculations ...")

	fmt.Print(time.Now().Format("01-02-2006 15:04:05.000000") + "\t")
	fmt.Printf("Values                      : ")
	fmt.Println(values)

	N := len(values)

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

	fmt.Print(time.Now().Format("01-02-2006 15:04:05.000000") + "\t")
	fmt.Printf("Autocorrelation of order %d = %f\n", k, autocorrelation)

	return autocorrelation
}

func AR_1(freeGPUMemTS map[int]float64, k int, steps int) int {
	// Get map data sorted
	keys := make([]int, 0)
	for k, _ := range freeGPUMemTS {
		keys = append(keys, k)
	}

        sort.Ints(keys)

        values := make([]float64, 0)
        for _, timestamp := range keys {
		val := freeGPUMemTS[timestamp]

		values = append(values, val)
	}

	N := len(values)

	fmt.Print(time.Now().Format("01-02-2006 15:04:05.000000") + "\t")
	fmt.Println("AR(1) Calculations ...")

	fmt.Print(time.Now().Format("01-02-2006 15:04:05.000000") + "\t")
	fmt.Print("values : ")
	fmt.Println(values)
	fmt.Print(time.Now().Format("01-02-2006 15:04:05.000000") + "\t")
	fmt.Print("x      : ")
	fmt.Println(values[0:(N - k - 1)])
	fmt.Print(time.Now().Format("01-02-2006 15:04:05.000000") + "\t")
	fmt.Print("y      : ")
	fmt.Println(values[k:(N - 1)])

	// Linear Regression - Least Squares
	var sum_x float64 = 0.0
	var sum_y float64 = 0.0
	var sum_xy float64 = 0.0
	var sum_x_2 float64 = 0.0

	for i := 0; i < (N - k); i++ {
		sum_x += values[i]
		sum_y += values[i + k]
		sum_xy += values[i] * values[i + k]
		sum_x_2 += values[i] * values[i]
	}

	var mean_x float64 = sum_x / float64(N - k)
	var mean_y float64 = sum_y / float64(N - k)
	var numerator float64 = sum_xy - float64(N - k) * mean_x * mean_y
	var denominator float64 = sum_x_2 - float64(N - k) * mean_x * mean_x

	// Calculate model parameters
	var phi_1 float64 = numerator / denominator
	var phi_0 float64 = mean_y - phi_1 * mean_x

	fmt.Print(time.Now().Format("01-02-2006 15:04:05.000000") + "\t")
	fmt.Print("MODEL : ")
	fmt.Printf("y[i] = %f + %f * y[i-1]\n", phi_0, phi_1)

	var val float64 = values[N-1]
	for i:=1; i < steps + 1; i++ {
		pred := phi_0 + phi_1 * val

		fmt.Print(time.Now().Format("01-02-2006 15:04:05.000000") + "\t")
		fmt.Printf("Prediction at step %d = %f\n", i, pred)

		val = pred
	}

	var freeGPUMemoryPred int = int(val)

	return freeGPUMemoryPred
}
