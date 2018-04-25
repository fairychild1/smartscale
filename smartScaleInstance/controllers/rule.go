package controllers

import (
	"encoding/json"
	"github.com/astaxie/beego"
	"smartScaleInstance/models"
	"smartScaleInstance/utility"
	//"net/http"
	//"io/ioutil"
	"fmt"
	"time"
	//"bytes"
	//"strconv"
)
type RuleController struct {
	beego.Controller
}

func (this * RuleController) ShowAllRulesForApp (){
	appguid := this.Ctx.Input.Param(":appguid")
	rulesInGivenApp,err :=models.GetAllRulesByAppGuid(appguid)
	var ruleResult []models.RuleTable
	if err != nil {
		this.Ctx.Output.SetStatus(400)
		this.Ctx.Output.Body([]byte("根据应用的guid查询数据库失败，请确认该guid是否存在，或者数据库是否正常运行"))
		return
	}
	for _,m := range *rulesInGivenApp {
		// m["RuleGuid"]是规则的guid，根据这个guid查到对应的规则
		fmt.Printf("规则的guid为%s\n",m["RuleGuid"])
		rule,err := models.GetRule(utility.SwitchNullInterface(m["RuleGuid"],false))
		if err != nil {
			this.Ctx.Output.SetStatus(400)
			this.Ctx.Output.Body([]byte("根据应用的表里保存的规则的guid查询规则失败，请确认数据库是否正常运行"))
			return
		} else {
			fmt.Printf("规则名是%s\n",rule.RuleName)
			ruleResult=append(ruleResult,*rule)
			//ruleResult[rule.RuleName]= *rule
		}
	}
	fmt.Printf("该应用%s上总共有%d个规则\n",appguid,len(ruleResult))
	for k,v := range ruleResult{
		fmt.Printf("第%d条规则是%s\n",k,v.RuleName)
	}
	this.Ctx.Output.SetStatus(200)
	if len(ruleResult) == 0 {
		//this.Data["json"] = ruleResult[0]
		this.ServeJSON()
	} else {
		this.Data["json"] = ruleResult[0]
		this.ServeJSON()
	}
	//this.Data["json"] = ruleResult[0]  //基于2017年6月30日的协定，每个应用只会有一条规则，所以这里返回规则数组的的第一条。
	//this.ServeJSON()
}

func (this * RuleController) GetAppStatus() {
	//获取appguid
	appguid := this.Ctx.Input.Param(":appguid")
	status,err := utility.GetAppStatus(appguid)
	if err != nil {
		fmt.Printf("获取应用%s的状态失败\n",appguid)
		this.Ctx.Output.SetStatus(400)
		this.Ctx.Output.Body([]byte("获取应用的状态失败"))
		return
	}
	if status == true {
		fmt.Printf("应用%s处于启动状态\n",appguid)
		this.Ctx.Output.SetStatus(200)
		this.Ctx.Output.Body([]byte("应用处于启动状态"))
		return
	}
	if status == false {
		fmt.Printf("应用%s处于停止状态\n",appguid)
		this.Ctx.Output.SetStatus(200)
		this.Ctx.Output.Body([]byte("应用处于停止状态"))
		return
	}

}

func (this * RuleController) ShowAppInstanceChangeLog() {
	appguid := this.Ctx.Input.Param(":appguid")
	instancelog,err := models.GetAppInstanceChangeLogFromDB(appguid)
	if err != nil {
		this.Ctx.Output.SetStatus(400)
		this.Ctx.Output.Body([]byte("根据应用的guid查找应用的实例数变更记录失败，请确认该应用的guid存在"))
		return
	}
	type log struct {
		ChangeInfo string
		ChangeDate interface{}
	}
	var outputlog []log
	for _,i := range *instancelog {
		outputlog = append(outputlog,log{ChangeInfo:i["ChangeInfo"].(string), ChangeDate:i["ChangeDate"]})
	}
	fmt.Printf("应用%s总共有%d条实例变更记录,他们分别是\n",appguid,len(outputlog))
	for _,v := range outputlog {
		fmt.Println(v.ChangeInfo+"\n")
	}
	this.Ctx.Output.SetStatus(200)
	this.Data["json"] = outputlog
	this.ServeJSON()
}

