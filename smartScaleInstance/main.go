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
	beego.InsertFilter("/v1/*", beego.BeforeRouter, controllers.FilterToken)
	beego.Router("/addnode/", &controllers.NodeController{}, "post:NewNode")
	beego.Router("/sendheartbeat/", &controllers.NodeController{}, "post:GetNodeHeartBeat")
	beego.Router("/v1/nodes/", &controllers.NodeController{}, "get:ListNodes")
	beego.Router("/v1/addrule/",&controllers.RuleController{},"post:NewRule")
	beego.Router("/v1/startrule/:ruleguid",&controllers.RuleController{},"post:StartRule")
	beego.Router("/v1/stoprule/:ruleguid",&controllers.RuleController{},"post:StopRule")
	beego.Router("/v1/deleterule/:ruleguid",&controllers.RuleController{},"delete:DeleteRule")
	beego.Router("/v1/updaterule/",&controllers.RuleController{},"post:UpdateRule")
	beego.Router("/addappinstancelog/",&controllers.RuleController{},"post:AddAppInstanceLog")
	beego.Router("/v1/getrulesforapp/:appguid",&controllers.RuleController{},"get:ShowAllRulesForApp")
	beego.Router("/v1/getinstancelog/:appguid",&controllers.RuleController{},"get:ShowAppInstanceChangeLog")
	beego.Router("/getappstatus/:appguid",&controllers.RuleController{},"get:GetAppStatus")
	beego.SetStaticPath("/views", "views")
	beego.Run()
}

