# iris v12 middleware

## Sample code
### log middleware
> https://github.com/kataras/iris/blob/v12/middleware/logger/logger.go
```
func main() {
    app := iris.New()
    app.Use(logmiddleware.New())
   	app.Get("/api/v1/test", func(ctx iris.Context) {
   		ctx.JSON(iris.Map{"status": 500, "msg": "ok", "data": nil})
   	})
    app.Run(iris.Addr(":8080"))
}
```