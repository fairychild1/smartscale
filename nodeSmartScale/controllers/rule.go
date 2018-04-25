package controllers

import (
	"encoding/json"
	"github.com/astaxie/beego"
	"time"
	"fmt"
	"strconv"
	//"reflect"
	"nodeSmartScale/utility"
	"nodeSmartScale/models"
	"strings"
	"net/http"
	"bytes"
)

type RuleController struct {
	beego.Controller
}

func (this * RuleController) RunRule() {
  //req := struct{App string; Active bool; Rule_type string; Year int; Day_week int; Month int; Day int; Hour int; Minute int; Min_cpu int; Max_cpu int; Min_mem int; Max_mem int; Min_instance int; Max_instance int; Instance_step int; Min_req int; Max_req int} {}
  //if err := json.Unmarshal(this.Ctx.Input.RequestBody, &req); err != nil {
  //  this.Ctx.Output.SetStatus(400)
  //  this.Ctx.Output.Body([]byte("err parameter!"))
  //  return
  //}
  //this.Data["json"] = `{"result" : "get the rule"}`
	//this.ServeJSON()
	req :=map[string]interface{}{}
	if err := json.Unmarshal(this.Ctx.Input.RequestBody, &req); err != nil {
		this.Ctx.Output.Body([]byte("can't use the kong interface!"))
		this.Ctx.Output.SetStatus(400)
		return

	}
	//for k, v := range req {
	//	fmt.Println(k,":",SwitchNullInterface(v,false))
	//	if k == "instance_step" {
	//		fmt.Println(k,":",SwitchNullInterface(v,true))
	//	}
	//}
	//启动一个新进程，运行规则。
	//新进程里面：调用小马哥的日志平台接口，查询应用的实例的cpu 内存 信息
	//根据查询到的cpu和内存信息，比对规则里的阀值设置，如果达到了，调用cc的登录的api，获取token，然后调用扩展应用的实例的api
	ruleguid := utility.SwitchNullInterface(req["rule_guid"],false)
	models.NewRuleStruct(ruleguid)

	//启动一个协程，处理controller发送过来的规则。规则内容在req里面，ruleguid是规则的guid，用来定位规则定位在models里面定义的数据结构AllRuleStruct里面的位置。管道用来触发该协程的结束。
	go RunRuleByLoop(req,ruleguid,models.AllRuleStruct[ruleguid].RuleChannel)
	utility.CurrentRuleNum += 1
	this.Data["json"] = req
	this.ServeJSON()
}

func RunRuleByLoop(v map[string]interface{},ruleguid string,quit chan bool) {
	for {
		select {
		case <- quit:
			fmt.Printf("节点停止运行规则%s\n",utility.SwitchNullInterface(v["rule_guid"],false))
			return
		default:
			DealWithRule(v,ruleguid)
			s := beego.AppConfig.String("sleeptime")
			sleeptime,_ := strconv.Atoi(s)
			fmt.Printf("规则%s对应的协程开始休眠，休眠时间为%d秒\n",ruleguid,sleeptime)
			time.Sleep(time.Duration(sleeptime) * time.Second)
			fmt.Printf("规则%s在节点上本次运行结束\n",ruleguid)
		}
	}
}

func DealWithRule(v map[string]interface{},ruleguid string){
	appguid := utility.SwitchNullInterface(v["app"],false)
	fmt.Printf("规则%s开始运行\n",ruleguid)

	//判断是否是调整时候后的窗口时间，如果是，那么就不检查应用的状态，如果不是，才检查应用的状态
	if flag := models.IfIsWindowTime(ruleguid); flag == true {
		return
	}

	//判断是否是查询日志系统的窗口时间，如果是，那么就不再查询日执行系统，如果不是，才查询日志系统里面应用的负载。
	if flag := models.IfIsLogSystemWindowTime(ruleguid); flag == true {
		return
	}

	//查询日志系统，查询完后设置查询日志系统的窗口时间
	cpuUsage,memUsage,flag,err:=utility.CallLogCollectionSystem(appguid)
	models.SetLogSystemWindowTime(ruleguid)

	if err !=nil {
		fmt.Println("从日志平台获取应用的相关信息失败\n")
		return
	} else {
		if flag == true {
			fmt.Printf("日志系统繁忙，查询应用%s的使用情况为空\n",appguid)
			return
		} else {
			fmt.Printf("cpu的平均使用是%.2f\n",cpuUsage)
			fmt.Printf("内存的平均使用是%.2f\n",memUsage)
		}
	}

	//如果是要判断内存是否超标，先查询应用的内存配额
	if strings.Contains(utility.SwitchTypeToString(v["rule_type"]),"mem") {
		fmt.Printf("规则的类型含有内存判断，也就是需要根据内存的使用情况来调整实例数\n")
		memoryQuota,err :=utility.GetMememoryQuota(appguid)
		if err != nil {
			fmt.Println(err)
			return
		} else {
			fmt.Printf("获取应用的内存配额成功,内存配额是%.2f\n",memoryQuota)
			memUsage = memUsage/memoryQuota
			fmt.Printf("内存使用了%f\n",memUsage)
		}
	}

	//根据规则和应用的负载情况，判断到底是增加实例，减少实例 还是不变
	i,flag :=utility.IfChangeAppInstance(cpuUsage,memUsage,v)
	fmt.Printf("i的值是%d , flag的值是%t\n",i,flag)

	//查询应用目前的实例数
	num,err := utility.GetInstanceNum(utility.SwitchNullInterface(v["app"],false))
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("应用目前的实例数是"+strconv.Itoa(num))
	}

	//根据步长，计算出应用新的实例数
	instanceStep := utility.SwitchTypeToInt(v["instance_step"])
	//instanceStep,_ := strconv.Atoi(utility.SwitchNullInterface(v["instance_step"],false))
	fmt.Printf("instanceStep的值是%d\n",instanceStep)
	num = num +i * instanceStep
	//如果调整后的实例数超过了规则里的最大值，或者低于最小值，那么不调整实例数
	fmt.Printf("新的实例数是%d\n",num)
	if num < utility.SwitchTypeToInt(v["min_instance"]) {
		flag = false
		fmt.Printf("调整后的实例数低于规则允许的最小值，不会调整实例数\n")
	}
	if num > utility.SwitchTypeToInt(v["max_instance"]) {
		flag = false
		fmt.Printf("调整后的实例数大于规则允许的最大值，不会调整实例数\n")
	}

	if flag == true {
		fmt.Printf("即将调整实例数\n")
	} else {
		fmt.Printf("不调整实例数\n")
		return
	}
	//根据步长扩展应用的实例数或者缩短实例数
	err =utility.ScaleInstance(utility.SwitchNullInterface(v["app"],false),num)
	if err != nil {
		fmt.Println("扩展应用的实例数失败")
		fmt.Println(err.Error())
	} else {
		fmt.Println("扩展应用的实例数成功")
		models.SetWindowTime(ruleguid)
		_ = utility.SendAppInstanceChangeInfo(utility.SwitchNullInterface(v["app"],false),i,num)
	}
}


