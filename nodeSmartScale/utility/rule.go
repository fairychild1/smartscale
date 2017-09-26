package utility
import (
	"time"
	"fmt"
	"github.com/astaxie/beego"
	"net/http"
	"strings"
	"io/ioutil"
	"encoding/json"
	"strconv"
	//"reflect"
	"bytes"
	"errors"
)


//根据应用的guid，查询日志系统，获取应用的cpu平均使用情况，内存平均使用情况。返回的4个值为cpu平均使用情况，内存平均使用情况，查询是否为空，错误类型。
func CallLogCollectionSystem(appguid string) (cpuUsage float64,memUsage float64,flag bool,err error){
	domain :=beego.AppConfig.String("domain")
	t := beego.AppConfig.String("checkTimeCycle")
	t1,_ := strconv.Atoi(t)
	halfHoursAgo :=time.Now().Add(time.Duration(-t1) * time.Minute).Format("2006-01-02 15:04:05")
	timenow :=time.Now().Format("2006-01-02 15:04:05")
	url :="https://truepaas-cty-log."+domain+"/log/getAppsByAppIdAndTime/"+appguid+"/"+halfHoursAgo+"/"+timenow+"/10"
	req, err := http.NewRequest("GET", url, strings.NewReader("name=wy"))
	if err != nil {
		fmt.Printf("无法跟日志系统通信\n")
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := &http.Client{}
	resp, err := client.Do(req)
	//defer resp.Body.Close()
	if err != nil {
		fmt.Println("无法向日志系统发送查询应用压力的请求")
		return cpuUsage,memUsage,flag,err
	} else {
		var tempStru [] map [string]interface{}
		body, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err = json.Unmarshal(body, &tempStru); err != nil {
			fmt.Println("无法解析数组形式的json串")
			return cpuUsage,memUsage,flag,err
		} else {
			cpuUsage,memUsage,flag,err = ComputeSum(tempStru)
			if err != nil {
				fmt.Println("获取应用的负载使用情况失败")
				return cpuUsage,memUsage,flag,err
			} else {
				fmt.Println("获取应用的负载使用情况成功")
				return cpuUsage,memUsage,flag,err
			}
			//ComputeSum(tempStru,"cpu")
			//fmt.Println("获取cpu的总共的值成功")
			//fmt.Printf("cpu的平均使用是%.2f \n",cpuUsage/float32(len(tempStru)))

		}
	}
}

func ComputeSum(tempStru [] map [string]interface{}) (float64,float64,bool,error){
	var cpuUsage float64 = 0.0
	var memUsage float64 = 0.0
	var err1 error

	for _,v := range tempStru{
		for k1,v1 := range v{
			if k1 == "cpu_percentage" {
				cpuUsage,_ = AddNewElement(v1,cpuUsage)
			} else if k1 == "memory_bytes" {
				memUsage,_ = AddNewElement(v1,memUsage)
			}
		}
	}
	if len(tempStru) == 0 {
		return 0.0,0.0,true,err1
	} else {
		return cpuUsage/float64(len(tempStru)),memUsage/float64(len(tempStru)),false,err1
	}
}

func AddNewElement(v interface{},count float64) (float64,error) {
	value,err :=strconv.ParseFloat(SwitchNullInterface(v,true),64)   //将字符串转换为64位的浮点数
	if err !=nil {
		fmt.Println("无法转换string为float64")
	} else {
		float := float64(value)
		count = count + float
	}
	return count,err
}

func ScaleInstance(appGuid string,instanceStep int) error {
	domain :=beego.AppConfig.String("domain")
	username :=beego.AppConfig.String("username")
	password :=beego.AppConfig.String("password")
	if ApiStruct == nil {
		ApiStruct =NewApi(domain,username,password,false,false)
	}

	js := []byte(fmt.Sprintf("{\"instances\": %d}", instanceStep))
	b := bytes.NewBuffer(js)

	r, err := ApiStruct.Put(fmt.Sprintf("/v2/apps/%s", appGuid), b)
	if err != nil {
		return err
	}

	if r.StatusCode != 201 {
		return errors.New(fmt.Sprintf("Could not scale to %d. Status: %d\n", instanceStep, r.StatusCode))
	}

	r.Body.Close()
	return nil
}

//将应用的实例数变更的信息发送给controller，让controller存放到数据库里面
func SendAppInstanceChangeInfo(appguid string,i int,num int) error {
	controllerip := beego.AppConfig.String("controllerip")
	controllerport := beego.AppConfig.String("controllerport")
	url := "http://"+controllerip+":"+controllerport+"/addappinstancelog/"
	jsonprep := map[string]interface{} {
		"app_guid": appguid,
		"app_log" : fmt.Sprintf("%d",num),
	}
	str,err :=json.Marshal(jsonprep)
	if err != nil {
		fmt.Println(err)
		return err
	}
	var jsonStr = []byte(string(str))
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		fmt.Printf("无法跟controller节点通信\n")
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("无法向controller发送应用的实例数的变更信息\n")
		return err
	} else {
		resp.Body.Close()
		return nil
	}
}

