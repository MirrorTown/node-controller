package controller

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	CoreListerV1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"node-controller/common"
	"node-controller/conf"
	"node-controller/models"
	"node-controller/util/logs"
	"sort"
	"strconv"
	"time"
)

type K8sWorkerController struct {
	kubeClientset kubernetes.Interface
	podList       CoreListerV1.PodLister
	workerList    CoreListerV1.NodeLister
	workerSynced  cache.InformerSynced
	workqueue     workqueue.RateLimitingInterface
}

type ResourceSummary struct {
	Total int64
	Used  int64
}

type NodeListSummary struct {
	// total nodes count
	Total int64
	// ready nodes count
	Ready int64
	// Schedulable nodes count
	Schedulable int64
}

type NodeListResult struct {
	NodeSummary   NodeListSummary `json:"nodeSummary"`
	CpuSummary    ResourceSummary `json:"cpuSummary"`
	MemorySummary ResourceSummary `json:"memorySummary"`
	Nodes         []Node          `json:"nodes"`
}

type Node struct {
	Name              string            `json:"name,omitempty"`
	Labels            map[string]string `json:"labels,omitempty"`
	CreationTimestamp metaV1.Time       `json:"creationTimestamp"`

	Spec NodeSpec `json:"spec,omitempty"`

	Status NodeStatus `json:"status,omitempty"`
}

type NodeSpec struct {
	Unschedulable bool `json:"unschedulable"`
	// If specified, the node's taints.
	// +optional
	Taints []v1.Taint         `json:"taints,omitempty"`
	Ready  v1.ConditionStatus `json:"ready"`
}

type NodeStatus struct {
	Capacity map[v1.ResourceName]string `json:"capacity,omitempty"`
	NodeInfo v1.NodeSystemInfo          `json:"nodeInfo,omitempty"`
}

type K8sWorkerQueueObj struct {
	Key    string   `json:"key"`
	OldObj *v1.Node `json:"old_obj"`
	Ope    string   `json:"ope"` // add / update / delete
}

func BuildK8sWorkerController(kubeclientset kubernetes.Interface) *K8sWorkerController {
	controller := &K8sWorkerController{
		kubeClientset: kubeclientset,
		podList:       conf.PodInformer.Lister(),
		workerList:    conf.WorkerInformer.Lister(),
		workerSynced:  conf.WorkerInformer.Informer().HasSynced,
		workqueue:     workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "k8sworkerlistener"),
	}

	conf.WorkerInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    controller.addFunc,
			UpdateFunc: controller.updateFunc,
			DeleteFunc: controller.deleteFunc,
		})

	return controller
}

func (c *K8sWorkerController) addFunc(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		runtime.HandleError(err)
		return
	}

	rqo := &K8sWorkerQueueObj{
		Key:    key,
		OldObj: nil,
		Ope:    common.ADD,
	}

	c.workqueue.AddRateLimited(rqo)
}

func (c *K8sWorkerController) updateFunc(oldObj, newObj interface{}) {
	oldWorker := oldObj.(*v1.Node)
	newWorker := newObj.(*v1.Node)
	if oldWorker.ResourceVersion == newWorker.ResourceVersion {
		return
	}

	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(newWorker); err != nil {
		runtime.HandleError(err)
		return
	}
	rqo := &K8sWorkerQueueObj{
		Key:    key,
		OldObj: oldWorker,
		Ope:    common.UPDATE,
	}
	c.workqueue.AddRateLimited(rqo)
}

func (c *K8sWorkerController) deleteFunc(obj interface{}) {
	var key string
	var err error
	if key, err = cache.DeletionHandlingMetaNamespaceKeyFunc(obj); err != nil {
		runtime.HandleError(err)
		return
	}

	rqo := &K8sWorkerQueueObj{
		Key:    key,
		OldObj: nil,
		Ope:    common.DELETE,
	}

	c.workqueue.AddRateLimited(rqo)
}

func (c *K8sWorkerController) Run(stop <-chan struct{}) error {
	//同步缓存
	if ok := cache.WaitForCacheSync(stop); !ok {
		logs.Error("同步缓存失败")
		return fmt.Errorf("failed to wait for caches to sync")
	}

	go wait.Until(c.runworker, time.Second*10, stop)
	return nil
}

func (c *K8sWorkerController) runworker() {
	for c.processNextWorkItem() {
	}
}

func (c *K8sWorkerController) processNextWorkItem() bool {
	defer recoverException()
	obj, shutdown := c.workqueue.Get()
	if shutdown {
		return false
	}
	err := func(obj interface{}) error {
		defer c.workqueue.Done(obj)
		var key string
		var ok bool
		var rqo *K8sWorkerQueueObj
		if rqo, ok = obj.(*K8sWorkerQueueObj); !ok {
			c.workqueue.Forget(obj)
			return fmt.Errorf("expected K8sWorkerQueueObj in workqueue but got %#v", obj)
		}
		if err := c.syncHandler(rqo); err != nil {
			return fmt.Errorf("error syncing '%s': %s", key, err.Error())
		}
		c.workqueue.Forget(obj)
		return nil
	}(obj)

	if err != nil {
		runtime.HandleError(err)
	}
	return true
}

func (c *K8sWorkerController) syncHandler(rqo *K8sWorkerQueueObj) error {
	key := rqo.Key
	switch rqo.Ope {
	case common.ADD:
		return c.add(key)
	case common.UPDATE:
		return c.sync(rqo)
	case common.DELETE:
		return c.sync(rqo)
	default:
		// log error
		return fmt.Errorf("K8sWorkerQueueObj is not expected")
	}
}

