package main

import (
	"encoding/json"
	"encoding/xml"
)

type CentralConfig struct {
	Erps          []Erp
	visibleConfig *visibleConfig
}

type visibleConfig struct {
	VisibleErps []*VisibleErp `xml:"erps>erp"`
}

func (o *CentralConfig) loadConfig() {
	iDB.Find(&o.Erps)

	c := &visibleConfig{}
	vErps := make([]*VisibleErp, len(o.Erps))
	for er := 0; er < len(o.Erps); er++ {
		erp := o.Erps[er]

		ve := &VisibleErp{}
		ve.CreatedAt = erp.CreatedAt
		ve.TypeName = erp.TypeName
		ve.VisibleName = erp.VisibleName
		ve.TypeName = erp.TypeName
		ve.TypeInt = erp.TypeInt
		ve.Value = erp.Value

		erp.loadErpEntries()
		ve.Entries = make([]*VisibleErpEntry, len(erp.Entries))
		for en := 0; en < len(erp.Entries); en++ {
			ent := erp.Entries[en]

			vn := &VisibleErpEntry{}
			vn.CreatedAt = ent.CreatedAt
			vn.UpdatedAt = ent.UpdatedAt
			vn.TechnicalName = ent.TechnicalName
			vn.VisibleName = ent.VisibleName
			ve.Entries[en] = vn

			ent.loadSyncFields()

			vn.Fields = make([]*VisibleSyncField, len(ent.SyncFields))
			for f := 0; f < len(ent.SyncFields); f++ {
				fi := ent.SyncFields[f]

				vf := &VisibleSyncField{}
				vf.CreatedAt = fi.CreatedAt
				vf.UpdatedAt = fi.UpdatedAt
				vf.EntryPk = fi.EntryPk
				vf.TechnicalName = fi.TechnicalName
				vf.JsonName = fi.JsonName
				vn.Fields[f] = vf

				fi.loadDecorators()
				vf.Decoratos = make([]*VisibleDecorator, len(fi.Decorators))
				for d := 0; d < len(fi.Decorators); d++ {
					de := fi.Decorators[d]
					vd := &VisibleDecorator{}
					vd.CreatedAt = de.CreatedAt
					vd.UpdatedAt = de.UpdatedAt
					vd.Description = de.Description
					vd.Name = de.Name
					vd.Params = de.Params
					vd.SortingOrder = de.SortingOrder
					vf.Decoratos[d] = vd
				}
			}
		}
		vErps[er] = ve
	}
	c.VisibleErps = vErps
	o.visibleConfig = c
}

func (o *CentralConfig) toJson() string {
	o.loadConfig()
	b, _ := json.MarshalIndent(o.visibleConfig, "", "    ")
	return string(b)
}

func (o *CentralConfig) toXml() string {
	o.loadConfig()
	b, _ := xml.MarshalIndent(o.visibleConfig, "  ", "    ")
	return string(b)
}