//停止正在运行的规则的协程
func (this * RuleController) StopRule() {
	ruleguid := this.Ctx.Input.Param(":ruleguid")
	models.AllRuleStruct[ruleguid].RuleChannel <- true
	utility.CurrentRuleNum -= 1
	this.Ctx.Output.SetStatus(200)
	this.Ctx.Output.Body([]byte("停止规则成功"))

}

func JsonFormat(appguid string, active bool) (json map[string]interface{}) {
		json = map[string]interface{} {
			"appguid": appguid,
			"active": active,
		}
	return json
}

func Sleep(){
	time.Sleep(30 * time.Second)
}
func PeriodicSendHeartBeat(){
	for{
		Sleep()
		SendHeartBeat()
	}
}

func SendHeartBeat(){
	localIp,port,apps := utility.GetLocalNodeIpPortMaxRule()
	//u :=ControllerHead()+"nodes/"+NodeGuid+"/"+localIp+"/"+port+"/"+strconv.Itoa(CurrentRuleNum)+"/"+apps
	////var jsonprep string = `"guid":"`+ NodeGuid +`"`
	//req, err := http.NewRequest("POST", u, strings.NewReader("name=wy"))
	//if err != nil {
	//	fmt.Printf("无法跟controller节点通信，controller节点可能挂了\n")
	//	return
	//}
	//req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	//client := &http.Client{}
	// resp, err := client.Do(req)
	// //defer resp.Body.Close()
	// if err != nil {
	//	fmt.Println("发送心跳到controller失败\n")
	//	 return
	// } else {
	//	fmt.Printf("发送心跳到controller成功\n")
	//	 resp.Body.Close()
	// }
	url := utility.ControllerHead()+"sendheartbeat/"
	jsonprep := map[string]interface{} {
		"nodeguid": utility.NodeGuid,
		"nodeip": localIp,
		"nodeport": port,
		"maxrules":apps,
	}
	str,err :=json.Marshal(jsonprep)
	if err !=nil {
		fmt.Println(err)
		return
	}

	var jsonStr = []byte(string(str))
	//从配置文件读取一个周期内，心跳最多能尝试的次数
	ctc := beego.AppConfig.String("checkTimeCycle")
	ctc1,_ := strconv.Atoi(ctc)

	req,err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	for i:=0; i<ctc1; i++ {
		resp,err := client.Do(req)
		//defer resp.Body.Close()
		if err == nil {
			resp.Body.Close()
			fmt.Println("发送心跳成功")
			return
		}
	}

	//心跳发送失败，检查是否所有规则没有停止，如果是，停止所有规则，如果没有，直接打印日志
	if utility.CurrentRuleNum == 0 {
		fmt.Println("节点向controller发送心跳失败")
		return
	}
	fmt.Println("节点向controller发送心跳失败，节点会停掉所有正在运行的规则")
	stopAllRules()
}


func stopAllRules() {
	for k,v := range models.AllRuleStruct {
		fmt.Printf("即将向管道发送停止信号，管道对应的规则guid是%s\n",k)
		v.RuleChannel <- true
		utility.CurrentRuleNum= utility.CurrentRuleNum-1
	}
	fmt.Println("停止所有运行的规则成功")
}

