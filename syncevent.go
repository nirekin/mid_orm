package main

import (
	"time"
)

type SyncEvent struct {
	Id        int
	CreatedAt time.Time
	UpdatedAt time.Time

	ErpEntryId int
	Imported   int64
	Updated    int64
	Deleted    int64
	PTime      int64
	NBEntries  int64
}

func addEvent(entry Entry, imported int64, updated int64, deleted int64, pTime int64, nbEntries int64) error {
	o := &SyncEvent{ErpEntryId: entry.Id, Imported: imported, Updated: updated, Deleted: deleted, PTime: pTime, NBEntries: nbEntries}
	iDB.Save(o)
	return nil
}
