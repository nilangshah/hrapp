#!/bin/bash

go build -o hrapp cmd/main.go
go build -o hrappClient client/hrapp.go

./hrapp > /tmp/hrapp.log 2>&1 &
sleep 20
./hrappClient -pretty=True > /tmp/hrappClient.log 2>&1
echo "Hrapp logs at /tmp/hrapp.log and Client logs at /tmp/hrappClient.log, will stop hrapp server"
kill $(ps -ef | grep hrapp | grep -v "grep" | awk '{print $2;}') > /dev/null

