package models

import(
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
	"fmt"
	"time"
)
func init() {
	// set default database
	orm.RegisterDataBase("default", "mysql", "smart_node:123456@/smart_scale_instances?charset=utf8", 30)

	// register model
	orm.RegisterModel(new(User),new(NodeTable),new(RuleTable),new(RuleNodeTable),new(AppRuleTable),new(AppInstanceChangeTable))

	// create table
	orm.RunSyncdb("default", false, true)
}



type User struct {
	Id   int
	Name string `orm:"size(100)"`
}

type NodeTable struct {
	NodeGuid string `orm:"pk;size(50)"`
	NodeIp string `orm:"size(20)"`
	NodePort int
	MaxApps int
}

//根据的guid查找这个app的所有instance的变更记录
func GetAllNodesFromDB() ([]*NodeTable,error) {
	o := orm.NewOrm()
	var nodes []*NodeTable
	_,err := o.QueryTable("node_table").All(&nodes)
	return nodes,err
}

//添加节点
func AddNodeInDB(newnode *NodeTable)(err error) {
	o := orm.NewOrm()
	_,err = o.Insert(newnode)
	if err !=nil {
		fmt.Println("添加节点信息到数据库失败")
		return err
	}
	return nil
}

type RuleTable struct {
	Guid string `orm:"pk;size(50)"`
	RuleName string `orm:"size(20)"`
	RuleType string `orm:"size(10)"`
	RuleActive bool `orm:"default(false)"`
	MinInstance int `orm:"default(1)"`
	MaxInstance int
	InstanceStep int `orm:"default(1)"`
	Year int
	DayWeek int
	Month int
	Day int
	Hour int
	Minute int
	MinCpu float64
	MaxCpu float64
	MinMem float64
	MaxMem float64
	MinReq int `orm:"default(0)"`
	MaxReq int `orm:"default(10000)"`
}

//添加规则
func AddRule(newrule *RuleTable)(err error) {
	o := orm.NewOrm()
	_,err = o.Insert(newrule)
	if err !=nil {
		fmt.Println("insert rule error")
		return err
	}
	return nil
}

//更新规则
func UpdateRule(rule *RuleTable) error {
	o := orm.NewOrm()
	_, err := o.Update(rule)
	if err == nil {
		fmt.Printf("更新规则成功\n")
		return err
	} else {
		return err
	}
}

//删除规则
func DelRule(ruleguid string) error{
	o := orm.NewOrm()
	num, err := o.Delete(&RuleTable{Guid: ruleguid})
	if err == nil {
		fmt.Printf("删除了%d条规则的记录\n",num)
	}
	return err
}

//定义规则与应用的对应关系的表
type AppRuleTable struct {
	RuleGuid string `orm:"pk;size(50)"`
	AppGuid string `orm:"size(50)"`
}

//添加规则和应用的关联关系，到数据库
func AddRuleApp(item *AppRuleTable) error {
	o := orm.NewOrm()
	_,err := o.Insert(item)
	if err != nil {
		fmt.Printf("添加规则与应用的关联关系到数据库失败\n")
		return err
	}
	fmt.Printf("添加规则与应用的关联关系到数据库成功\n")
	return nil
}

//删除规则与应用的关联记录
func DelRuleAppRelation(ruleguid string) error {
	o := orm.NewOrm()
	_, err := o.Delete(&AppRuleTable{RuleGuid: ruleguid})
	if err == nil {
		fmt.Printf("删除规则与应用的关联记录成功\n")
	}
	return err
}

//根据规则的guid，查找规则绑定的应用
func GetAppByRuleGuid(ruleguid string) (*AppRuleTable,error){
	o :=orm.NewOrm()
	app :=AppRuleTable{RuleGuid : ruleguid}
	err := o.Read(&app)
	if err == orm.ErrNoRows {
		fmt.Printf("在规则应用关系表里面，查询不到规则的guid%s对应的记录\n",ruleguid)
		return &app,err
	} else if err == orm.ErrMissPK {
		fmt.Println("找不到规则的主键")
		return &app,err
	} else {
		fmt.Printf("在规则应用关系表里找到了记录，应用的guid是%s\n",app.RuleGuid)
		return &app,err
	}
}


//定义应用的实例数修改记录的表，将实例的修改记录存放到数据库里面
type AppInstanceChangeTable struct {
	Id int
	AppGuid string `orm:"size(50)"`
	ChangeInfo string `orm:"size(1000)"`
	ChangeDate time.Time
}

//根据app的guid查找这个app的所有instance的变更记录
func GetAppInstanceChangeLogFromDB(appguid string) (*[]orm.Params,error) {
	o :=orm.NewOrm()
	var instancelog []orm.Params
	_, err := o.QueryTable("app_instance_change_table").Filter("app_guid", appguid).Values(&instancelog)
	return &instancelog,err
}

