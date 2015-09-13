package main

import (
	"encoding/json"
	"fmt"
	"time"
)

type RequestAddContent struct {
	FieldId     int
	DecoratorId int
	PDecorator  *PredefinedDecorator
}

type VisibleDecorator struct {
	CreatedAt    time.Time `xml:"creationDate"`
	UpdatedAt    time.Time `xml:"updateDate"`
	Name         string    `xml:"name"`
	Description  string    `xml:"description"`
	SortingOrder int       `xml:"sortingOrder"`
	Params       string    `xml:"params"`
}

type Decorator struct {
	Id           int
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Name         string
	Description  string
	SortingOrder int
	Params       string
	DecoratorId  int
	SyncFieldId  int
}

func (o *Decorator) decorate(s string) string {
	pDef := decorators[o.DecoratorId]
	return pDef.FDecorate(pDef, o, s)
}

func (o *Decorator) getParamValue(name string) string {
	var f interface{}
	_ = json.Unmarshal([]byte(o.Params), &f)
	m := f.(map[string]interface{})
	return fmt.Sprintf("%v", m[name])
}

func deleteDecorators(o SyncField) {
	iDB.Where("sync_field_id = ?", o.Id).Delete(Decorator{})
}
