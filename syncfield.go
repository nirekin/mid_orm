package main

import (
	"time"
)

type VisibleSyncField struct {
	CreatedAt     time.Time           `xml:"creationDate"`
	UpdatedAt     time.Time           `xml:"updateDate"`
	TechnicalName string              `xml:"technicalName"`
	JsonName      string              `xml:"jsonName"`
	EntryPk       bool                `xml:"erpPk"`
	Decoratos     []*VisibleDecorator `xml:"decorators>decorator"`
}

type SyncField struct {
	Id            int
	CreatedAt     time.Time
	UpdatedAt     time.Time
	TechnicalName string
	JsonName      string
	EntryPk       bool
	EntryId       int
	Decorators    []Decorator
}

func (p *SyncField) loadDecorators() {
	iDB.Where("sync_field_id = ?", p.Id).Find(&p.Decorators)
}

func (o *SyncField) decorate(s string) (string, string) {
	for _, val := range o.Decorators {
		s = val.decorate(s)
	}
	return o.JsonName, encodeUTF(s)
}

func (o *SyncField) NbDecorator() int {
	o.loadDecorators()
	if o.Decorators == nil {
		return 0
	} else {
		return len(o.Decorators)
	}
}

func (o *SyncField) reOrderDecorators() {
	cpt := 1
	for i := 0; i < len(o.Decorators); i++ {
		o.Decorators[i].SortingOrder = cpt
		cpt++
	}
}

func (o *SyncField) AfterDelete() error {
	deleteDecorators(*o)
	return nil
}

func deleteSyncFields(o Entry) {
	var fields []SyncField
	iDB.Where("entry_id = ?", o.Id).Find(&fields)
	for _, val := range fields {
		deleteDecorators(val)
	}
	iDB.Where("entry_id = ?", o.Id).Delete(SyncField{})
}
