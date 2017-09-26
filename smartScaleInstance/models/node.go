package models

import (
	"fmt"
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"io"
	"time"
)

var TotalNodes []*Node
const CHECKNODEACTIVETIME int =60

type Node struct {
	Guid    string  // Unique identifier
	LastActiveTime time.Time //上次活跃时间
	IsActive bool //当前是否活跃
	IsRestart bool //该节点是否刚刚重启
	MaxApps int //节点允许运行的最大规则数
	CurrentApps int //节点当前运行的规则数
	IsNeedMigrate bool //该节点上的规则是否需要迁移
	IsMigrating bool //该节点上的规则是否正在迁移
	Ip string // node节点的ip
	Port int //node节点用到的端口
}
func NewNode(ip string,port int,maxapps int) (*Node, error) {
	if ip == "" {
		return nil, fmt.Errorf("empty ip")
	}
	ipPortStr := fmt.Sprintf("%s%d",ip,port)
	for _,v := range TotalNodes {
		if ipPortStr == fmt.Sprintf("%s%d",v.Ip,v.Port) {
			v.IsRestart= true
			v.IsActive= true
			return v,nil
		}
	}
	tempguid :=GetGuid()
	t :=time.Now()
	return &Node{Guid: tempguid,LastActiveTime: t,IsActive: true, IsRestart: false, MaxApps: maxapps,CurrentApps: 0,IsNeedMigrate: false,IsMigrating: false,Ip: ip, Port : port}, nil
}

//根据节点的guid查找节点的ip和端口信息
func GetNodeIpAndPortByNodeGuid(nodeguid string) (string,int) {
	for _,v := range TotalNodes {
		if v.Guid == nodeguid {
			return v.Ip,v.Port
		}
	}
	fmt.Printf("找不到guid对应的节点\n")
	return "",0
}

//将节点运行的规则数减1
func NodeRunningRuleDelOne(nodeguid string){
	for _,v := range TotalNodes {
		if v.Guid == nodeguid {
			v.CurrentApps = v.CurrentApps -1
			return
		}
	}
}

//将节点运行的规则数加1
func NodeRunningRuleAddOne(nodeguid string){
	for _,v := range TotalNodes {
		if v.Guid == nodeguid {
			v.CurrentApps=v.CurrentApps+1
			return
		}
	}
}

//给规则选择一个运行节点，返回节点的guid，节点的ip，节点的端口，还有是否挑选节点成功的标志flag，总共周期是60秒，60秒内找到了，flag为true,没找到，flag为false
func ChooseNodeForRule()(string,string,int,bool){
	var flag bool=false
	var tempCurrentApps int=0
	for tempTime:=0;tempTime<CHECKNODEACTIVETIME;tempTime=tempTime+10 {
		flag,tempCurrentApps =checkArrayNull(flag,tempCurrentApps)
		if flag == false{
			time.Sleep(10 *time.Second)
			continue
		} else {
			nodeguid,nodeip,nodeport := ChooseNodeFromArray(tempCurrentApps)
			if nodeguid == "" {
				return nodeguid,nodeip,nodeport,false
			} else {
			return nodeguid,nodeip,nodeport,true
			}
		}
	}
	return "","",0,false
}

func ChooseNodeFromArray(tempCurrentApps int) (string,string,int){
	var temp int=tempCurrentApps
	var tempGuid,tempIp string
	var tempPort int
	for _,value := range TotalNodes {
		if value.IsActive == false {
			continue
		}
		if value.CurrentApps<=temp {
			temp=value.CurrentApps
			tempGuid=value.Guid
			tempIp=value.Ip
			tempPort=value.Port
		}
	}
	return tempGuid,tempIp,tempPort
}

//检查保存节点信息的内存块队列是否为空，如果是，返回false，如果非空，则判断这个队列是否有一个节点是活跃状态，如果有，返回这个节点运行的规则数，和true，如果没有，返回false。
func checkArrayNull(flag bool,tempCurrentApps int) (bool,int) {
	if len(TotalNodes) >=1 {
		for _,v := range TotalNodes{
			if v.IsActive == true {
				flag=true
				tempCurrentApps=v.CurrentApps
				return flag,tempCurrentApps
			}
		}
		return false,0
	}else {
		return false,0
	}
}

//将节点添加到节点池中，如果节点池已经存在该节点信息，返回true，如果不存在，返回false
func AddNodeToPool(node *Node) bool{
	for _,v := range TotalNodes {
		if node.Guid == v.Guid {
			return true
		}
	}
	TotalNodes=append(TotalNodes,node)
	return false
}

func Save(node *Node){
	for _,v := range TotalNodes {
		if node.Guid == v.Guid {
			return
		}
	}
	//fmt.Printf("添加的节点ip是%s 端口是%d\n",node.Ip,node.Port)
	TotalNodes=append(TotalNodes,node)
}

//生成32位md5字串
func GetMd5String(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}
//生成Guid字串
func GetGuid() string {
	b := make([]byte, 48)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}
	return GetMd5String(base64.URLEncoding.EncodeToString(b))
}

