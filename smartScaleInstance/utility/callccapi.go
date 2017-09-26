package utility
import (
	"fmt"
	"github.com/astaxie/beego"
	"io/ioutil"
	"encoding/json"
)

//返回app的应用状态，true表示在运行，false表示没有运行
func GetAppStatus(appguid string) (bool,error) {
	domain :=beego.AppConfig.String("domain")
	username :=beego.AppConfig.String("username")
	password :=beego.AppConfig.String("password")
	if ApiStruct == nil {
		ApiStruct =NewApi(domain,username,password,false,false)
	}
	r, err := ApiStruct.Get(fmt.Sprintf("/v2/apps/%s",appguid))
	if err != nil {
		return false,err
	}

	body, _ := ioutil.ReadAll(r.Body)
	tempStru :=  map[string] interface{}{}
	entity := map[string] interface{}{}
	if err := json.Unmarshal(body, &tempStru); err != nil {
		fmt.Println(err)
		return false,err
	}

	for k,_ := range tempStru{
		if k == "error_code" {
			return false,fmt.Errorf("%s","无法获取应用的状态")
		}
	}


	//拿到entity字段的内容
	en,err :=json.Marshal(tempStru["entity"])     //将最外层json结果的某个字段（value为json）编译成json格式
	if err != nil {
		fmt.Println(err)
		return false,err
	}

	err = json.Unmarshal(en,&entity)      //将编译后的json格式内容反向解析回来
	if err != nil {
		fmt.Println(err)
		return false,err
	}

	status := entity["state"].(string)
	fmt.Printf("应用的状态的字符是%s\n",status)
	if status == "STOPPED" {
		return false,nil
	} else if status == "STARTED" {
		return true,nil
	} else {
		return false,fmt.Errorf("无法识别的应用状态码")
	}

}
