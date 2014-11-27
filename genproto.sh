#!/bin/sh

#Using protoc version 2.5.0

cd gauge-proto
PATH=$PATH:$GOPATH/bin protoc --go_out=../ spec.proto
PATH=$PATH:$GOPATH/bin protoc --go_out=../ messages.proto
PATH=$PATH:$GOPATH/bin protoc --go_out=../ api.proto
cd ..
sed  -i.backup '/import main1 "spec.pb"/d' messages.pb.go && sed  -i.backup 's/main1.//g' messages.pb.go && rm messages.pb.go.backup
sed  -i.backup '/import main1 "spec.pb"/d' api.pb.go && sed  -i.backup 's/main1.//g' api.pb.go && rm api.pb.go.backup
go fmt

