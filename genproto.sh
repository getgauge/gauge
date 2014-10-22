#!/bin/sh

#Using protoc version 2.5.0

cd gauge-proto
PATH=$PATH:$GOPATH/bin protoc --go_out=../gauge spec.proto
PATH=$PATH:$GOPATH/bin protoc --go_out=../gauge messages.proto
PATH=$PATH:$GOPATH/bin protoc --go_out=../gauge api.proto
cd ..
sed  -i.backup '/import main1 "spec.pb"/d' gauge/messages.pb.go && sed  -i.backup 's/main1.//g' gauge/messages.pb.go && rm gauge/messages.pb.go.backup
sed  -i.backup '/import main1 "spec.pb"/d' gauge/api.pb.go && sed  -i.backup 's/main1.//g' gauge/api.pb.go && rm gauge/api.pb.go.backup
cd gauge && go fmt && cd ..