func (c *K8sWorkerController) add(key string) error {
	_, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		logs.Error("invalid resource key: %s", key)
		return fmt.Errorf("invalid resource key: %s", key)
	}

	worker, err := c.workerList.Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			logs.Info("vm %s is removed", key)
			return nil
		}
		runtime.HandleError(fmt.Errorf("failed to list vm  %s/%s", key, err.Error()))
		return err
	}
	fmt.Println(worker.Name, worker.Status.Phase)

	return err
}

// sync
// 1.diff routes between old and new objects
// 2.delete routes not exist
func (c *K8sWorkerController) sync(rqo *K8sWorkerQueueObj) error {
	key := rqo.Key
	_, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		logs.Error("invalid resource key: %s", key)
		return fmt.Errorf("invalid resource key: %s", key)
	}

	switch {
	case rqo.Ope == common.UPDATE:
		worker, err := c.workerList.Get(name)
		if err != nil {
			if errors.IsNotFound(err) {
				logs.Info("vm %s is removed", key)
				return nil
			}
			runtime.HandleError(fmt.Errorf("failed to list vm %s/%s", key, err.Error()))
			return err
		}
		fmt.Println(worker.Name, worker.Status.Conditions[0].Type, worker.Status.Conditions[0].Status, worker.Status.Conditions[1].Type, worker.Status.Conditions[1].Status)
		return nil
	case rqo.Ope == common.DELETE:
		//TODO worker节点宕机的告警通知
		record := &models.Record{
			User:        "",
			HostName:    name,
			Description: "worker节点从k8s集群失联",
			Status:      0,
			CreateTime:  nil,
			UpdateTime:  nil,
		}
		err := models.RecordMode.Add(record)
		if err != nil {
			logs.Error("记录告警记录失败, ", err)
		}
		fmt.Println("delete")
		return nil
	default:
		return fmt.Errorf("not expected in (WorkerController) sync")
	}
}

func (c *K8sWorkerController) ListNode() (*NodeListResult, error) {
	nodeList, err := c.workerList.List(labels.Everything())
	if err != nil {
		return nil, err
	}

	nodes := make([]Node, 0)
	ready := 0
	schedulable := 0

	// unit m  1 core = 1000m
	var avaliableCpu int64 = 0
	// unit Byte
	var avaliableMemory int64 = 0

	avaliableNodeMap := make(map[string]*v1.Node)

	for _, node := range nodeList {
		isReady := false
		isSchedulable := false
		for _, condition := range node.Status.Conditions {
			if condition.Type == v1.NodeReady && condition.Status == v1.ConditionTrue {
				ready += 1
				isReady = true
			}

		}
		if !node.Spec.Unschedulable {
			schedulable += 1
			isSchedulable = true
		}

		if isReady && isSchedulable {
			avaliableNodeMap[node.Name] = node

			cpuQuantity := node.Status.Allocatable[v1.ResourceCPU]
			memoryQuantity := node.Status.Allocatable[v1.ResourceMemory]
			// unit m
			avaliableCpu += cpuQuantity.MilliValue()
			// unit Byte
			avaliableMemory += memoryQuantity.Value()
		}

		nodes = append(nodes, toNode(node))
	}

	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].Name < nodes[j].Name
	})

	resourceList, err := c.podUsedResourcesOnAvaliableNode(avaliableNodeMap)
	if err != nil {
		return nil, err
	}

	return &NodeListResult{
		NodeSummary: NodeListSummary{
			Total:       int64(len(nodes)),
			Ready:       int64(ready),
			Schedulable: int64(schedulable),
		},
		CpuSummary: ResourceSummary{
			Total: avaliableCpu / 1000,
			Used:  resourceList.Cpu / 1000,
		},
		MemorySummary: ResourceSummary{
			Total: avaliableMemory / (1024 * 1024 * 1024),
			Used:  resourceList.Memory / (1024 * 1024 * 1024),
		},
		Nodes: nodes,
	}, nil
}

func (c *K8sWorkerController) podUsedResourcesOnAvaliableNode(avaliableNodeMap map[string]*v1.Node) (*common.ResourceList, error) {
	result := &common.ResourceList{}
	cachePods, err := c.podList.List(labels.Everything())
	if err != nil {
		return nil, err
	}

	for _, pod := range cachePods {
		// Exclude Pod on Unavailable Node
		_, ok := avaliableNodeMap[pod.Spec.NodeName]
		if pod.Status.Phase == v1.PodFailed || pod.Status.Phase == v1.PodSucceeded || pod.DeletionTimestamp != nil || !ok {
			continue
		}

		resourceList := common.ContainersRequestResourceList(pod.Spec.Containers)

		result.Cpu += resourceList.Cpu
		result.Memory += resourceList.Memory
	}

	return result, nil
}

func toNode(knode *v1.Node) Node {

	node := Node{
		Name:              knode.Name,
		Labels:            knode.Labels,
		CreationTimestamp: knode.CreationTimestamp,
		Spec: NodeSpec{
			Unschedulable: knode.Spec.Unschedulable,
			Taints:        knode.Spec.Taints,
		},
		Status: NodeStatus{
			NodeInfo: knode.Status.NodeInfo,
		},
	}

	capacity := make(map[v1.ResourceName]string)

	for resourceName, value := range knode.Status.Capacity {
		if resourceName == v1.ResourceCPU {
			// cpu unit core
			capacity[resourceName] = strconv.Itoa(int(value.Value()))
		}
		if resourceName == v1.ResourceMemory {
			// memory unit Gi
			capacity[resourceName] = strconv.Itoa(int(value.Value() / (1024 * 1024 * 1024)))
		}
	}
	node.Status.Capacity = capacity

	for _, condition := range knode.Status.Conditions {
		if condition.Type == v1.NodeReady {
			node.Spec.Ready = condition.Status
		}
	}

	return node
}
