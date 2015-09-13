package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"strconv"
	"strings"
	"time"
)

type ExtractedContent struct {
	ErpEntryId int
	ErpPk      string
	Content    string
}

type VisibleErpEntry struct {
	CreatedAt     time.Time           `xml:"creationDate"`
	UpdatedAt     time.Time           `xml:"updateDate"`
	TechnicalName string              `xml:"technicalName"`
	VisibleName   string              `xml:"visibleName"`
	Fields        []*VisibleSyncField `xml:"fields>field"`
}

type Entry struct {
	Id            int
	CreatedAt     time.Time
	UpdatedAt     time.Time
	TechnicalName string
	VisibleName   string
	ErpId         int
	BlockSize     int
	SyncFields    []SyncField
	Fields        []ErpRField
}

type ExtractedContentMap map[string]*ExtractedContent

const (
	SELECT_TABLE_MYSQL = "select COLUMN_NAME from information_schema.columns where TABLE_SCHEMA = ? AND TABLE_NAME =?"
	DATA_TABLE_NAME    = "data_erp_entry_content_"
)

func (p *Entry) loadSyncFields() {
	iDB.Where("entry_id = ?", p.Id).Find(&p.SyncFields)
}

func (o *Entry) lazyLoadRFields() error {
	erp := &Erp{Id: o.ErpId}
	iDB.First(erp)

	if erp.TypeInt == MYSQL_TYPE {
		desiredSchema := getMySqlSchema(erp.Value)
		dbCErp, err := sql.Open("mysql", erp.Value)
		if err != nil {
			return err
		}

		var tResult [10]ErpRField
		result := tResult[0:0]

		st, err := dbCErp.Prepare(SELECT_TABLE_MYSQL)
		if err != nil {
			return err
		}
		rows, err := st.Query(desiredSchema, o.TechnicalName)
		if err != nil {
			return err
		}

		var name string
		for rows.Next() {
			e := ErpRField{EntryId: o.Id}
			err := rows.Scan(&name)
			if err != nil {
				return err
			}
			e.Name = name
			e.loadUsed()
			result = append(result, e)
		}
		o.Fields = result
		return nil
	} else {
		result := make([]ErpRField, 1)
		result[0].EntryId = o.Id
		result[0].Used = 0
		result[0].Name = "ERP Type not implemented yet"
		o.Fields = result

	}
	return nil
}

func (o *Entry) ping(nbRows int) ([]string, error) {
	l := o.SyncFields
	extractSentence := o.getExtractSentence()
	if extractSentence == "" {
		return make([]string, 0), nil
	}

	if nbRows <= 1 {
		nbRows = 1
	}
	erp := &Erp{Id: o.ErpId}
	iDB.First(erp)

	result := make([]string, nbRows)

	if erp.TypeInt == MYSQL_TYPE {
		dbCErp, err := sql.Open("mysql", erp.Value)
		if err != nil {
			return nil, err
		} else {
			defer dbCErp.Close()
		}
		st, err := dbCErp.Prepare(extractSentence)
		if err != nil {
			return nil, err
		} else {
			defer st.Close()
		}
		rows, err := st.Query()
		if err != nil {
			return nil, err
		}

		var content string
		cpt := 0
		for rows.Next() {
			err := rows.Scan(&content)
			if err == nil {
				ps := strings.Split(content, MYSQL_TYPE_SPLIT)
				lp := len(ps)
				mapD := map[string]string{}
				for i := 0; i < lp; i++ {
					str := strings.Replace(ps[i], MYSQL_TYPE_EMPTY, "", -1)
					fN, val := l[i].decorate(str)
					mapD[fN] = val
				}
				outJson, _ := json.Marshal(mapD)
				result[cpt] = string(outJson)
			}
			if cpt == nbRows-1 {
				break
			}
			cpt++
		}
	}
	return result, nil
}

func (o *Entry) getExtractSentence() string {
	erp := &Erp{Id: o.ErpId}
	iDB.First(erp)
	if erp.TypeInt == MYSQL_TYPE {
		l := o.SyncFields
		if len(l) == 0 {
			return ""
		}
		fl := fmt.Sprintf("SELECT concat_ws('%s' ", MYSQL_TYPE_SPLIT)
		for _, val := range l {
			fl += ",IF(" + val.TechnicalName + "<>\"\"," + val.TechnicalName + ",\"" + MYSQL_TYPE_EMPTY + "\")"
		}
		fl += ")"
		s := fmt.Sprintf("%s FROM %s.%s", fl, getMySqlSchema(erp.Value), o.TechnicalName)
		return s
	}
	return ""
}

func getNowMillisecond() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func (o *Entry) getImportationTableSchema() string {
	return fmt.Sprintf("`mid_orm`.`%s`", o.getImportationTable())
}

func (o *Entry) getImportationTable() string {
	return DATA_TABLE_NAME + strconv.Itoa(o.Id)
}

func (o *Entry) AfterCreate() {
	o.checkImportationTableName()
}

func (o *Entry) AfterDelete() error {
	iDB.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", o.getImportationTableSchema()))
	deleteSyncFields(*o)
	return nil
}

func (o *Entry) checkImportationTableName() error {
	// CREATE THE DATA TABLE TO STORE THE IMPORTED CONTENT
	sql := "CREATE TABLE IF NOT EXISTS " + o.getImportationTableSchema() +
		" ( `id` int(11) NOT NULL AUTO_INCREMENT,  `active` tinyint(1) NOT NULL,`content` text" +
		",  `creationDate` bigint(20) unsigned DEFAULT NULL, `erpPk` varchar(255) DEFAULT NULL," +
		"`lastUpdate` bigint(20) unsigned DEFAULT NULL,`name` varchar(255) DEFAULT NULL,`processedFromERP`" +
		" tinyint(1) NOT NULL,PRIMARY KEY (`id`)) ENGINE=InnoDB AUTO_INCREMENT=13222 DEFAULT CHARSET=latin1;"

	iDB.Exec(sql)
	return nil
}

func deleteEntries(o Erp) {
	var entries []Entry
	iDB.Where("erp_id = ?", o.Id).Find(&entries)
	for _, val := range entries {
		deleteSyncFields(val)
		iDB.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", val.getImportationTableSchema()))
	}
	iDB.Where("erp_id = ?", o.Id).Delete(Entry{})
}
