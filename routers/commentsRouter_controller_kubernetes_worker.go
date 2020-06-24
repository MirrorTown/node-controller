package routers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context/param"
)

func init() {

	beego.GlobalControllerRouter["node-controller/controller/kubernetes/worker:WorkerController"] = append(beego.GlobalControllerRouter["node-controller/controller/kubernetes/worker:WorkerController"],
		beego.ControllerComments{
			Method:           "List",
			Router:           `/list`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

}
