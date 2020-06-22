package models

import (
	"github.com/astaxie/beego/orm"
	"sync"
)

var (
	globalOrm orm.Ormer
	once      sync.Once

	RecordMode *recordModel
)

func init() {
	orm.RegisterModel(new(Record))

	// init models
	RecordMode = &recordModel{}
}

// singleton init ormer ,only use for normal db operation
// if you begin transaction，please use orm.NewOrm()
func Ormer() orm.Ormer {
	once.Do(func() {
		globalOrm = orm.NewOrm()
	})
	return globalOrm
}
