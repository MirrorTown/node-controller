package common

import (
	"encoding/json"
	"github.com/astaxie/beego"
	v1 "k8s.io/api/core/v1"
	"node-controller/util/logs"
	"runtime"
	"time"
)

type Markdown struct {
	Title string `json:"title"`
	Text  string `json:"text"`
}

type At struct {
	AtMobiles string `json:"atMobiles"`
	IsAtAll   string `json:"isAtAll"`
}

func ContainersRequestResourceList(containers []v1.Container) *ResourceList {
	var cpuUsage, memoryUsage int64
	for _, container := range containers {
		// unit m
		cpuUsage += container.Resources.Requests.Cpu().MilliValue()
		// unit Byte
		memoryUsage += container.Resources.Requests.Memory().Value()
	}
	return &ResourceList{
		Cpu:    cpuUsage,
		Memory: memoryUsage,
	}
}

func Send2Hook(content []Ready2Send, now string, t string, url string) {
	defer func() {
		if e := recover(); e != nil {
			buf := make([]byte, 16384)
			buf = buf[:runtime.Stack(buf, false)]
			logs.Error("Panic in Send2Hook:%v\n%s", e, buf)
		}
	}()

	if t == "recover" {
		for _, i := range content {
			data, _ := json.Marshal(
				struct {
					Msgtype  string   `json:"msgtype"`
					Markdown Markdown `json:"markdown"`
					At       At       `json:"at"`
				}{
					Msgtype: "markdown",
					Markdown: Markdown{
						Title: i.Title,
						Text:  i.Alerts,
					},
					At: At{
						AtMobiles: i.User,
						IsAtAll:   "false",
					},
				})
			HttpPost(url, nil, nil, data)
		}
	} else {
		for _, i := range content {
			i.Alerts = "- [告警名称] " + i.Title + "\n" +
				"- [K8s事件]" + i.Alerts + "\n" +
				"- [操作确认]" + "(" + beego.AppConfig.String("WebUrl") + "/alerts_confirm/" + "?start=" + i.Start + ")"
			data, _ := json.Marshal(
				struct {
					Msgtype  string   `json:"msgtype"`
					Markdown Markdown `json:"markdown"`
					At       At       `json:"at"`
				}{
					Msgtype: "markdown",
					Markdown: Markdown{
						Title: i.Title,
						Text:  i.Alerts,
					},
					At: At{
						AtMobiles: i.User,
						IsAtAll:   "false",
					},
				})
			HttpPost(url, nil, map[string]string{"Content-Type": "application/json"}, data)
			time.Sleep(1 * time.Second)
		}
	}
}
