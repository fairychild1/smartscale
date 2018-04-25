package utility
import (
	"smartScaleInstance/models"
	"github.com/astaxie/beego"
	"time"
	//"reflect"
	"fmt"
	"strconv"
	"strings"
	"net/http"
	"encoding/json"
	"bytes"
	//"io/ioutil"

)
func Sleep(i int){
	time.Sleep(time.Duration(i) * time.Second)
}

func RefreshRuleInDeadNode(){
	i := beego.AppConfig.String("checknodecycle")
	c := beego.AppConfig.String("checknodetimes")  //检查的次数，如果checknodecycle * checknodetimes,节点还是没有恢复正常，那么就将规则放到其他节点运行
	cycle,_:=strconv.Atoi(i)
	times,_ :=strconv.Atoi(c)
	//d :=time.Now().Sub(t1).Seconds()
	//fmt.Println("type is :",reflect.TypeOf(d))
	//fmt.Printf("时间差是 %d\n",int(d))
	for {
		Sleep(cycle)
		for _,v :=range models.TotalNodes {
			if IsNodeActive(v,cycle,times) == true {
				if v.IsRestart == true {
					recoverNode(v.Guid,v.Ip,v.Port)
					v.IsRestart = false

				} else {
					fmt.Printf("node is %s , status is %t \n",v.Guid,v.IsActive)
				}
			} else {
				//将死掉的节点的is_active状态设置为false,将运行的规则数设置为0
				RefreshNodeStatus(v)

				//判断节点上是否还有规则，如果有，就迁移，如果没有，就跳过
				if v.CurrentApps == 0 {
					continue
				}

				fmt.Printf("节点 %s 挂了,即将切换这个节点上所运行的规则到另外一个节点\n",v.Guid)
				fmt.Printf("该节点的状态是%t 该节点上运行的规则数是%d\n",v.IsActive,v.CurrentApps)


				//查找这个节点运行的所有规则
				rulesingivennodepoint,err:=models.HuntForAllRulesInGivenNode(v.Guid)
				if err != nil {
					fmt.Printf("查找在指定节点 %s 上运行的规则失败\n",v.Guid)
					continue
				}


				for _, m := range *rulesingivennodepoint {
					//对于每条规则，先找一个合适的节点
					//2017-7-17此处逻辑需要重新修改，appGuid,isActiveString,ruleGuid三个的值需要重新运算
					ruleGuid := SwitchNullInterface(m["RuleGuid"],false)
					app,err := models.GetAppByRuleGuid(ruleGuid)
					if err != nil {
						fmt.Printf("在规则应用关系表里找不到规则%s对应的记录，请确认该规则的guid有效",ruleGuid)
						continue
					}
					appstatus,err := GetAppStatus(app.AppGuid)

					//2017-7-17此处逻辑需要重新修改，appGuid,isActiveString,ruleGuid三个的值需要重新运算

					fmt.Printf("数据库里获取到的规则%s对应的appguid的值是%s\n",ruleGuid,app.AppGuid)
					nodeguid,nodeip,nodeport,flag :=models.ChooseNodeForRule()
					if flag == false {
						fmt.Printf("找不到可用的节点来运行规则 %s \n",ruleGuid)
						continue
					}
					//根据规则的guid找到该规则的全部内容
					fmt.Printf("即将切换如下规则到新的节点:%s ,ip 是 %s 端口是 %d \n",nodeguid,nodeip,nodeport)
					rule,err :=models.GetRule(ruleGuid)
					if err != nil {
						fmt.Printf("找不到guid %s 对应的规则\n",ruleGuid)
						continue
					}

					//将规则发送到新的节点
					err =SendRuleToNode(rule,app.AppGuid,appstatus,nodeip,nodeport)
					if err != nil {
						fmt.Printf("将规则发送到节点  ip是%s 端口是%d 失败，请检查该节点的网络",nodeip,nodeport)
						continue
					} else {
						//规则发送到新的节点成功，将节点的运行规则数+1，更新数据库app表里这个规则和原来死掉的节点对应数据为规则和新的节点
						models.NodeRunningRuleAddOne(nodeguid)

						//节点池中，当前挂掉的节点运行的规则数减1
						v.CurrentApps= v.CurrentApps -1

						err =models.UpdateRuleOnNode(ruleGuid,nodeguid)
						if err != nil {
							fmt.Printf("更新死掉的节点上运行的规则所在的节点为新节点失败，稍后重试")
							continue
						}
					}

				}

				//将每条规则平均发布到其他节点：对于每条规则，先找一个合适的节点，将节点的运行规则数+1，然后发送规则到该节点，然后更新数据库app表里这个规则和原来死掉的节点对应数据为规则和新的节点
				//nodeguid,nodeip,nodeport,flag:=models.ChooseNodeForRule()
				//if flag == false {
				//	fmt.Printf("没有找到可用的新节点，来运行规则")
				//}


			}
		}
	}
}

