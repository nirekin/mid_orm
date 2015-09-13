package main

import (
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"html/template"
	"net/http"
	"runtime"
	"strconv"
)

var iDB gorm.DB

var jsonHtmlTmpl = template.Must(template.New("jsonHtml").Parse(`
	<pre>{{.}}</pre>
`))

var xmlHtmlTmpl = template.Must(template.New("xmlHtml").Parse(`
	<pre>{{.}}</pre>
`))

const (
	// MYSQL
	MYSQL_TYPE       = 1
	MYSQL_TYPE_SPLIT = "__/$/__"
	MYSQL_TYPE_EMPTY = "__/#/__"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	defer fmt.Printf("stopped \n")

	go StartSync()

	http.HandleFunc("/admin/", adminHandler)
	http.HandleFunc("/erpsources/", erpsourcesHandler)
	http.HandleFunc("/addMySQL/", addMySQLHandler)
	http.HandleFunc("/createMySQL/", createMySQLHandler)
	http.HandleFunc("/editErp/", editErpHandler)
	http.HandleFunc("/deleteErp/", deleteErpHandler)
	http.HandleFunc("/updateMySQL/", updateMySQLHandler)
	http.HandleFunc("/erpListTables/", erpListTablesHandler)

	http.HandleFunc("/createErpEntry/", createErpEntryHandler)
	http.HandleFunc("/erpentries/", erpentriesHandler)
	http.HandleFunc("/deleteErpEntry/", deleteErpEntryHandler)
	http.HandleFunc("/editErpEntry/", editEntryHandler)
	http.HandleFunc("/updateErpEntry/", updateErpEntryHandler)

	http.HandleFunc("/erpListFields/", erpListFieldsHandler)
	http.HandleFunc("/createSyncField/", createSyncFieldHandler)
	http.HandleFunc("/deleteSyncField/", deleteSyncFielddHandler)
	http.HandleFunc("/editSyncField/", editSyncFieldHandler)
	http.HandleFunc("/updateSyncField/", updateSyncFieldHandler)

	http.HandleFunc("/addDecorator/", addDecoratorHandler)
	http.HandleFunc("/predefinedDecorator/", predefinedDecoratorHandler)
	http.HandleFunc("/requestAddDecorator/", requestAddDecoratorHandler)
	http.HandleFunc("/deleteDecorator/", deleteDecoratorHandler)
	http.HandleFunc("/deleteDecoratorInAdd/", deleteDecoratorInAddHandler)
	http.HandleFunc("/requestAddDecoratoParam/", requestAddDecoratorParamHandler)

	http.HandleFunc("/pingAsyncErpEntry/", pingAsyncErpEntryHandler)
	http.HandleFunc("/pingAsyncTestErpEntry/", pingAsyncTestErpEntryHandler)

	http.HandleFunc("/syncErpEntry/", syncErpEntryHandler)

	http.HandleFunc("/configJSon/", configJSonHandler)
	http.HandleFunc("/configXml/", configXmlHandler)

	http.Handle("/", http.FileServer(http.Dir("./resources")))
	http.Handle("/resources/", http.StripPrefix("/resources/", http.FileServer(http.Dir("resources"))))

	http.ListenAndServe(":8090", nil)
}

func adminHandler(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("./template/admin.html")
	t.Execute(w, nil)
}

func addMySQLHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/createMySQL.html", http.StatusFound)
}

func erpsourcesHandler(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("./template/erpsources.html")
	var erps []Erp
	iDB.Find(&erps)
	t.Execute(w, erps)
}

func createMySQLHandler(w http.ResponseWriter, r *http.Request) {
	o := &Erp{}
	o.TypeInt = MYSQL_TYPE
	o.TypeName = "MySql"
	o.VisibleName = sFName(r)
	o.Value = sFValue(r)
	iDB.Create(o)
	http.Redirect(w, r, "/erpsources", http.StatusFound)
}

func editErpHandler(w http.ResponseWriter, r *http.Request) {
	i, _ := readIntUrl(r)
	erp := &Erp{Id: i}
	iDB.First(erp)
	if erp.TypeInt == MYSQL_TYPE {
		t, _ := template.ParseFiles("./template/erpEditMySql.html")
		t.Execute(w, erp)
	} else {
		t, _ := template.ParseFiles("./template/erpEditTODO.html")
		t.Execute(w, erp)
	}
}

func updateMySQLHandler(w http.ResponseWriter, r *http.Request) {
	i, _ := iFId(r)
	erp := &Erp{Id: i}
	iDB.First(erp)
	erp.VisibleName = sFName(r)
	erp.Value = sFValue(r)
	iDB.Save(erp)
	http.Redirect(w, r, "/erpsources", http.StatusFound)
}

func deleteErpHandler(w http.ResponseWriter, r *http.Request) {
	i, _ := readIntUrl(r)
	erp := &Erp{Id: i}
	iDB.Delete(erp)
	http.Redirect(w, r, "/erpsources", http.StatusFound)
}

