FROM golang:alpine

ENV SERVICE_NAME hrapp
ENV SERVICE_VERSION v1-0

COPY bin/${SERVICE_NAME} /opt/app/${SERVICE_NAME}/bin/
COPY bin/${SERVICE_NAME}-client /opt/app/${SERVICE_NAME}/bin/

COPY client/certs/127.0.0.1.crt /opt/app/${SERVICE_NAME}/client/certs/127.0.0.1.crt
COPY client/certs/127.0.0.1.key /opt/app/${SERVICE_NAME}/client/certs/127.0.0.1.key
COPY client/certs/root-ca.crt /opt/app/${SERVICE_NAME}/client/certs/root-ca.crt
COPY resource/dockerscript/client.sh .
RUN chmod 755 client.sh

COPY grpcserver/certs/mydomain.com.crt /opt/app/${SERVICE_NAME}/certs/mydomain.com.crt
COPY grpcserver/certs/mydomain.com.key /opt/app/${SERVICE_NAME}/certs/mydomain.com.key
COPY grpcserver/certs/root-ca.crt /opt/app/${SERVICE_NAME}/certs/root-ca.crt

EXPOSE 8080 8086