#!/bin/bash

function stopSmartScale() 
{
	PID=`ps -ef|grep smartScaleInstance|grep -v grep|awk '{print $2}'`
	echo "pid is ${PID}"
	kill -9 ${PID}
	echo "smartScale is stopped"
}



function startSmartScale()
{
	ps -ef|grep smartScaleInstance|grep -v grep
	if [ $? -ne 0 ];then
		echo "smartScale is stopped"
	else
		echo "smartScale is running"
		stopSmartScale;
		
	fi

	rm smartScaleInstance
	go build

	echo "smartScale will be restart"
	nohup /usr/local/gopath/src/smartScaleInstance/smartScaleInstance >> /var/log/smartscale.log &
	if [ $? -eq 0 ];then
	echo "smartScale is restart"
	else
		echo "smartScale restart failed"
		exit 1
	fi	
}

if [ $# != 1 ] ; then 
	echo "USAGE: $0 stop|start" 
	echo " e.g.: run.sh stop" 
	exit 1; 
else
	if [ $1 = "stop" ];then
		ps -ef|grep smartScaleInstance|grep -v grep
		if [ $? -ne 0 ];then
			echo "smartScale is stopped"
			exit 0
		else
			stopSmartScale;
			exit 0
		fi
	fi
	if [ $1 = "start" ];then
		startSmartScale;
		exit 0
	fi
	echo "command is not supported, command should be stop|start"
	exit 1
fi