func GetInstanceNum(appGuid string) (int,error) {
	domain :=beego.AppConfig.String("domain")
	username :=beego.AppConfig.String("username")
	password :=beego.AppConfig.String("password")
	if ApiStruct == nil {
		ApiStruct =NewApi(domain,username,password,false,false)
	}
	r, err := ApiStruct.Get(fmt.Sprintf("/v2/apps/%s/instances", appGuid))
	if err != nil {
		return 0,err
	}

	body, _ := ioutil.ReadAll(r.Body)
	tempStru :=  map[string] interface{}{}
	if err := json.Unmarshal(body, &tempStru); err != nil {
		fmt.Println(err)
		return 0,err
	} else{
		return len(tempStru),nil
	}
}

func GetMememoryQuota(appGuid string) (float64,error) {
	domain :=beego.AppConfig.String("domain")
	url :="https://message."+domain+"/message/getAppResourceQuota/"+appGuid+"/"
	req, err := http.NewRequest("GET", url, strings.NewReader("name=wy"))
	if err != nil {
		fmt.Printf("无法跟日志系统通信，从而获取内存的配额\n")
		return 0.0,err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := &http.Client{}
	resp, err := client.Do(req)
	//defer resp.Body.Close()
	if err != nil {
		fmt.Println("无法从日志系统获取应用的内存配额")
		return 0.0,err
	} else {
		var tempStru map [string]interface{}
		body, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()  //关闭连接，避免连接过多
		if err = json.Unmarshal(body, &tempStru); err != nil {
			fmt.Println("无法解析从日志系统获取的应用的内存配额的json串")
			return 0.0,err
		}
		mem,_ :=strconv.Atoi(SwitchNullInterface(tempStru["memory_quota"],false))
		fmt.Printf("内存配额是%d\n",mem)
		return float64(mem),err
	}
}

//根据应用的负载，规则的内容，判断是否改变应用的实例数.
//如果规则为cpu类型，判断是否大于规则的最大cpu阀值，或者小于规则的最小cpu阀值
//如果规则为mem类型，判断是否大于规则的最大mem阀值，或者小于规则的最小mem阀值
//如果规则为cpu,mem类型，当负载大于规则的最大cpu阀值或者大于规则的最大mem阀值时，那么就扩展。当同时小于规则的最小cpu阀值和最小mem阀值时，才缩减实例数。
//返回的参数为整数，增加为1，减少为-1,如果不变，则为0.第二个参数为是否调整实例数，true为调整，false为不调整.
func IfChangeAppInstance(cpuUsage float64,memUsage float64,v map[string]interface{}) (int,bool) {
	ruleType := SwitchTypeToString(v["rule_type"])
	if ruleType == "cpu" {
		if SwitchTypeToFloat64(v["min_cpu"]) > cpuUsage*100 {
			return -1,true
		}
		if SwitchTypeToFloat64(v["max_cpu"]) < cpuUsage*100 {
			return 1,true
		}
		return 0,false
	} else if ruleType == "mem" {
		if SwitchTypeToFloat64(v["min_mem"]) > memUsage*100 {
			return -1,true
		}
		if SwitchTypeToFloat64(v["max_mem"]) < memUsage*100 {
			return 1,true
		}
		return 0,false
	} else if ruleType == "cpu,mem" {
		if SwitchTypeToFloat64(v["min_cpu"]) > cpuUsage*100 && SwitchTypeToFloat64(v["min_mem"]) > memUsage*100 {
			return -1,true
		}
		if (SwitchTypeToFloat64(v["max_cpu"]) < cpuUsage*100) || (SwitchTypeToFloat64(v["max_mem"]) < memUsage*100) {
			return 1,true
		}
	}
	return 0,false
}

//将空接口转换成32位浮点数
func SwitchTypeToFloat32(temp interface{}) float32 {
	result,_:= strconv.ParseFloat(SwitchNullInterface(temp,true),32)
	return float32(result)
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

//根据服务实例的guid，获取服务的名字
func GetServiceNameByInstanceGuid(instanceGuid string) (string,error) {
	domain :=beego.AppConfig.String("domain")
	username :=beego.AppConfig.String("username")
	password :=beego.AppConfig.String("password")
	if ApiStruct == nil {
		ApiStruct =NewApi(domain,username,password,false,false)
	}

	r, err := ApiStruct.Get(fmt.Sprintf("/v2/service_instances/%s", instanceGuid))
	if err != nil {
		fmt.Println("根据实例的guid调用具体api失败")
		return "",err
	} else {
		body, _ := ioutil.ReadAll(r.Body)
		tempStru :=  map[string] interface{}{}
		entity := map[string] interface{}{}
		if err = json.Unmarshal(body, &tempStru); err != nil {
			fmt.Println(err)
			return "",err
		}

		en,err :=json.Marshal(tempStru["entity"])
		if err != nil {
			fmt.Println(err)
			return "",err
		}


		err = json.Unmarshal(en,&entity)
		if err != nil {
			fmt.Println(err)
			return "",err
		}

		fmt.Println("解析成功")
		return SwitchNullInterface(entity["name"],false),err
	}
}