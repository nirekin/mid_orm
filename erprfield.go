package main

import ()

type ErpRField struct {
	EntryId int
	Name    string
	Used    int
}

const ()

func (o *ErpRField) loadUsed() error {
	var fields []SyncField
	iDB.Where("technical_name = ?", o.Name).
		Where("entry_id = ?", o.EntryId).
		Find(&fields)

	o.Used = len(fields)
	return nil
}