func (this * RuleController) AddAppInstanceLog(){
	req :=map[string]interface{}{}
	if err := json.Unmarshal(this.Ctx.Input.RequestBody, &req); err != nil {
		this.Ctx.Output.Body([]byte("can't use the kong interface!"))
		this.Ctx.Output.SetStatus(400)
		return
	}
	appguid := utility.SwitchNullInterface(req["app_guid"],false)
	log := utility.SwitchNullInterface(req["app_log"],false)
	now := time.Now()
	a :=models.AppInstanceChangeTable{AppGuid:appguid, ChangeInfo: log, ChangeDate: now}
	if err :=models.AddAppInstanceChangeRecord(&a); err != nil {
		this.Ctx.Output.Body([]byte("添加应用的实例的变更记录到数据库失败"))
		this.Ctx.Output.SetStatus(400)
		return
	}
	this.Ctx.Output.Body([]byte("添加应用的实例的变更记录到数据库成功"))
	this.Ctx.Output.SetStatus(200)

}

func (this * RuleController) DeleteRule () {
	ruleguid := this.Ctx.Input.Param(":ruleguid")
	//nodeguid :=models.GetNodeByRuleGuid(ruleguid)
	var err error


	//在规则表里面删除规则的记录
	err = models.DelRule(ruleguid)
	if err != nil {
		fmt.Printf("删除数据库中规则的相关信息失败，请检查数据库是否正常运行，或者检查该规则的guid是否正确")
		this.Ctx.Output.SetStatus(400)
		this.Ctx.Output.Body([]byte("删除数据库中规则的相关信息失败，请检查数据库是否正常运行，或者检查该规则的guid是否正确"))
		return
	}

	//在规则与应用的关系表里删除规则应用关联记录
	err = models.DelRuleAppRelation(ruleguid)
	if err != nil {
		fmt.Printf("删除数据库中应用与规则的关联关系失败，请检查数据库是否正常运行，或者检查该规则的guid是否正确")
		this.Ctx.Output.SetStatus(400)
		this.Ctx.Output.Body([]byte("删除数据库中应用与规则的关联关系失败，请检查数据库是否正常运行，或者检查该规则的guid是否正确"))
		return
	}

	//在规则与节点关系表里删除规则节点关联记录
	err = models.DelNodeRuleRelation(ruleguid)
	if err != nil {
		fmt.Printf("删除数据库中规则与节点的关联关系失败，请检查数据库是否正常运行，或者检查该规则的guid是否正确")
		this.Ctx.Output.SetStatus(400)
		this.Ctx.Output.Body([]byte("删除数据库中规则与节点的关联关系失败，请检查数据库是否正常运行，或者检查该规则的guid是否正确"))
		return
	}

	fmt.Printf("删除规则%s成功\n",ruleguid)
	this.Ctx.Output.SetStatus(200)
	this.Ctx.Output.Body([]byte("删除规则成功"))

}

//启动规则
func (this * RuleController) StartRule(){
	ruleguid := this.Ctx.Input.Param(":ruleguid")
	//根据rule的guid获取这个rule绑定的app的guid.
	//app.AppActive为应用是否活跃，app.AppGuid为应用的guid
	app,err := models.GetAppByRuleGuid(ruleguid)
	if err != nil {
		this.Ctx.Output.SetStatus(400)
		this.Ctx.Output.Body([]byte(fmt.Sprintf("在规则应用关系表里找不到规则%s对应的记录，请确认该规则的guid有效",ruleguid)))
		return
	}
	//判断应用是否处于active状态，是的，才调用node的api，否则，返回提示信息，“非运行状态的应用，不能启动它的规则”
	appstatus,err := utility.GetAppStatus(app.AppGuid)
	if err != nil {
		this.Ctx.Output.SetStatus(400)
		this.Ctx.Output.Body([]byte(fmt.Sprintf("无法获取应用%s的状态,在不知道该应用状态的前提下，无法启动规则",app.AppGuid)))
		return
	} else if appstatus == false {
		this.Ctx.Output.SetStatus(400)
		this.Ctx.Output.Body([]byte(fmt.Sprintf("应用%s处于停止状态，无法启动该应用的规则",app.AppGuid)))
		return
	}

	//根据rule的guid，获取rule的完整信息
	rule,err := models.GetRule(ruleguid)
	if err != nil {
		this.Ctx.Output.SetStatus(400)
		this.Ctx.Output.Body([]byte(fmt.Sprintf("在规则的表里找不到规则%s对应的信息，请确保该规则的guid有效",ruleguid)))
		return
	}

	//调用算法，找到一个node来运行这个规则，拿到node的guid
	//根据节点负载情况挑选可以运行规则的节点
	nodeguid,nodeip,nodeport,flag :=models.ChooseNodeForRule()
	if flag == false {
		this.Ctx.Output.SetStatus(400)
		this.Ctx.Output.Body([]byte("没有可用的用来运行规则的节点"))
		return
	}

	//调用运行规则的api，将规则发送到节点
	fmt.Println("nodeguid is:"+nodeguid)
	err =utility.SendRuleToNode(rule,app.AppGuid,appstatus,nodeip,nodeport)
	if err != nil {
		this.Ctx.Output.SetStatus(400)
		this.Ctx.Output.Body([]byte("没有可以运行规则的节点，或者已经存在的节点的运行规则的数量达到了上限"))
		return
	}
	//在节点池中，将节点运行的规则数+1
	models.NodeRunningRuleAddOne(nodeguid)


	//将规则与节点的对应关系，写进数据库的第二章表，规则节点关系表里面。
	rulenode :=models.RuleNodeTable{RuleGuid:ruleguid, NodeGuid:nodeguid}
	err =models.AddNodeRuleRelation(&rulenode)
	if err != nil {
		this.Ctx.Output.SetStatus(400)
		this.Ctx.Output.Body([]byte("存储规则,节点间的映射关系到数据库失败了!"))
		return
	}

	//更新规则的状态，将rule_table中的rule_active修改为true
	err =models.UpdateRuleStatus(ruleguid,true)
	if err!= nil {
		this.Ctx.Output.SetStatus(400)
		this.Ctx.Output.Body([]byte("在数据库更新规则的状态为启动状态失败!"))
		return
	}

	this.Ctx.Output.SetStatus(200)
	this.Ctx.Output.Body([]byte(fmt.Sprintf("启动规则%s成功",ruleguid)))

}

