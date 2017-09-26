package utility
import (
	"fmt"
	"net"
	"net/http"
	//"net/url"
	"os"
  	"bytes"
  	"io/ioutil"
	//"time"
	//"strings"
	"encoding/json"
	"github.com/astaxie/beego"
	//"strconv"
)

var NodeGuid string
var CurrentRuleNum int = 0

func RegisterSelfAsNode() {
	url :=ControllerHead()+"addnode"
	localIp,port,apps := GetLocalNodeIpPortMaxRule()
	var jsonprep string = `{"ip":"`+ localIp +`" , "port":`+port+`, "maxapps":`+apps+`}`
	var jsonStr = []byte(jsonprep)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		 fmt.Println("Unable to reach the server.")
	} else {
		 body, _ := ioutil.ReadAll(resp.Body)
		 tempStru := struct {Guid string}{}
		 if err := json.Unmarshal(body, &tempStru); err != nil {
			 fmt.Println("can not save the guid")
			 return
		 } else {
			 NodeGuid = tempStru.Guid
		 }
   }
}

//获取本地的ip，端口，和最大允许运行的规则数
func GetLocalNodeIpPortMaxRule() (string,string,string){
	maxRules :=beego.AppConfig.String("quotaOfAppRulesIneveryNode")
	localIp := GetLocalIp()
	port :=beego.AppConfig.String("httpport")
	return localIp,port,maxRules
}


func GetLocalIp() (ip string){
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		os.Stderr.WriteString("Oops:" + err.Error())
		os.Exit(1)
	}
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
        return ipnet.IP.String()
			}
		}
	}
	return "not found"
}

func ControllerHead() string{
	controllerip := beego.AppConfig.String("controllerip")
	controllerport :=beego.AppConfig.String("controllerport")
	return "http://"+controllerip+":"+controllerport+"/"
}

//将空接口转换成字符串。第二个参数如果为true，则将浮点数转换为整数，否则保留为浮点数
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
