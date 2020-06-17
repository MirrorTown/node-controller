package main

import (
	"flag"
	"fmt"
	"github.com/astaxie/beego"
	"k8s.io/apimachinery/pkg/labels"
	"node-controller/conf"
	"node-controller/controller"
	nodeInformer "node-controller/generated/informers/externalversions"
	_ "node-controller/routers"
	"time"
)

func main() {
	flag.Parse()
	nodeClientSet := conf.InitVirtulMachineClient()
	kubeClientSet := conf.GetKubeClient()
	sharedInformerFactory := nodeInformer.NewSharedInformerFactory(nodeClientSet, 0)
	stop := make(chan struct{})
	c := &controller.VirtulMachineController{
		KubeClientSet:             kubeClientSet,
		VirtulMachineClientSet:    nodeClientSet,
		SharedInformerFactory:     sharedInformerFactory,
		CoreSharedInformerFactory: conf.CoreSharedInformerFactory,
		Stop:                      stop,
	}

	c.PodListener()
	go c.CoreSharedInformerFactory.Start(stop)

	c.VirtulMachineListener()
	go func() {
		time.Sleep(time.Duration(5) * time.Second)
		c.SharedInformerFactory.Start(stop)
		fmt.Println("List: ")
		fmt.Println(c.SharedInformerFactory.Nodecontroller().V1alpha1().VirtulMachines().Lister().List(labels.Everything()))
	}()

	if beego.BConfig.RunMode == "dev" {
		beego.BConfig.WebConfig.DirectoryIndex = true
		beego.BConfig.WebConfig.StaticDir["/swagger"] = "swagger"
	}

	beego.Run()
}
