package main

import (
	"encoding/json"
	"fmt"
	"github.com/bluebuff/iris-middleware/v12/logmiddleware"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/hero"
	"strconv"
	"time"
)

var app = iris.New()
var Success = iris.Map{"status": 0, "msg": "ok", "data": nil}
var Failed = iris.Map{"status": 500, "msg": "ok", "data": nil}

func main() {
	app.Use(logmiddleware.New(logmiddleware.Config{
		Status:                    true,
		IP:                        true,
		Method:                    true,
		Path:                      true,
		Query:                     true,
		RequestBody:               true,
		ResponseBody:              true,
		MessageContextKeys:        []string{"token"},
		MessageRequestHeaderKeys:  []string{"Content-Type", "User-Agent"},
		MessageResponseHeaderKeys: []string{"Content-Type"},
		LogFunc:                   LoggerRecord,
		LogFuncCtx:                nil,
		Skippers:                  nil,
		Skip:                      skipRecord,
	}))
	app.PartyFunc("/api/v1", apiParty)
	app.Run(iris.Addr(":8080"))
}

func LoggerRecord(call *logmiddleware.ApiCall) {
	data, _ := json.Marshal(call)
	fmt.Println(string(data))
}

func skipRecord(ctx iris.Context) bool {
	logModel := ctx.URLParam("logMode")
	b, err := strconv.ParseBool(logModel)
	if err != nil {
		return false
	}
	return !b
}

func setValue(ctx iris.Context) {
	ctx.Values().Set("token", "abc123")
	ctx.Next()
}

func apiParty(party iris.Party) {
	party.Use(setValue)
	party.Get("/test", hero.Handler(doTest))
	party.Post("/post", hero.Handler(doPost))
}

func doTest(ctx iris.Context) interface{} {
	name := ctx.URLParam("name")
	age := ctx.URLParam("age")
	fmt.Println("name:", name)
	fmt.Println("age:", age)
	time.Sleep(time.Second * 2)
	return iris.Map{"status": 0, "msg": "ok", "data": nil}
}

type Student struct {
	Name string `json:"name" gorm:"name"`
	Age  uint32 `json:"age" gorm:"age"`
}

func doPost(ctx iris.Context) interface{} {
	stu := new(Student)
	if err := ctx.ReadJSON(stu); err != nil {
		return Failed
	}
	time.Sleep(time.Second)
	return iris.Map{"status": 0, "msg": "ok", "data": stu}
}
