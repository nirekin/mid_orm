package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"strings"
	"time"
)

const (
	SELECT_ERP_MYSQL = "SELECT TABLE_NAME FROM information_schema.tables WHERE TABLE_SCHEMA = ?"
)

type VisibleErp struct {
	CreatedAt   time.Time          `xml:"creationDate"`
	UpdatedAt   time.Time          `xml:"updateDate"`
	TypeInt     int                `xml:"typeInt"`
	TypeName    string             `xml:"type"`
	VisibleName string             `xml:"name"`
	Value       string             `xml:"value"`
	Entries     []*VisibleErpEntry `xml:"entries>entry"`
}

type Erp struct {
	Id          int
	CreatedAt   time.Time
	UpdatedAt   time.Time
	TypeInt     int
	TypeName    string
	VisibleName string
	Value       string
	Sources     []ErpSource
	Entries     []Entry
}

func (p *Erp) loadErpEntries() {
	iDB.Where("erp_id = ?", p.Id).Find(&p.Entries)
}

func (p *Erp) lazyLoadTables() error {
	if p.TypeInt == MYSQL_TYPE {
		desiredSchema := getMySqlSchema(p.Value)
		dbCErp, err := sql.Open("mysql", p.Value)
		if err != nil {
			return err
		}
		defer dbCErp.Close()

		var tResult [10]ErpSource
		result := tResult[0:0]

		st, err := dbCErp.Prepare(SELECT_ERP_MYSQL)
		if err != nil {
			return err
		}
		rows, err := st.Query(desiredSchema)
		if err != nil {
			return err
		}

		nameLoaded := false
		for rows.Next() {
			nameLoaded = true
			e := ErpSource{}
			e.ErpId = p.Id
			err := rows.Scan(&e.Name)
			if err != nil {
				return err
			}
			e.loadUsed()
			result = append(result, e)
		}
		if nameLoaded {
			p.Sources = result
		} else {
			p.Sources = make([]ErpSource, 0)
		}
		return nil
	} else {
		result := make([]ErpSource, 1)
		result[0].Name = "ERP Type not implemented yet"
		p.Sources = result
		return nil
	}
	return nil
}

func (p *Erp) HasSources() bool {
	return len(p.Sources) > 0
}

func (o *Erp) AfterDelete() error {
	deleteEntries(*o)
	return nil
}

func getMySqlSchema(value string) string {
	s := strings.Split(value, "/")
	desiredSchema := s[len(s)-1]
	return desiredSchema
}