//停止规则
func (this * RuleController) StopRule(){
	ruleguid := this.Ctx.Input.Param(":ruleguid")
	var err error
	err = stopRuleByRuleGuid(ruleguid)
	if err != nil {
		fmt.Printf("停止规则%s失败\n",ruleguid)
		this.Ctx.Output.SetStatus(400)
		this.Ctx.Output.Body([]byte("停止规则失败，请检查日志，并检查后台数据库状态，controller的状态，是否有可用的node"))
		return
	}

	fmt.Printf("停止规则%s成功\n",ruleguid)
	this.Ctx.Output.SetStatus(200)
	this.Ctx.Output.Body([]byte("停止规则成功"))
}

//更新规则
func (this * RuleController) UpdateRule() {
	req :=map[string]interface{}{}
	var err error
	if err = json.Unmarshal(this.Ctx.Input.RequestBody, &req); err != nil {
		this.Ctx.Output.Body([]byte("can't use the kong interface!"))
		this.Ctx.Output.SetStatus(400)
		return

	}

	ruleguid := utility.SwitchTypeToString(req["rule_guid"])
	rule := models.RuleTable{
		Guid : ruleguid,
		RuleName : utility.SwitchTypeToString(req["rule_name"]),
		RuleType : utility.SwitchTypeToString(req["rule_type"]),
		MinInstance : utility.SwitchTypeToInt(req["min_instance"]),
		MaxInstance : utility.SwitchTypeToInt(req["max_instance"]),
		InstanceStep : utility.SwitchTypeToInt(req["instance_step"]),
		Year : utility.SwitchTypeToInt(req["year"]),
		DayWeek : utility.SwitchTypeToInt(req["day_week"]),
		Month : utility.SwitchTypeToInt(req["month"]),
		Day : utility.SwitchTypeToInt(req["day"]),
		Hour : utility.SwitchTypeToInt(req["hour"]),
		Minute : utility.SwitchTypeToInt(req["minute"]),
		MinCpu : utility.SwitchTypeToFloat64(req["min_cpu"]),
		MaxCpu : utility.SwitchTypeToFloat64(req["max_cpu"]),
		MinMem : utility.SwitchTypeToFloat64(req["min_mem"]),
		MaxMem : utility.SwitchTypeToFloat64(req["max_mem"]),
		MinReq : utility.SwitchTypeToInt(req["min_req"]),
		MaxReq : utility.SwitchTypeToInt(req["max_req"]),
	}
	err = models.UpdateRule(&rule)
	if err != nil {
		this.Ctx.Output.SetStatus(400)
		this.Ctx.Output.Body([]byte("更新规则失败，请检查数据库的状态"))
		return
	}

	this.Ctx.Output.SetStatus(200)
	this.Ctx.Output.Body([]byte("更新规则成功"))
}

