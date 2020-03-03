package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"
	"errors"
	"strings"
	"github.com/jupp0r/go-priority-queue"
//	"github.com/promqueriesutil"

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

// Scheduler name
const schedulerName = "custom-scheduler"

// Defines the predicateFunc type which is a function that gets as
// input a node and a pod and returns a bool value.
type predicateFunc func(node *v1.Node, pod *v1.Pod) bool

// Defines the priorityFunc type which is a function that gets as
// input a node and a pod and returns an int value.
type priorityFunc func(node *v1.Node, pod *v1.Pod) int

// A Scheduler is a struct with the following fields:
// clientset 
// podPriorityQueue - A priority queue that contains pointers to the pods that are going to be scheduled
// nodeLister - The nodes that are in the cache of the node informer
// predicates - A predicateFunc slice 
// priorities - A priorityFunc slice
type Scheduler struct {
	clientset  *kubernetes.Clientset
	podPriorityQueue pq.PriorityQueue
	nodeLister listersv1.NodeLister
	predicates []predicateFunc
	priorities []priorityFunc
}

// NewScheduler - Creates a new scheduler
func NewScheduler(podPriorityQueue pq.PriorityQueue, quit chan struct{}) Scheduler {
	// Returns a config object which uses the service account Kubernetes gives to pods.
	// It's intended for clients that expect to be running inside a pod running on Kubernetes.
	// It will return ErrNotInCluster if called from a process not running in a Kubernetes 
	// environment.
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatal(err)
	}
	// Creates a new Clientset for the given config. If config's RateLimiter is not set
	// and QPS and Burst are acceptable, NewForConfig will generate a rate-limiter in 
	// configShallowCopy.
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}
	// Returns the new Scheduler.
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

// initInformers - Creates node and pod informer
func initInformers(clientset *kubernetes.Clientset, podPriorityQueue pq.PriorityQueue, quit chan struct{}) listersv1.NodeLister {
	// Provides shared informers for resources in all known API group versions.
	factory := informers.NewSharedInformerFactory(clientset, 0)

	// Creates a node informer. The node informer keeps a cache of all the nodes
	// in the cluster and updates the state of the cache if a new node is added,
	// or an exisitng node is deleted. (We don't have to query the apiserver for
	// the list of nodes every time we're scheduling a pod)
	nodeInformer := factory.Core().V1().Nodes()
	// Define what should be done when a new node is added.
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

	/***** Create a map that contains the duration of the measured pods. *****/
        podDurationMap := make(map[string]int)

        podDurationMap["matmul32768"] = 189
        podDurationMap["onnxmobilenetss"] = 71
        podDurationMap["tfssdmobilenetss"] = 89
	podDurationMap["lstmep1"] = 36

	/***** *****/

	// Creates a pod informer. The pod informer keeps a cache of all the pods
	// in the cluster and updates the state of the cache if a new pod is added,
	// or an exisitng pod is deleted.
	podInformer := factory.Core().V1().Pods()
	// Define what should be done when a new pod is added.
	podInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			pod, ok := obj.(*v1.Pod)
			if !ok {
				log.Println("This is not a pod.")
				return
			}
			// Add the pod to the podPriorityQueue if and only if it is not assigned to a node
			// and the SchedulerName is schedulerName.
			if pod.Spec.NodeName == "" && pod.Spec.SchedulerName == schedulerName {
				priority := 1 / podDurationPredictor(pod.Name, podDurationMap)

				fmt.Print("Pod Name=" + pod.Name)
				fmt.Println(" Assigned Priority=" + fmt.Sprintf("%f", priority))
				fmt.Println(" ")

				podPriorityQueue.Insert(pod, priority)
			}
		},
	})
	// Initializes all requested informers.
	factory.Start(quit)
	// Returns a list with all the nodes in the cache of the nodeInformer.
	return nodeInformer.Lister()
}

func main() {
	fmt.Println("Custom Kubernetes GPU scheduler")
	fmt.Println()
	// Seed for random number generator used for the random priorities
	rand.Seed(time.Now().Unix())
	// Create a priority queue for the pods.
	podPriorityQueue := pq.New()
	// quit is a struct channel
	quit := make(chan struct{})
	// We ensure that when main() is finished the quit channel will be closed
	defer close(quit)
	// Creates a new scheduler by using NewScheduler function
	scheduler := NewScheduler(podPriorityQueue, quit)
	// Starts running the scheduler
	scheduler.Run(quit)
}

func (s *Scheduler) Run(quit chan struct{}) {
	// Loops until quit channel is closed and executes ScheduleOne function every period time.
	wait.Until(s.ScheduleOne, 0, quit)
}

