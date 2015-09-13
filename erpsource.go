package main

import ()

type ErpSource struct {
	ErpId int
	Name  string
	Used  int
}

func (o *ErpSource) loadUsed() error {
	var entries []Entry
	iDB.Where("technical_name = ?", o.Name).
		Where("erp_id = ?", o.ErpId).
		Find(&entries)

	o.Used = len(entries)
	return nil
}
