package main

import (
	_ "smartScaleInstance/routers"
	"github.com/astaxie/beego"
	"smartScaleInstance/controllers"
	//"smartScaleInstance/models"
	"smartScaleInstance/utility"
)

func main() {
	controllers.RecoverNodePool()
	go utility.RefreshRuleInDeadNode()
	beego.Router("/addnode/", &controllers.NodeController{}, "post:NewNode")
	beego.Router("/sendheartbeat/", &controllers.NodeController{}, "post:GetNodeHeartBeat")
	beego.Router("/nodes/", &controllers.NodeController{}, "get:ListNodes")
	beego.Router("/addrule/",&controllers.RuleController{},"post:NewRule")
	beego.Router("/startrule/:ruleguid",&controllers.RuleController{},"post:StartRule")
	beego.Router("/stoprule/:ruleguid",&controllers.RuleController{},"post:StopRule")
	beego.Router("/deleterule/:ruleguid",&controllers.RuleController{},"delete:DeleteRule")
	beego.Router("/updaterule/",&controllers.RuleController{},"post:UpdateRule")
	beego.Router("/addappinstancelog/",&controllers.RuleController{},"post:AddAppInstanceLog")
	beego.Router("/getrulesforapp/:appguid",&controllers.RuleController{},"get:ShowAllRulesForApp")
	beego.Router("/getinstancelog/:appguid",&controllers.RuleController{},"get:ShowAppInstanceChangeLog")
	beego.Router("/getappstatus/:appguid",&controllers.RuleController{},"get:GetAppStatus")
	beego.Run()
}

