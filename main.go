package main

import (
	//_ "news/docs"
	_ "news/routers"

	"github.com/astaxie/beego"
)

func main() {
	//if beego.RunMode == "dev" {
	//	beego.DirectoryIndex = true
	//	beego.StaticDir["/swagger"] = "swagger"
	//}
	beego.Run()
}
