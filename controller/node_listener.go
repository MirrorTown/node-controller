package controller

import (
	"fmt"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	apiv1alpha1 "node-controller/api/virtulmachinecontroller/v1alpha1"
	"node-controller/common"
	clientSet "node-controller/generated/clientset/versioned"
	clientScheme "node-controller/generated/clientset/versioned/scheme"
	nodeInformer "node-controller/generated/informers/externalversions/virtulmachinecontroller/v1alpha1"
	"node-controller/generated/listers/virtulmachinecontroller/v1alpha1"
	"node-controller/util/logs"
	"time"
)

// recover any exception
func recoverException() {
	if err := recover(); err != nil {
		logs.Error(err)
	}
}

type VirtulMachineListenerController struct {
	nodeClientset clientSet.Interface
	nodeList      v1alpha1.VirtulMachineLister
	nodeSynced    cache.InformerSynced
	workqueue     workqueue.RateLimitingInterface
}

type VirtulMachineQueueObj struct {
	Key    string                     `json:"key"`
	OldObj *apiv1alpha1.VirtulMachine `json:"old_obj"`
	Ope    string                     `json:"ope"` // add / update / delete
}

func BuildVirtulMachineListenerController(
	api6RouteClientset clientSet.Interface,
	api6RouteInformer nodeInformer.VirtulMachineInformer) *VirtulMachineListenerController {

	runtime.Must(clientScheme.AddToScheme(scheme.Scheme))
	controller := &VirtulMachineListenerController{
		nodeClientset: api6RouteClientset,
		nodeList:      api6RouteInformer.Lister(),
		nodeSynced:    api6RouteInformer.Informer().HasSynced,
		workqueue:     workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "VirtulMachineListener"),
	}
	api6RouteInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    controller.addFunc,
			UpdateFunc: controller.updateFunc,
			DeleteFunc: controller.deleteFunc,
		})
	return controller
}

func (n *VirtulMachineListenerController) addFunc(obj interface{}) {
	fmt.Println("add:", obj.(*apiv1alpha1.VirtulMachine).Name)
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		runtime.HandleError(err)
		return
	}
	rqo := &VirtulMachineQueueObj{Key: key, OldObj: nil, Ope: common.ADD}
	n.workqueue.AddRateLimited(rqo)
}

func (n *VirtulMachineListenerController) updateFunc(oldObj, newObj interface{}) {
	oldVirtulMachine := oldObj.(*apiv1alpha1.VirtulMachine)
	newVirtulMachine := newObj.(*apiv1alpha1.VirtulMachine)
	if oldVirtulMachine.ResourceVersion == newVirtulMachine.ResourceVersion {
		return
	}
	//n.addFunc(newObj)
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(newObj); err != nil {
		runtime.HandleError(err)
		return
	}
	rqo := &VirtulMachineQueueObj{Key: key, OldObj: oldVirtulMachine, Ope: common.UPDATE}
	n.workqueue.AddRateLimited(rqo)
}

func (n *VirtulMachineListenerController) deleteFunc(obj interface{}) {
	var key string
	var err error
	key, err = cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		runtime.HandleError(err)
		return
	}
	rqo := &VirtulMachineQueueObj{Key: key, OldObj: nil, Ope: common.DELETE}
	n.workqueue.AddRateLimited(rqo)
}

func (n *VirtulMachineListenerController) Run(stop <-chan struct{}) error {
	//defer c.workqueue.ShutDown()
	// 同步缓存
	if ok := cache.WaitForCacheSync(stop); !ok {
		logs.Error("同步缓存失败")
		return fmt.Errorf("failed to wait for caches to sync")
	}
	go wait.Until(n.runWorker, time.Second*10, stop)
	return nil
}

func (n *VirtulMachineListenerController) runWorker() {
	for n.processNextWorkItem() {
	}
}

func (n *VirtulMachineListenerController) processNextWorkItem() bool {
	defer recoverException()
	obj, shutdown := n.workqueue.Get()
	if shutdown {
		return false
	}
	err := func(obj interface{}) error {
		defer n.workqueue.Done(obj)
		var key string
		var ok bool
		var rqo *VirtulMachineQueueObj
		if rqo, ok = obj.(*VirtulMachineQueueObj); !ok {
			n.workqueue.Forget(obj)
			return fmt.Errorf("expected NodeQueueObj in workqueue but got %#v", obj)
		}
		// 在syncHandler中处理业务
		if err := n.syncHandler(rqo); err != nil {
			return fmt.Errorf("error syncing '%s': %s", key, err.Error())
		}

		n.workqueue.Forget(obj)
		return nil
	}(obj)
	if err != nil {
		runtime.HandleError(err)
	}
	return true
}

func (n *VirtulMachineListenerController) syncHandler(rqo *VirtulMachineQueueObj) error {
	key := rqo.Key
	switch {
	case rqo.Ope == common.ADD:
		return n.add(key)
	case rqo.Ope == common.UPDATE:
		// 1.first add new route config
		if err := n.add(key); err != nil {
			// log error
			return err
		} else {
			// 2.then delete routes not exist
			return n.sync(rqo)
		}
	case rqo.Ope == common.DELETE:
		return n.sync(rqo)
	default:
		// log error
		return fmt.Errorf("VirtulMachineQueueObj is not expected")
	}
}

func (n *VirtulMachineListenerController) add(key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		logs.Error("invalid resource key: %s", key)
		return fmt.Errorf("invalid resource key: %s", key)
	}

	node, err := n.nodeList.VirtulMachines(namespace).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			logs.Info("vm %s is removed", key)
			return nil
		}
		runtime.HandleError(fmt.Errorf("failed to list vm  %s/%s", key, err.Error()))
		return err
	}
	if node.Status.Phase == "" {
		node.Status.Phase = "Running"
		_, err = n.nodeClientset.NodecontrollerV1alpha1().VirtulMachines(namespace).UpdateStatus(node)
		if err != nil {
			logs.Error("Update status error ,", err.Error())
		}
		fmt.Println(node.Status.Phase)
	}

	return err
}

// sync
// 1.diff routes between old and new objects
// 2.delete routes not exist
func (n *VirtulMachineListenerController) sync(rqo *VirtulMachineQueueObj) error {
	key := rqo.Key
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		logs.Error("invalid resource key: %s", key)
		return fmt.Errorf("invalid resource key: %s", key)
	}

	switch {
	case rqo.Ope == common.UPDATE:
		node, err := n.nodeList.VirtulMachines(namespace).Get(name)
		if err != nil {
			if errors.IsNotFound(err) {
				logs.Info("vm %s is removed", key)
				return nil
			}
			runtime.HandleError(fmt.Errorf("failed to list vm %s/%s", key, err.Error()))
			return err
		}
		fmt.Println(node.Name, node.Status.Phase)
		fmt.Println("update")
		return nil
	case rqo.Ope == common.DELETE:
		fmt.Println("delete")
		return nil
	default:
		return fmt.Errorf("not expected in (VirtualMachineController) sync")
	}
}
