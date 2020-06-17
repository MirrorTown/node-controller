package controller

import (
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	clientSet "node-controller/generated/clientset/versioned"
	"node-controller/generated/informers/externalversions"
)

type VirtulMachineController struct {
	KubeClientSet             kubernetes.Interface
	VirtulMachineClientSet    clientSet.Interface
	SharedInformerFactory     externalversions.SharedInformerFactory
	CoreSharedInformerFactory informers.SharedInformerFactory
	Stop                      chan struct{}
}

func (c *VirtulMachineController) VirtulMachineListener() {
	nlc := BuildVirtulMachineListenerController(
		c.VirtulMachineClientSet,
		c.SharedInformerFactory.Nodecontroller().V1alpha1().VirtulMachines())
	nlc.Run(c.Stop)
}

func (c *VirtulMachineController) PodListener() {
	plc := BuildPodListenerController(c.KubeClientSet)
	plc.Run(c.Stop)
}
