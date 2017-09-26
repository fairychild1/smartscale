#!/bin/bash
function stopnodeSmartScale() 
{
	PID=`ps -ef|grep nodeSmartScale|grep -v grep|awk '{print $2}'`
	echo "pid is ${PID}"
	kill -9 ${PID}
	echo "nodeSmartScale is stopped"
}



function startnodeSmartScale()
{
	ps -ef|grep nodeSmartScale|grep -v grep
	if [ $? -ne 0 ];then
		echo "nodeSmartScale is stopped"
	else
		echo "nodeSmartScale is running"
		stopnodeSmartScale;
		
	fi

	rm nodeSmartScale
	go build

	echo "nodeSmartScale will be restart"
	nohup /usr/local/gopath/src/nodeSmartScale/nodeSmartScale >> /var/log/smartscale.log &
	if [ $? -eq 0 ];then
	echo "nodeSmartScale is restart"
	else
		echo "nodeSmartScale restart failed"
	fi
}

if [ $# != 1 ] ; then 
	echo "USAGE: $0 stop|start" 
	echo " e.g.: run.sh stop" 
	exit 1; 
else
	if [ $1 = "stop" ];then
		ps -ef|grep nodeSmartScale|grep -v grep
		if [ $? -ne 0 ];then
			echo "nodeSmartScale is stopped"
			exit 0
		else
			stopnodeSmartScale;
			exit 0
		fi
	fi
	if [ $1 = "start" ];then
		startnodeSmartScale;
		exit 0
	fi
	echo "command is not supported, command should be stop|start"
	exit 1
fi