func (s *Scheduler) ScheduleOne() {
	// Loops until the pod priority queue is not empty.
	pod, _ := s.podPriorityQueue.Pop()

	for pod == nil {
		pod, _ = s.podPriorityQueue.Pop()
		time.Sleep(1 * time.Second)
	}

	p := pod.(*v1.Pod)

	fmt.Println("Found a pod to schedule:", p.Namespace, "/", p.Name)

	// Finds the fitting node.
	node, err := s.findFit(p)
	if err != nil {
		log.Println("Cannot find node that fits pod.", err.Error())
		return
	}

	// Binds the pod to the node.
	err = s.bindPod(p, node)
	if err != nil {
		log.Println("Failed to bind pod.", err.Error())
		return
	}

	// Emits event.
	message := fmt.Sprintf("Placed pod [%s/%s] on %s\n", p.Namespace, p.Name, node)

	err = s.emitEvent(p, message)
	if err != nil {
		log.Println("Failed to emit scheduled event.", err.Error())
		return
	}

	fmt.Println(message)

	// Wait until the scheduled pod is finished. If we don't wait the pod will be scheduled
	// to kube-gpu node but the gpu resource will be occupied causing an UnexpectedAdmissionError
	// for the pod.
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatal(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	for {
		pod, _ := clientset.CoreV1().Pods("default").Get(p.Name, metav1.GetOptions{})
		f := pod.Status.Phase == "Succeeded"
		if f {
			break
		}
	}

	// Update pod priorities.
	pqMap := s.podPriorityQueue.GetLookup()

	for key, _ := range pqMap {
		// Update code base.
		// fmt.Print("Pod Name="key.(*v1.Pod).Name)
		// fmt.Println(" New Assigned Priority=" + fmt.Sprintf("%f", priority))
		// s.podPriorityQueue.UpdatePriority(key.(*v1.Pod), priority)

		fmt.Println(key.(*v1.Pod).Name)
	}

}

func (s *Scheduler) findFit(pod *v1.Pod) (string, error) {
	// Returns ALL the available nodes. (labels.Everything() - Returns a selector that matches all labels)
	nodes, err := s.nodeLister.List(labels.Everything())
	if err != nil {
		return "", err
	}
	// Returns the nodes that satisfy all the Scheduler predicates.
	filteredNodes := s.runPredicates(nodes, pod)
	if len(filteredNodes) == 0 {
		return "", errors.New("Failed to find node that fits pod.")
	}
	// Creates a map which stores the names of the filtered nodes
	// and their corresponding priorities.
	priorities := s.prioritize(filteredNodes, pod)
	// Returns the name of the node with the highest priority.
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
	// Creates a node slice which contains all the nodes that satisfy all 
	// of the Scheduler predicates.
	filteredNodes := make([]*v1.Node, 0)
	for _, node := range nodes {
		if s.predicatesApply(node, pod) {
			filteredNodes = append(filteredNodes, node)
		}
	}
	// Prints all the filtered nodes.
	log.Println("Nodes that fit:")
	for _, n := range filteredNodes {
		log.Println(n.Name)
	}
	return filteredNodes
}

// Applies all the Scheduler predicates.
func (s *Scheduler) predicatesApply(node *v1.Node, pod *v1.Pod) bool {
	for _, predicate := range s.predicates {
		if !predicate(node, pod) {
			return false
		}
	}
	return true
}

// containsGPUPredicate - Returns true only if the node is kube-gpu
func containsGPUPredicate(node *v1.Node, pod *v1.Pod) bool {
	return (node.Name == "kube-gpu-1")
}

func (s *Scheduler) prioritize(nodes []*v1.Node, pod *v1.Pod) map[string]int {
	// Creates a map which stores the priorities of nodes.
	priorities := make(map[string]int)
	// For each node apply all the Scheduler priority functions and
	// store them to priorities map.
	for _, node := range nodes {
		for _, priority := range s.priorities {
			priorities[node.Name] += priority(node, pod)
		}
	}
	log.Println("Calculated priorities:", priorities)
	return priorities
}

// Returns the name of the node with the highest priority.
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

// randomPriority - Returns an integer in [0, 100)
func randomPriority(node *v1.Node, pod *v1.Pod) int {
	return rand.Intn(100)
}

// podDurationPredictor - Returns the estimated duration 
func podDurationPredictor(podName string, podDurationMap map[string]int) float64 {
	source := rand.NewSource(time.Now().UnixNano())
        randGen := rand.New(source)

	var podDuration float64 = 1.0
	for name, duration := range podDurationMap {
		if strings.Contains(podName, name) {
			podDuration = float64(duration)
			break
		}
	}

	var estimatedDuration float64 = podDuration + (2 * randGen.Float64() - 1) * 0.1 * podDuration

	fmt.Println(podName + " Estimated Duration = " + fmt.Sprintf("%f", estimatedDuration) + " sec")

	return estimatedDuration
}

