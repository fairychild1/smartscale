package models
import (
	"fmt"
	"time"
	"github.com/astaxie/beego"
	"strconv"
)

//定义规则对应的数据结构，每个规则在启动后，就会创建这样一个相关的数据结构，有一个对应的管道，当向这个管道传输一个true时，那么这个规则会停止。WindowFlag是窗口时间，当为true是，表示规则刚才生效过，对应的应用的实例已经调整过，现在是窗口时间，不需要再检查应用的状态，WindowPoint是调整规则对应的应用的实例数的时间点。LogSystemWindowFlag是查询日志系统的窗口时间，当为true时，表示刚刚查询过日志系统，现在是窗口时间，不需要再查询日志系统，LogSystemWindowPoint是查询日志系统的时间点。
type RuleStruct struct {
	RuleChannel chan bool
	WindowFlag bool
	WindowPoint time.Time
	LogSystemWindowFlag bool
	LogSystemWindowPoint time.Time
}

//定义保存所有管道的数据结构
var AllRuleStruct map[string] *RuleStruct

func init(){
	AllRuleStruct = make(map[string] *RuleStruct)
}

func NewRuleStruct(ruleguid string) {
	c := make(chan bool)
	s := &RuleStruct{
		RuleChannel: c,
		WindowFlag: false,
		WindowPoint: time.Now(),
		LogSystemWindowFlag: false,
		LogSystemWindowPoint: time.Now(),
	}
	//将管道保存到AllChannel里面
	AllRuleStruct[ruleguid] = s
}

//func VisitChannel(){
//	for k,_ := range AllChannel {
//		fmt.Printf("管道的key是%s\n",k)
//	}
//}

//判断是否是窗口时间，如果是，返回true，否则，返回false。当WindowsFlag为true时，检查当前时间与WindowPoint之间的时间差，如果时间差大于配置文件里面设置的windowTime,那么将WindowFlag设置为false，返回false。如果时间差小于windowTime，那么返回true.
func IfIsWindowTime(ruleguid string) bool{
	rulestruct := AllRuleStruct[ruleguid]
	windowFlag := rulestruct.WindowFlag
	windowPoint := rulestruct.WindowPoint
	windowtime := beego.AppConfig.String("windowTime")
	wt,_ := strconv.Atoi(windowtime)
	if windowFlag == true {
		timenow := time.Now()
		if int(timenow.Sub(windowPoint).Minutes()) >= wt {
			windowFlag = false
			rulestruct.WindowFlag = windowFlag
			return false
		} else {
			fmt.Printf("刚刚调整过应用的实例数，等窗口时间过了，会再次检查应用的负载\n")
			return true
		}
	} else {
		return false
	}
}

//判断是否是查询日志系统的窗口时间，如果是，返回true，如果不是，返回false。当LogSystemWindowFlag为true时，检查当前时间与LogSystemWindowPoint之间的时间差，如果时间差大于配置文件里面设置的logSystemWindowTime,那么将LogSystemWindowFlag设置为false，同时返回false,如果小于logSystemWindowTime，那么返回true.
func IfIsLogSystemWindowTime(ruleguid string) bool {
	rulestruct := AllRuleStruct[ruleguid]
	logSystemWindowFlag := rulestruct.LogSystemWindowFlag
	logSystemWindowPoint := rulestruct.LogSystemWindowPoint
	logSystemWindowTime := beego.AppConfig.String("logSystemWindowTime")
	lswt,_ := strconv.Atoi(logSystemWindowTime)
	if logSystemWindowFlag == true {
		timenow := time.Now()
		if int(timenow.Sub(logSystemWindowPoint).Minutes()) >= lswt {
			logSystemWindowFlag = false
			rulestruct.LogSystemWindowFlag = logSystemWindowFlag
			return false
		} else {
			fmt.Printf("刚刚查询过日志系统，等窗口时间过了，会再次通过日志系统查询规则%s对应的应用的负载\n",ruleguid)
			return true
		}
	} else {
		return false
	}
}

//设置窗口时间。当检查应用的负载，发现超过了阀值，再调用api调整完实例数后，需要设置窗口时间。
func SetWindowTime(ruleguid string)  {
	rulestruct := AllRuleStruct[ruleguid]
	rulestruct.WindowFlag = true
	rulestruct.WindowPoint = time.Now()
}

//设置查询日执行系统的窗口时间。当刚刚查询完日志系统，需要设置日志系统的窗口时间。
func SetLogSystemWindowTime(ruleguid string) {
	rulestruct := AllRuleStruct[ruleguid]
	rulestruct.LogSystemWindowFlag = true
	rulestruct.LogSystemWindowPoint = time.Now()
}


