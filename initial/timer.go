package initial

import (
	"github.com/astaxie/beego"
	"node-controller/common"
	"node-controller/models"
	"node-controller/util/logs"
	"runtime"
	"time"
)

func init() {
	go func() {
		for {
			current := time.Now()
			time.Sleep(time.Duration(60-current.Second()) * time.Second)
			now := time.Now().Format("2006-01-02 15:04:05")
			go func() {
				defer func() {
					if e := recover(); e != nil {
						buf := make([]byte, 16384)
						buf = buf[:runtime.Stack(buf, false)]
						logs.Error("Panic in timer:%v\n%s", e, buf)
					}
				}()
				var info []models.Record
				info, err := models.RecordMode.List()
				if err != nil {
					logs.Error("Query Record table error, ", err)
				}
				if len(info) == 0 {
					logs.Info("No alter need to be handler!")
					return
				}

				logs.Info("Alerts to send:%v", info)

				ready2Send := make([]common.Ready2Send, 0)
				for _, v := range info {
					singleInfo := common.Ready2Send{
						Title:  v.HostName,
						Start:  v.CreateTime.Format("2006-01-02 15:04:05"),
						User:   v.User,
						Alerts: v.Description,
					}
					ready2Send = append(ready2Send, singleInfo)
				}
				alertUrl := beego.AppConfig.String("AlertUrl")
				common.Send2Hook(ready2Send, now, "alter", alertUrl)
				//TODO: Recover msg to send
				//common.Lock.Lock()
				//recover2send := common.Recover2Send
				//common.Recover2Send = map[string]map[[2]int64]*common.Ready2Send{
				//	"LANXIN": map[[2]int64]*common.Ready2Send{},
				//	//"HOOK":   map[[2]int64]*common.Ready2Send{},
				//}
				//common.Lock.Unlock()
				//logs.Alertloger.Info("Recoveries to send:%v", recover2send)
				//RecoverSender(recover2send, now)
			}()
		}
	}()
}