func SendRuleToNode(rt *models.RuleTable,appguid string,active bool,ip string,port int) error{
	url := "http://"+ip+":"+strconv.Itoa(port)+"/runrule/"
	jsonprep := map[string]interface{} {
		"app": appguid,
		"rule_guid": rt.Guid,
		"active": active,
		"rule_type":rt.RuleType,
		"year":rt.Year,
		"day_week" : rt.DayWeek,
		"month" : rt.Month,
		"day" : rt.Day,
		"hour" : rt.Hour,
		"minute" : rt.Minute,
		"min_cpu" : rt.MinCpu,
		"max_cpu" : rt.MaxCpu,
		"min_mem" : rt.MinMem,
		"max_mem" : rt.MaxMem,
		"min_instance" : rt.MinInstance,
		"max_instance" : rt.MaxInstance,
		"instance_step" : rt.InstanceStep,
		"min_req" : rt.MinReq,
		"max_req" : rt.MaxReq,
	}
	str,err :=json.Marshal(jsonprep)
	if err != nil {
		fmt.Println(err)
		return err
	}
	var jsonStr = []byte(string(str))
	req,_ := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp,err := client.Do(req)
	//defer resp.Body.Close()
	if err != nil {
		fmt.Println("Unable to post the rule to node.")
		return err
	} else {
		resp.Body.Close()
		return nil
	}
}

//判断节点是否处于活跃状态，如果是，返回true，如果不是，返回false
func IsNodeActive(node *models.Node,cycle int,times int) bool {
	if node.IsActive == false {
		return false
	}
	//如果当前时间跟节点最后一次的活跃时间的差值大于心跳周期的times倍，那么该节点就死了
	if int(time.Now().Sub(node.LastActiveTime).Seconds()) >cycle * times {
		return false
	}
	return true
}

func RefreshNodeStatus(node *models.Node){
	node.IsActive=false
}

//向node发送停止运行某个规则的请求
func StopRule(ruleguid string,nodeip string,nodeport int) error {
	u := "http://"+nodeip+":"+fmt.Sprintf("%d",nodeport)+"/stoprule/"+ruleguid
	req, _ := http.NewRequest("POST", u, strings.NewReader("name=wy"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := &http.Client{}
	resp, err := client.Do(req)
	//defer resp.Body.Close()
	if err != nil {
		fmt.Println("发送停止运行规则的请求到节点失败")
	} else {
		fmt.Println("发送停止运行规则的请求到节点成功\n")
		resp.Body.Close()
	}
	return err
}

//将空接口转换为64为浮点数
func SwitchTypeToFloat64(temp interface{}) float64 {
	result,_:= strconv.ParseFloat(SwitchNullInterface(temp,true),64)
	return float64(result)
}

//将空接口转换为字符串
func SwitchTypeToString(temp interface{}) string {
	return SwitchNullInterface(temp,false)
}

//将空接口转换为整数
func SwitchTypeToInt(temp interface{}) int {
	result,_ := strconv.Atoi(SwitchNullInterface(temp,true))
	return result
}


//节点重启了，将该节点本来运行的规则重新发送到节点上
func recoverNode(nodeguid string,nodeip string,nodeport int) {
	rules := models.GetRulesInNode(nodeguid)
	for _,v := range rules {
		rule,_ := models.GetRule(v.RuleGuid)
		app,_:= models.GetAppByRuleGuid(v.RuleGuid)
		appguid := app.AppGuid
		SendRuleToNode(rule,appguid,true,nodeip,nodeport)
	}
}



//将空接口类型转换成string，如果空接口是浮点数，可以要求是否将浮点数转换成整数，然后再转换成string
func SwitchNullInterface(i interface{},floattoint bool) string {
	switch v := i.(type) {
	case string:
		return fmt.Sprintf("%s",v)
	case int:
		return fmt.Sprintf("%d", v)
	case float32:
		if floattoint == true{
			return fmt.Sprintf("%d", int(v))
		}else {
			return fmt.Sprintf("%.2f", v)
		}
	case float64:
		if floattoint == true{
			return fmt.Sprintf("%d", int(v))
		}else {
			return fmt.Sprintf("%.2f", v)
		}
	case bool:
		return fmt.Sprintf("%t", v)
	}
	return "not supported type"
}