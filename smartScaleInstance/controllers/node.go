package controllers

import (
	"encoding/json"
	"github.com/astaxie/beego"
	"smartScaleInstance/models"
	"smartScaleInstance/utility"
	"time"
	"fmt"
	//"strconv"
)

type NodeController struct {
	beego.Controller
}

func (this *NodeController) GetNodeHeartBeat() {
	//nodeguid := this.Ctx.Input.Param(":guid")   //获取传递过来的guid
	//nodeip := this.Ctx.Input.Param(":ip")   //获取传递过来的ip
	//nodeport := this.Ctx.Input.Param(":port")   //获取传递过来的port
	//p,_ := strconv.Atoi(nodeport)
	//rulenum := this.Ctx.Input.Param(":rulenum")   //获取传递过来的node上运行的规则数量
	//r,_ := strconv.Atoi(rulenum)
	//maxrulenum := this.Ctx.Input.Param(":maxrulenum")   //获取传递过来的node上运行运行的最大规则数量
	//m,_ := strconv.Atoi(maxrulenum)
	//for index,element := range models.TotalNodes{
	//	if element.Guid == nodeguid {
	//		models.TotalNodes[index].LastActiveTime=time.Now()
	//		models.TotalNodes[index].IsActive=true
	//		models.TotalNodes[index].CurrentApps=r
	//		this.Ctx.Output.SetStatus(200)
	//		this.Ctx.Output.Body([]byte("send heartbeart ok"))
	//		return
	//	}
	//}
	req :=map[string]interface{}{}
	if err := json.Unmarshal(this.Ctx.Input.RequestBody, &req); err != nil {
		this.Ctx.Output.Body([]byte("send heatbeat parameter json is null!"))
		this.Ctx.Output.SetStatus(400)
		return
	}
	nodeip := utility.SwitchTypeToString(req["nodeip"])
	nodeport := utility.SwitchTypeToInt(req["nodeport"])
	for _,v := range models.TotalNodes {
		if v.Guid == utility.SwitchTypeToString(req["nodeguid"]) {
			v.LastActiveTime = time.Now()
			v.IsActive = true
			v.MaxApps = utility.SwitchTypeToInt(req["maxrules"])
			this.Ctx.Output.SetStatus(200)
			this.Ctx.Output.Body([]byte(fmt.Sprintf("接收到ip是%s,端口是%d心跳成功",nodeip,nodeport)))
			return
		}
	}

	this.Ctx.Output.SetStatus(400)
	this.Ctx.Output.Body([]byte(fmt.Sprintf("找不到节点ip为%s 端口为%d的信息，该节点是新节点",nodeip,nodeport)))

}


func (this *NodeController) NewNode() {
	req := struct{ Ip string; Port int ;MaxApps int }{}
	if err := json.Unmarshal(this.Ctx.Input.RequestBody, &req); err != nil {
		this.Ctx.Output.SetStatus(400)
		this.Ctx.Output.Body([]byte("ip empty!"))    //返回的是一串字符串
		return
	}
	node,err :=models.NewNode(req.Ip,req.Port,req.MaxApps)
	if err != nil {
		this.Ctx.Output.SetStatus(400)
		this.Ctx.Output.Body([]byte("add node failed"))    //返回的是一串字符串
		return
	}
	if isNodeInPool := models.AddNodeToPool(node); isNodeInPool == true{
		fmt.Printf("ip是%s端口是%d的节点之前已经在controller注册，controller恢复该节点的状态成功\n",req.Ip,req.Port)
		this.Ctx.Output.SetStatus(200)
		tempStru := struct{Guid string}{}
		tempStru.Guid=node.Guid
		tempJson,_ := json.Marshal(tempStru)
		this.Ctx.Output.Body([]byte(string(tempJson)))
		return
	}

	nodeitem := models.NodeTable{NodeGuid: node.Guid, NodeIp: req.Ip, NodePort: req.Port,MaxApps: req.MaxApps}
	if err = models.AddNodeInDB(&nodeitem); err != nil {
		this.Ctx.Output.SetStatus(400)
		this.Ctx.Output.Body([]byte("添加节点的信息到数据库失败，请检查数据库的连接状态是否正常"))    //返回的是一串字符串
		return
	}


	this.Ctx.Output.SetStatus(200)
//	this.Ctx.Output.Body([]byte("add node success!"))
	tempStru := struct{Guid string}{}
	tempStru.Guid=node.Guid
	//this.Data["json"] = `{"guid":"`+ node.Guid+ `"}`   //返回的是json串
	tempJson,_ := json.Marshal(tempStru)
	//this.Data["json"]=string(tempJson)
	//fmt.Println(string(tempJson))
	this.Ctx.Output.Body([]byte(string(tempJson)))
}

func (this *NodeController) ListNodes(){
	res := struct{ Nodes []*models.Node }{models.TotalNodes}
	this.Data["json"] = res
	this.ServeJSON()
}

func RecoverNodePool() {
	allNodes,err := models.GetAllNodesFromDB()
	if err != nil {
		fmt.Println("无法从数据库读取节点的信息，恢复节点池失败，请确认数据库运行正常")
		return
	}
	for _,v := range allNodes {
		//获取该节点上正在运行的规则数量
		currentApps:= len(models.GetRulesInNode(v.NodeGuid))
		node := &models.Node{Guid: v.NodeGuid,LastActiveTime: time.Now(),IsActive: true, IsRestart: false, MaxApps: v.MaxApps,CurrentApps: currentApps,IsNeedMigrate: false,IsMigrating: false,Ip: v.NodeIp, Port : v.NodePort}
		models.Save(node)
	}
}