//添加app的实例修改记录
func AddAppInstanceChangeRecord(item *AppInstanceChangeTable) error{
	o := orm.NewOrm()
	_,err := o.Insert(item)
	if err != nil {
		fmt.Printf("应用%s的实例数发生变更，添加应用的实例数变更记录失败\n",item.AppGuid)
		return err
	}
	fmt.Printf("应用%s的实例数发生变更，添加应用的实例数变更记录成功\n",item.AppGuid)
	return nil
}

//定义规则与节点的对应关系的表
type RuleNodeTable struct {
	RuleGuid string `orm:"pk;size(50)"`
	NodeGuid string `orm:"size(50)"`
}


func GetRulesInNode(nodeguid string) []*RuleNodeTable{
	o :=orm.NewOrm()
	var rules []*RuleNodeTable
	_, _ = o.QueryTable("rule_node_table").Filter("node_guid", nodeguid).All(&rules)
	return rules
}

func AddNodeRuleRelation(item *RuleNodeTable)(err error){
	o := orm.NewOrm()
	_,err = o.Insert(item)
	if err !=nil {
		fmt.Println("添加规则与节点的对应关系到数据库失败")
		return err
	}
	return nil
}

func DelNodeRuleRelation(ruleguid string) error {
	o := orm.NewOrm()
	num, err := o.Delete(&RuleNodeTable{RuleGuid: ruleguid})
	if err == nil {
		fmt.Printf("删除了规则节点关系表中节点、规则的关系记录%d条\n",num)
	}
	return err
}

//根据rule的guid，查找这个rule运行在哪个node上面,返回节点的guid
func GetNodeByRuleGuid(ruleguid string) string{
	o :=orm.NewOrm()
	relation :=RuleNodeTable{RuleGuid : ruleguid}
	err := o.Read(&relation)
	if err == orm.ErrNoRows {
		fmt.Println("查询不到规则所在的节点关系")
		return ""
	} else if err == orm.ErrMissPK {
		fmt.Println("找不到关系表的主键")
		return ""
	} else {
		return relation.NodeGuid
	}
}

//根据app的guid，查询这个app上对应的所有规则
func GetAllRulesByAppGuid(appguid string)(*[]orm.Params,error){
	o :=orm.NewOrm()
	var rulesingivenapp []orm.Params
	_, err := o.QueryTable("app_rule_table").Filter("app_guid", appguid).Values(&rulesingivenapp)
	return &rulesingivenapp,err
}

//根据rule的guid查找这条规则
func GetRule(ruleguid string)(*RuleTable,error){
	o :=orm.NewOrm()
	rule :=RuleTable{Guid : ruleguid}
	err := o.Read(&rule)
	if err == orm.ErrNoRows {
		fmt.Println("查询不到规则所在的表")
		return &rule,err
	} else if err == orm.ErrMissPK {
		fmt.Println("找不到规则的主键")
		return &rule,err
	} else {
		fmt.Println(rule.Guid, rule.RuleName, rule.RuleType)
		return &rule,err
	}
}

//修改rule表中rule的状态字段rule_active
func UpdateRuleStatus(ruleguid string,status bool) error {
	o := orm.NewOrm()
	rule := RuleTable{Guid: ruleguid}
	if err :=o.Read(&rule); err ==nil {
		rule.RuleActive = status
		_, err := o.Update(&rule)
		if err == nil {
			fmt.Printf("更新规则%s的状态为%t成功\n",ruleguid,status)
			return err
		} else {
			fmt.Printf("更新规则%s的状态为%t失败\n",ruleguid,status)
			return err
		}

	}else {
		return err
	}
}

//更新规则所在的节点.这是数据库的更新操作。
func UpdateRuleOnNode(ruleguid string,newnodeguid string) error{
	o := orm.NewOrm()
	rulenode := RuleNodeTable{RuleGuid: ruleguid}
	if err :=o.Read(&rulenode); err ==nil {
		rulenode.NodeGuid = newnodeguid
		_, err := o.Update(&rulenode)
		if err == nil {
			fmt.Printf("更新运行规则的节点为新节点成功\n")
			return err
		} else {
			return err
		}

	}else {
		return err
	}
}

//查找在某个节点上运行的所有规则
func HuntForAllRulesInGivenNode(nodeguid string) (*[]orm.Params,error){
	o :=orm.NewOrm()
	var rulesingivennode []orm.Params
	_, err := o.QueryTable("rule_node_table").Filter("node_guid", nodeguid).Values(&rulesingivennode)
	return &rulesingivennode,err

}

func Test_db() {
	o := orm.NewOrm()

	user := User{Name: "slene"}

	// insert
	id, err := o.Insert(&user)
	fmt.Printf("ID: %d, ERR: %v\n", id, err)

	// update
	user.Name = "astaxie"
	num, err := o.Update(&user)
	fmt.Printf("NUM: %d, ERR: %v\n", num, err)

	// read one
	u := User{Id: user.Id}
	err = o.Read(&u)
	fmt.Printf("ERR: %v\n", err)

	// delete
	num, err = o.Delete(&u)
	fmt.Printf("NUM: %d, ERR: %v\n", num, err)
}
