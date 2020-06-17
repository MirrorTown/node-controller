package controller

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	CoreListerV1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"node-controller/common"
	"node-controller/conf"
	"node-controller/util/logs"
	"time"
)

type PodListenerController struct {
	kubeClientset kubernetes.Interface
	podList       CoreListerV1.PodLister
	podSynced     cache.InformerSynced
	workqueue     workqueue.RateLimitingInterface
}

type PodQueueObj struct {
	Key    string  `json:"key"`
	OldObj *v1.Pod `json:"old_obj"`
	Ope    string  `json:"ope"` // add / update / delete
}

func BuildPodListenerController(kubeclientset kubernetes.Interface) *PodListenerController {
	controller := &PodListenerController{
		kubeClientset: kubeclientset,
		podList:       conf.PodInformer.Lister(),
		podSynced:     conf.PodInformer.Informer().HasSynced,
		workqueue:     workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "podlistener"),
	}

	conf.PodInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    controller.addFunc,
			UpdateFunc: controller.updateFunc,
			DeleteFunc: controller.deleteFunc,
		})

	return controller
}

func (c *PodListenerController) addFunc(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		runtime.HandleError(err)
		return
	}

	rqo := &PodQueueObj{
		Key:    key,
		OldObj: nil,
		Ope:    common.ADD,
	}

	c.workqueue.AddRateLimited(rqo)
}

func (c *PodListenerController) updateFunc(oldObj, newObj interface{}) {
	oldPod := oldObj.(*v1.Pod)
	newPod := newObj.(*v1.Pod)
	if oldPod.ResourceVersion == newPod.ResourceVersion {
		return
	}

	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(newPod); err != nil {
		runtime.HandleError(err)
		return
	}
	rqo := &PodQueueObj{
		Key:    key,
		OldObj: oldPod,
		Ope:    common.UPDATE,
	}
	c.workqueue.AddRateLimited(rqo)
}

func (c *PodListenerController) deleteFunc(obj interface{}) {
	var key string
	var err error
	if key, err = cache.DeletionHandlingMetaNamespaceKeyFunc(obj); err != nil {
		runtime.HandleError(err)
		return
	}

	rqo := &PodQueueObj{
		Key:    key,
		OldObj: nil,
		Ope:    common.DELETE,
	}

	c.workqueue.AddRateLimited(rqo)
}

func (c *PodListenerController) Run(stop <-chan struct{}) error {
	//同步缓存
	if ok := cache.WaitForCacheSync(stop); !ok {
		logs.Error("同步缓存失败")
		return fmt.Errorf("failed to wait for caches to sync")
	}

	go wait.Until(c.runworker, time.Second*10, stop)
	return nil
}

func (c *PodListenerController) runworker() {
	for c.processNextWorkItem() {
	}
}

func (c *PodListenerController) processNextWorkItem() bool {
	defer recoverException()
	obj, shutdown := c.workqueue.Get()
	if shutdown {
		return false
	}
	err := func(obj interface{}) error {
		defer c.workqueue.Done(obj)
		var key string
		var ok bool
		var rqo *PodQueueObj
		if rqo, ok = obj.(*PodQueueObj); !ok {
			c.workqueue.Forget(obj)
			return fmt.Errorf("expected PodQueueObj in workqueue but got %#v", obj)
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

func (c *PodListenerController) syncHandler(rqo *PodQueueObj) error {
	key := rqo.Key
	switch rqo.Ope {
	case common.ADD:
		return c.add(key)
	case common.UPDATE:
		// 1.first add new route config
		if err := c.add(key); err != nil {
			// log error
			return err
		} else {
			// 2.then delete routes not exist
			return c.sync(rqo)
		}
	case common.DELETE:
		return c.sync(rqo)
	default:
		// log error
		return fmt.Errorf("PodQueueObj is not expected")
	}
}

func (c *PodListenerController) add(key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		logs.Error("invalid resource key: %s", key)
		return fmt.Errorf("invalid resource key: %s", key)
	}

	pod, err := c.podList.Pods(namespace).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			logs.Info("vm %s is removed", key)
			return nil
		}
		runtime.HandleError(fmt.Errorf("failed to list vm  %s/%s", key, err.Error()))
		return err
	}
	fmt.Println(pod.Name, pod.Status.Phase)

	return err
}

// sync
// 1.diff routes between old and new objects
// 2.delete routes not exist
func (c *PodListenerController) sync(rqo *PodQueueObj) error {
	key := rqo.Key
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		logs.Error("invalid resource key: %s", key)
		return fmt.Errorf("invalid resource key: %s", key)
	}

	pod, err := c.podList.Pods(namespace).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			logs.Info("vm %s is removed", key)
			return nil
		}
		runtime.HandleError(fmt.Errorf("failed to list vm %s/%s", key, err.Error()))
		return err
	}
	fmt.Println(pod.Name, pod.Status.Phase)
	switch {
	case rqo.Ope == common.UPDATE:
		fmt.Println("update")
		return nil
	case rqo.Ope == common.DELETE:
		fmt.Println("delete")
		return nil
	default:
		return fmt.Errorf("not expected in (PodController) sync")
	}
}
