package models

import (
	"node-controller/util/logs"
	"time"
)

type RecordStatus int32

const (
	RecordAlerted    RecordStatus = 1
	RecordTobeAltert RecordStatus = 0

	TableNameRecord = "record"
)

type recordModel struct{}

type Record struct {
	Id          int64        `orm:"auto" json:"id,omitempty"`
	User        string       `orm:"null;size(128)" json:"user,omitempty"`
	HostName    string       `orm:"null;size(128)" json:"hostName,omitempty"`
	Description string       `orm:"null;size(512)" json:"description,omitempty"`
	Status      RecordStatus `orm:"default(0)" json:"status"`
	CreateTime  *time.Time   `orm:"auto_now_add;type(datetime)" json:"createTime,omitempty"`
	UpdateTime  *time.Time   `orm:"auto_now;type(datetime)" json:"updateTime,omitempty"`
}

func (*Record) TableName() string {
	return TableNameRecord
}

func (*recordModel) List() ([]Record, error) {
	var records []Record
	_, err := Ormer().QueryTable(new(Record)).
		Filter("Status", 0).
		All(&records)

	return records, err
}

func (*recordModel) Add(record *Record) error {
	record.CreateTime = nil
	_, err := Ormer().Insert(record)

	return err
}

func (*recordModel) UpdateStatus(hostName string, status RecordStatus) error {
	v := &Record{
		HostName: hostName,
	}

	if err := Ormer().Read(v, "HostName"); err != nil {
		logs.Error("为找到相应主机信息, ", err)
		return err
	} else {
		v.UpdateTime = nil
		v.Status = status
		_, err = Ormer().Update(v, "Status", "UpdateTime")
		return err
	}

	return nil
}
