/*创建数据库和用户*/
CREATE DATABASE IF NOT EXISTS smart_scale_instances DEFAULT CHARSET utf8 COLLATE utf8_general_ci;
grant all privileges on smart_scale_instances.* to 'smart_node'@'%' IDENTIFIED by '123456';
911602c1c755e01e7610a4b6692dd37e

/*创建规则表*/
CREATE TABLE rule_table (
rule_guid varchar(50),
rule_name varchar(20),
rule_type varchar(10),
year int,
day_week int,
month int,
day int,
hour int,
minute int,
min_cpu float,
max_cpu float,
min_mem float,
max_mem float,
min_req int default 0,
max_req int default 10000,
min_instance int default 0,
max_instance int,
instance_step int,
primary key (rule_guid)
);

/*创建规则和应用对应的表*/
create table app_table (
app_guid varchar(50),
active boolean,
rule_guid varchar(50),
node_guid varchar(50),
primary key (app_guid),
foreign key(rule_guid) references rule_table(rule_guid)
);

/*创建记录应用的变更操作日志的表*/
create table app_instance_change_table (
app_guid varchar(50),
change_info varchar(1000),
change_date date,
primary key (app_guid)
)

/*创建保存节点信息的表*/
create table node_table (
node_guid varchar(50),
node_ip varchar(20),
node_port int,
primary key (node_guid)
)