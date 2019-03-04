#!/bin/sh
cd /opt/app/hrapp/bin
./hrapp-client -certpath=/opt/app/hrapp/client/certs/127.0.0.1.crt -keypath=/opt/app/hrapp/client/certs/127.0.0.1.key -capath=/opt/app/hrapp/client/certs/root-ca.crt -pretty=True