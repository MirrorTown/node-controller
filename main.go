package main

import (
	"flag"
	"github.com/astaxie/beego"
	"node-controller/conf"
	"node-controller/controller"
	nodeInformer "node-controller/generated/informers/externalversions"
	"node-controller/initial"
	_ "node-controller/models"
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

	//c.PodListener()
	c.WorkerListener()
	go c.CoreSharedInformerFactory.Start(stop)

	c.VirtulMachineListener()
	go func() {
		time.Sleep(time.Duration(5) * time.Second)
		c.SharedInformerFactory.Start(stop)
	}()

	if beego.BConfig.RunMode == "dev" {
		beego.BConfig.WebConfig.DirectoryIndex = true
		beego.BConfig.WebConfig.StaticDir["/swagger"] = "swagger"
	}

	initial.InitDb()

	beego.Run()
}
