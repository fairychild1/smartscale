#!/bin/bash
log_path=/var/log/

#清理掉指定日期前的日志
DAYS=3

#生成昨天的日志文件
cp ${log_path}smartscale.log ${log_path}smartscale_$(date -d "yesterday" +"%Y%m%d").log
echo > ${log_path}smartscale.log

##删除3天前的历史日志
find ${log_path} -name "smartscale*.log" -type f -mtime +3 -exec rm {} \;
