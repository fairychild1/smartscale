#说明
- 本项目用go编写，能通过给cloudfoundry的应用设置规则（cpu，内存，周期性，访问量等），来自动扩展或缩小应用的实例。
- 项目分为controler端和node端，controller自带数据库，记录规则，节点，应用等信息，并负责分发规则到各node。node运行具体的规则，并周期性的向controller发送心跳，如果node通过规则将应用的实例变动了，那么node还会将应用的实例变更信息发送给controller。
- controller有一个node池，运行在内存，保存着node以及node上的规则数，node是否活跃，上次活跃时间等信息，controller本身是无状态的，在controller重启后，node池的信息会自动恢复。
- node可以横向扩展，在node重启后，controller会重新将该node运行的规则发送给node。如果node挂掉，controller会将该node上运行的规则调度到其他node上运行。

#API说明
##获取所有的node信息
`get http://controllerip:port/nodes/`

##获取某个应用对应的规则
`get http://10.89.128.18:8080/getrulesforapp/[app-guid]`

##添加规则
```
post http://10.89.128.18:8080/addrule/
{
  "rule_name" : "tank-wy的规则",
  "rule_type" : "mem",
  "app_guid" : "5517e2e1-e8c7-4100-bd57-7f37791c9e38",
  "app_active" : true,
  "min_instance" : 1,
  "max_instance" : 6,
  "instance_step" :1,
  "year" : 0,
  "day_week" : 0,
  "month" : 0,
  "day" : 0,
  "hour" : 0,
  "minute" : 0,
  "min_cpu" : 10.0,
  "max_cpu" : 80.0,
  "min_mem" : 10,
  "max_mem" : 30,
  "min_req" : 0,
  "max_req" : 10000
}
```

##更新规则
```
post http://10.89.128.18:8080/updaterule/
{
	"rule_guid" : "eb5639ff4d7033f23aa4ee49c5e96963",
	"rule_name" : "cpu小于65%",
	"rule_type" : "mem",
	"app_active" : true,
	"min_instance" : 1,
	"max_instance" : 6,
	"instance_step" :1,
	"year" : 0,
	"day_week" : 0,
	"month" : 0,
	"day" : 0,
	"hour" : 0,
	"minute" : 0,
	"min_cpu" : 10.0,
	"max_cpu" : 80.0,
	"min_mem" : 1.0,
	"max_mem" : 80.0,
	"min_req" : 0,
	"max_req" : 10000
}
```

##删除规则
`delete http://10.89.128.18:8080/deleterule/30e2cac6853ca7385bd26157fc7277db`

##启动规则
`post http://10.89.128.18:8080/startrule/af3960069160e0cb77e85868d40716b8`

##停止规则
`post http://10.89.128.18:8080/stoprule/8d4a47af1b1c0c0e472347191b764043`

##获取uaa的token
```
post http://uaa.truepaas.com/oauth/token?username=wangying@chutianyun.gov.cn&password=123456&grant_type=password
[{"key":"Content-Type","value":"application/x-www-form-urlencoded","description":""},{"key":"Authorization","value":"Basic Y2Y6","description":""},{"key":"Accept","value":"application/json","description":""}]

```

#原理图说明
![](https://i.imgur.com/GjcfEXa.png)
说明：
1.	node在启动后会自动向controller节点注册自己，controller节点会将node的信息保存在内存的一块区域，具体为node池。
2.	node负责最终运行规则。
3.	前端web发送规则，controller负责接收规则，接收到规则后通过scheduller进行调度；
4.	controller的scheduller会从node池里面挑选一个处于活跃状态，并且运行的规则数最少的node，作为运行的node，然后将规则发送给node；
5.	node在接收到规则后，启动一个协程，运行该规则；
6.	前端web发送停止某个规则的请求，controller收到消息后，从数据库里面查询该规则对应的节点，然后向节点发送停止运行该规则的请求。node收到controller发来的请求后，停止运行该规则的协程。
