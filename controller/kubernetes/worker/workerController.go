package worker

import (
	"node-controller/controller"
)

type WorkerController struct {
	controller.ResultHandlerController
}

func (c *WorkerController) URLMapping() {
	c.Mapping("List", c.List)
}

func (c *WorkerController) Prepare() {

}

// @Title List
// @Description find All Node Status
// @Success 200 {object} NodeListResult success
// @router /list [get]
func (c *WorkerController) List() {

	nodeListResult := controller.NodeList

	c.Success(nodeListResult)
}
