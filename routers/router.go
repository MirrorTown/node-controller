package pkg

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
	"github.com/astaxie/beego/plugins/cors"
	"net/http"
	"node-controller/util/hack"
)

func init() {
	beego.InsertFilter("*", beego.BeforeRouter, cors.Allow(&cors.Options{
		//AllowAllOrigins: true,
		AllowOrigins:     []string{"http://10.*.*.*:*", "http://localhost:*", "http://127.0.0.1:*", "http://172.*.*.*:*", "http://192.*.*.*:*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*", "content-time"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	beego.Get("/healthz", func(ctx *context.Context) {
		ctx.Output.SetStatus(http.StatusOK)
		ctx.Output.Body(hack.Slice("ok"))
	})
}
