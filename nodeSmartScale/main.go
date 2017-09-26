package main

import (
	_ "nodeSmartScale/routers"
	"github.com/astaxie/beego"
	"nodeSmartScale/controllers"
	"nodeSmartScale/utility"
	//"fmt"
)

func main() {
	utility.RegisterSelfAsNode()
	go controllers.PeriodicSendHeartBeat()
	beego.Router("/runrule/", &controllers.RuleController{}, "post:RunRule")
	beego.Router("/stoprule/:ruleguid",&controllers.RuleController{},"post:StopRule")
	beego.Run()
}
