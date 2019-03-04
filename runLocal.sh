#!/bin/bash

go build -o hrapp cmd/main.go
go build -o hrappClient client/hrapp.go

./hrapp > /tmp/1.log &
sleep 15
./hrappClient