func erpListTablesHandler(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("./template/erpListTables.html")
	i, _ := readIntUrl(r)
	erp := &Erp{Id: i}
	iDB.First(erp)
	err := erp.lazyLoadTables()
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	t.Execute(w, erp)
}

func createErpEntryHandler(w http.ResponseWriter, r *http.Request) {
	entry := &Entry{}
	i, _ := iFId(r)
	entry.ErpId = i
	entry.TechnicalName = sFSourceName(r)
	entry.VisibleName = entry.TechnicalName
	iDB.Create(entry)
	http.Redirect(w, r, "/editErp/"+strconv.Itoa(i), http.StatusFound)
}

func erpentriesHandler(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("./template/erpentries.html")
	var entries []Entry
	iDB.Find(&entries)
	t.Execute(w, entries)
}

func deleteErpEntryHandler(w http.ResponseWriter, r *http.Request) {
	i, _ := readIntUrl(r)
	entry := &Entry{Id: i}
	iDB.Delete(entry)
	http.Redirect(w, r, "/erpentries", http.StatusFound)
}

func editEntryHandler(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("./template/editErpEntry.html")
	i, _ := readIntUrl(r)
	entry := &Entry{Id: i}
	iDB.Preload("Decorators").Preload("SyncFields").First(entry)
	t.Execute(w, entry)
}

func updateErpEntryHandler(w http.ResponseWriter, r *http.Request) {
	i, _ := iFId(r)
	entry := &Entry{Id: i}
	iDB.First(entry)
	entry.VisibleName = sFName(r)
	bs, _ := iFBlock(r)
	entry.BlockSize = bs
	iDB.Save(entry)
	http.Redirect(w, r, "/erpentries", http.StatusFound)
}

func erpListFieldsHandler(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("./template/erpListFields.html")
	i, _ := readIntUrl(r)
	entry := &Entry{Id: i}
	iDB.First(entry)
	entry.lazyLoadRFields()
	t.Execute(w, entry)
}

func createSyncFieldHandler(w http.ResponseWriter, r *http.Request) {
	field := &SyncField{}
	field.EntryId, _ = iFId(r)
	field.TechnicalName = sFFieldName(r)
	field.JsonName = field.TechnicalName
	iDB.Save(field)
	http.Redirect(w, r, "/editErpEntry/"+strconv.Itoa(field.EntryId), http.StatusFound)
}

func deleteSyncFielddHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("deleteSyncFielddHandler\n")
	idf, _ := iFField(r)
	field := &SyncField{Id: idf}
	iDB.Delete(field)
	http.Redirect(w, r, "/editErpEntry/"+sFEntry(r), http.StatusFound)
}

func editSyncFieldHandler(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("./template/editSyncField.html")
	i, _ := readIntUrl(r)
	field := &SyncField{Id: i}
	iDB.Preload("Decorators").First(field)

	//_ = field.loadDbDecorators()
	t.Execute(w, field)
}

func updateSyncFieldHandler(w http.ResponseWriter, r *http.Request) {
	fieldId := sFId(r)
	jsonName := sFJSonName(r)
	i, _ := strconv.Atoi(fieldId)
	field := &SyncField{Id: i}
	iDB.Preload("Decorators").First(field)

	field.EntryPk = eqstring(r.FormValue("ErpPk"), "on")
	field.JsonName = jsonName
	iDB.Save(field)
	http.Redirect(w, r, "/editSyncField/"+fieldId, http.StatusFound)
}

func addDecoratorHandler(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("./template/addDecorator.html")
	i, _ := readIntUrl(r)
	field := &SyncField{Id: i}
	iDB.Preload("Decorators").First(field)
	t.Execute(w, field)
}

func requestAddDecoratorHandler(w http.ResponseWriter, r *http.Request) {
	decI, _ := iFDec2(r)
	fieldId := sFField(r)
	idf, _ := iFField(r)
	predefinedDec := decorators[decI]

	pDec := decorators[decI]
	if pDec.Params == nil {
		t, _ := template.ParseFiles("./template/addDecorator.html")
		field := &SyncField{Id: idf}
		iDB.Preload("Decorators").First(field)

		d := &Decorator{}
		d.DecoratorId = decI
		d.Params = ""
		d.SyncFieldId = idf
		d.SortingOrder = len(field.Decorators) + 1
		d.Name = predefinedDec.Name
		d.Description = predefinedDec.Description
		field.Decorators = append(field.Decorators, *d)
		iDB.Save(field)

		t.Execute(w, field)
	} else {
		t, _ := template.ParseFiles(pDec.Template)
		rac := &RequestAddContent{}
		rac.FieldId = idf
		rac.DecoratorId = decI
		rac.PDecorator = decorators[decI]
		t.Execute(w, rac)
	}
	http.Redirect(w, r, "/editSyncField/"+fieldId, http.StatusNotFound)
}

func deleteDecoratorHandler(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("./template/editSyncField.html")
	idf, _ := iFField(r)
	idd, _ := iFDec(r)

	d := &Decorator{Id: idd}
	iDB.Delete(d)

	field := &SyncField{Id: idf}
	iDB.Preload("Decorators").First(field)
	field.reOrderDecorators()
	iDB.Save(field)
	t.Execute(w, field)
}