func (this * RuleController) NewRule(){
	req := struct{
		Rule_name string
		Rule_type string
		App_guid string
		Min_instance int
		Max_instance int
		Instance_step int
		Year int
		Day_week int
		Month int
		Day int
		Hour int
		Minute int
		Min_cpu float64
		Max_cpu float64
		Min_mem float64
		Max_mem float64
		Min_req int
		Max_req int
	}{}
	var err error
	if err = json.Unmarshal(this.Ctx.Input.RequestBody, &req); err != nil {
		this.Ctx.Output.SetStatus(400)
		this.Ctx.Output.Body([]byte("can't receive the request context!"))    //返回的是一串字符串
		return
	}
	//给新规则生成一个guid
	ruleguid :=models.GetGuid()


	//将规则存入数据库
	rule:=models.RuleTable{Guid:ruleguid, RuleName : req.Rule_name , RuleType:req.Rule_type, RuleActive:false, MinInstance: req.Min_instance, MaxInstance: req.Max_instance, InstanceStep: req.Instance_step, Year:req.Year, DayWeek:req.Day_week, Month:req.Month, Day:req.Day, Hour:req.Hour, Minute:req.Minute, MinCpu:req.Min_cpu, MaxCpu:req.Max_cpu, MinMem:req.Min_mem, MaxMem:req.Max_mem, MinReq:req.Min_req, MaxReq:req.Max_req}
	err =models.AddRule(&rule)
	if err != nil {
		this.Ctx.Output.SetStatus(400)
		this.Ctx.Output.Body([]byte("存储规则到数据库失败"))
		return
	}

	//将规则，应用的关系存入规则应用关系表里面
	ruleapp := models.AppRuleTable{RuleGuid:ruleguid,  AppGuid:req.App_guid}
	err = models.AddRuleApp(&ruleapp)
	if err != nil {
		this.Ctx.Output.SetStatus(400)
		this.Ctx.Output.Body([]byte("添加规则应用的关联信息到数据库失败"))
		return
	}


	this.Ctx.Output.Body([]byte("添加规则成功"))
}

func FloatToString(f float32) string{
	return fmt.Sprintf("%.2f", f)
}

func stopRuleByRuleGuid(ruleguid string) error {
	var err error
	nodeguid :=models.GetNodeByRuleGuid(ruleguid)

	//删除数据库中保存节点、规则关系表里面的记录
	err =models.DelNodeRuleRelation(ruleguid)
	if err != nil {
		fmt.Printf("删除数据库中规则、节点关系的记录失败，请确认数据库正常运行")
		return err
	}

	//查找运行规则的node的ip和port
	if nodeguid == "" {
		fmt.Printf("找不到运行规则的节点\n")
		//this.Ctx.Output.SetStatus(200)
		//this.Ctx.Output.Body([]byte("找不到运行规则的节点,有可能controller节点重启过或者之前运行该规则的node重启过，该规则已经停止了运行，不用担心"))
		return fmt.Errorf("%s","找不到运行规则的节点")
	}
	nodeip,nodeport := models.GetNodeIpAndPortByNodeGuid(nodeguid)
	if nodeip == "" && nodeport == 0 {
		fmt.Printf("找不到node的guid %s 对应的节点的ip和端口，controller有可能重启过，请等待30秒后再试\n",nodeguid)
		return fmt.Errorf("找不到node的guid %s 对应的节点的ip和端口，controller有可能重启过，请等待30秒后再试\n",nodeguid)
	}

	//发起停止规则的API请求
	err = utility.StopRule(ruleguid,nodeip,nodeport)
	if err != nil {
		fmt.Printf("发送消息到节点，请求停止运行规则失败\n")
		//this.Ctx.Output.SetStatus(400)
		//this.Ctx.Output.Body([]byte(fmt.Sprintf("发送消息到节点，请求停止运行规则失败,请检查节点%s 端口是%d的运行状态",nodeip,nodeport)))
		return fmt.Errorf("%s","发送消息到节点，请求停止运行规则失败")
	}

	//将内存块里运行该规则的节点当前运行的规则数减1
	models.NodeRunningRuleDelOne(nodeguid)

	//将规则的状态修改为停止
	err = models.UpdateRuleStatus(ruleguid,false)
	if err != nil {
		fmt.Printf("在数据库更新规则的状态为停止状态失败\n")
		//this.Ctx.Output.SetStatus(400)
		//this.Ctx.Output.Body([]byte(fmt.Sprintf("在数据库更新规则的状态为停止状态失败")))
		return fmt.Errorf("%s","在数据库更新规则的状态为停止状态失败")
	}

	return nil
}



