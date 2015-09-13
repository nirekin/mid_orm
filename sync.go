package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type partialEvent struct {
	Imported int64
	Updated  int64
}

func StartSync() {
	c := make(chan int)

	ticker := time.NewTicker(10 * time.Minute)
	go func() {
		i := 0
		for _ = range ticker.C {
			var entries []Entry
			iDB.Find(&entries)

			for _, val := range entries {
				entry := &Entry{Id: val.Id}
				iDB.Preload("SyncFields").First(entry)
				go synchronize(*entry)
				i++
			}
			c <- i
		}
	}()
	for {
		fmt.Printf("Sync Done %v at %v\n", <-c, time.Now())
	}
}

func synchronize(o Entry) error {
	err := o.checkImportationTableName()
	if err != nil {
		fmt.Printf("err 01 %v\n", err)
	}
	extractSentence := o.getExtractSentence()
	if extractSentence == "" {
		return nil
	}
	erp := &Erp{Id: o.ErpId}
	iDB.First(erp)

	var tKeys [10]string
	keys := tKeys[0:0]

	if erp.TypeInt == MYSQL_TYPE {
		dbCErp, err := sql.Open("mysql", erp.Value)
		if err != nil {
			return err
		} else {
			defer dbCErp.Close()
		}
		st, err := dbCErp.Prepare(extractSentence)
		if err != nil {
			return err
		} else {
			defer st.Close()
		}

		rows, err := st.Query()
		if err != nil {
			return err
		}
		var content string
		syncFs := o.SyncFields
		ecMap := make(map[string]*ExtractedContent)
		for rows.Next() {
			if err := rows.Scan(&content); err == nil {
				var pkContent string
				extractedString := strings.Split(content, MYSQL_TYPE_SPLIT)
				lExtractedString := len(extractedString)
				mapD := map[string]string{}
				for i := 0; i < lExtractedString; i++ {
					str := strings.Replace(extractedString[i], MYSQL_TYPE_EMPTY, "", -1)
					if syncFs[i].EntryPk {
						pkContent += str
					}
					fN, val := syncFs[i].decorate(str)
					mapD[fN] = val
				}
				outJson, _ := json.Marshal(mapD)
				ecMap[pkContent] = &ExtractedContent{ErpEntryId: o.Id, ErpPk: pkContent, Content: string(outJson)}
				keys = append(keys, pkContent)
			}
		}
		syncPrepare(o)
		blockSize := o.BlockSize
		if blockSize == 0 {
			blockSize = 100
		}

		chanLen := 0

		var chanEvent chan partialEvent
		var imported, updated, deleted int64 = 0, 0, 0
		iB, mod := 0, 0

		timeMSStart := getNowMillisecond()

		lenKeys := len(keys)
		if lenKeys <= blockSize {
			chanLen = 1
		} else {
			iB = lenKeys / blockSize
			mod = lenKeys % blockSize
			if mod > 0 {
				chanLen = iB + 1
			} else {
				chanLen = iB
			}
		}

		chanEvent = make(chan partialEvent, chanLen)
		if chanLen == 1 {
			go insertOrUpdate(chanEvent, o, ecMap, keys)
		} else {
			for i := 0; i < iB; i++ {
				go insertOrUpdate(chanEvent, o, ecMap, keys[i*blockSize:(i+1)*blockSize])
			}
			if mod > 0 {
				go insertOrUpdate(chanEvent, o, ecMap, keys[iB*blockSize:lenKeys])
			}
		}

		for i := 0; i < chanLen; i++ {
			s := <-chanEvent
			imported = imported + s.Imported
			updated = updated + s.Updated
		}
		deleted, _ = syncClean(o)
		timeMSStop := getNowMillisecond()
		_ = addEvent(o, imported, updated, deleted, timeMSStop-timeMSStart, int64(len(keys)))
		return nil
	} else {
		return nil
	}
}

func syncPrepare(e Entry) error {
	iDB.Exec("UPDATE " + e.getImportationTableSchema() + " SET processedFromERP=0")
	return nil
}

func syncClean(e Entry) (int64, error) {
	var cptBefore, cptAfter int64 = 0, 0
	iDB.Table(e.getImportationTable()).Count(&cptBefore)
	sql := fmt.Sprintf("DELETE FROM %v WHERE processedFromERP=0", e.getImportationTableSchema())
	iDB.Exec(sql)
	iDB.Table(e.getImportationTable()).Count(&cptAfter)
	return cptBefore - cptAfter, nil
}

func insertOrUpdate(ch chan partialEvent, entry Entry, ec ExtractedContentMap, keys []string) error {
	var inserted, updated int64 = 0, 0
	var cptE int
	for _, k := range keys {
		c := ec[k]
		iDB.Table(entry.getImportationTable()).Where(fmt.Sprintf("erpPk = '%v'", c.ErpPk)).Count(&cptE)
		if cptE == 1 {
			iDB.Table(entry.getImportationTable()).Where(fmt.Sprintf("erpPk = '%v'", c.ErpPk)).Where(fmt.Sprintf("content = '%v'", c.Content)).Count(&cptE)
			if cptE == 0 {
				sql := fmt.Sprintf("UPDATE %v SET content='%v', lastUpdate=%v, processedFromERP=1 WHERE erpPk='%v'", entry.getImportationTableSchema(), c.Content, getNowMillisecond(), c.ErpPk)
				iDB.Exec(sql)
				updated++
			} else {
				sql := fmt.Sprintf("UPDATE %v SET processedFromERP=1 WHERE erpPk='%v'", entry.getImportationTableSchema(), c.ErpPk)
				iDB.Exec(sql)
			}
		} else {
			n := getNowMillisecond()
			sql := fmt.Sprintf("INSERT %v SET active=1, content='%v', creationDate=%v, erpPk='%v', lastUpdate=%v, name='%v', processedFromERP=1", entry.getImportationTableSchema(), c.Content, n, c.ErpPk, n, entry.TechnicalName)
			iDB.Exec(sql)
			inserted++
		}
	}
	ch <- partialEvent{inserted, updated}
	return nil
}