func deleteDecoratorInAddHandler(w http.ResponseWriter, r *http.Request) {
	fieldId := sFField(r)
	idd, _ := iFDec(r)
	idf, _ := iFField(r)

	d := &Decorator{Id: idd}
	iDB.Delete(d)

	field := &SyncField{Id: idf}
	iDB.Preload("Decorators").First(field)
	field.reOrderDecorators()
	iDB.Save(field)

	http.Redirect(w, r, "/addDecorator/"+fieldId, http.StatusFound)
}

func requestAddDecoratorParamHandler(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("./template/addDecorator.html")
	decI, _ := iFDec2(r)
	idf, _ := iFField(r)

	field := &SyncField{Id: idf}
	iDB.Preload("Decorators").First(field)

	predefinedDec := decorators[decI]
	params := predefinedDec.Params

	var content string
	for i := 0; i < len(params); i++ {
		sp := r.FormValue(params[i].Name)
		content += fmt.Sprintf("\"%s\":\"%s\",", params[i].Name, sp)
	}
	content = fmt.Sprintf("{%s}", content[:len(content)-1])

	d := &Decorator{DecoratorId: decI, SyncFieldId: idf}
	d.Name = predefinedDec.Name
	d.Description = predefinedDec.Description
	d.Params = content
	d.SortingOrder = len(field.Decorators) + 1
	field.Decorators = append(field.Decorators, *d)
	iDB.Save(field)

	t.Execute(w, field)
}

func pingAsyncTestErpEntryHandler(w http.ResponseWriter, r *http.Request) {
	idf, _ := iFField(r)
	field := &SyncField{Id: idf}
	iDB.Preload("Decorators").First(field)

	testContent := sFTextContent(r)

	mapD := map[string]string{}
	fN, val := field.decorate(testContent)
	mapD[fN] = val
	outJson, _ := json.Marshal(mapD)
	w.Write([]byte(testContent + "<BR>"))
	w.Write([]byte(string(outJson) + "<BR>"))
}

func pingAsyncErpEntryHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("pingAsyncErpEntryHandler \n")
	i, _ := readIntUrl(r)
	entry := &Entry{Id: i}
	iDB.Preload("SyncFields").First(entry)
	lines, _ := entry.ping(20)
	for _, val := range lines {
		fmt.Printf("line %v\n", val)
		w.Write([]byte(val + "<BR>"))
	}
}

func syncErpEntryHandler(w http.ResponseWriter, r *http.Request) {
	i, _ := readIntUrl(r)
	entry := &Entry{Id: i}
	iDB.Preload("SyncFields").First(entry)
	err := synchronize(*entry)
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	http.Redirect(w, r, "/erpentries", http.StatusFound)
}

func predefinedDecoratorHandler(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("./template/predDecorator.html")
	pd := getPredefinedDecorator()
	t.Execute(w, pd)
}

func configJSonHandler(w http.ResponseWriter, r *http.Request) {
	c := &CentralConfig{}
	s := c.toJson()
	jsonHtmlTmpl.Execute(w, template.JS(s))
}

func configXmlHandler(w http.ResponseWriter, r *http.Request) {
	c := &CentralConfig{}
	s := c.toXml()
	xmlHtmlTmpl.Execute(w, template.JS(s))
}

func init() {
	defer fmt.Printf("Init DONE\n")

	db, err := gorm.Open("mysql", "root:admin@/mid_orm?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		panic(err)
	}
	iDB = db
	initDb(db)
	initDecorators()
}

func initDb(db gorm.DB) {
	defer fmt.Printf("Init DB DONE! \n")

	db.DB().Ping()
	db.DB().SetMaxIdleConns(10)
	db.DB().SetMaxOpenConns(100)

	db.CreateTable(&Erp{})
	db.CreateTable(&Entry{})

	db.CreateTable(&SyncField{})
	db.Model(&Entry{}).Related(&SyncField{}, "EntryId")

	db.CreateTable(&Decorator{})
	db.Model(&SyncField{}).Related(&Decorator{}, "DecoratorId")

	db.CreateTable(&SyncEvent{})

}

var sFTextContent = readFormString("TestContent")
var iFDec = readFormInt("DecId")
var iFId = readFormInt("Id")
var sFName = readFormString("Name")
var sFValue = readFormString("Value")
var sFSourceName = readFormString("SourceName")
var iFBlock = readFormInt("BlockSize")
var sFFieldName = readFormString("FieldName")
var iFField = readFormInt("FieldId")
var sFEntry = readFormString("EntryId")
var sFId = readFormString("Id")
var sFJSonName = readFormString("JsonName")
var iFDec2 = readFormInt("DecoratorId")
var sFField = readFormString("FieldId")
var iFLimit = readFormInt("Limit")
var sFLikeOnPk = readFormString("LikeOnErpPk")
var sFLikeOnContent = readFormString("LikeOnContent")
