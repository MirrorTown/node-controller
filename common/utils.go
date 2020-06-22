package common

import (
	"encoding/json"
	"github.com/astaxie/beego"
	v1 "k8s.io/api/core/v1"
	"node-controller/util/logs"
	"runtime"
	"time"
)

type Text struct {
	Title   string `json:"title"`
	Content string `json:"content"`
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
					Msgtype string `json:"msgtype"`
					Text    Text   `json:"text"`
					At      At     `json:"at"`
				}{
					Msgtype: "text",
					Text: Text{
						Title:   i.Title,
						Content: i.Alerts,
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
			i.Alerts = i.Alerts + "===" + beego.AppConfig.String("WebUrl") + "/alerts_confirm/" + "?start=" + i.Start
			data, _ := json.Marshal(
				struct {
					Msgtype string `json:"msgtype"`
					Text    Text   `json:"text"`
					At      At     `json:"at"`
				}{
					Msgtype: "text",
					Text: Text{
						Title:   i.Title,
						Content: i.Alerts,
					},
					At: At{
						AtMobiles: i.User,
						IsAtAll:   "false",
					},
				})
			HttpPost(url, nil, nil, data)
			time.Sleep(1 * time.Second)
		}
	}
}
