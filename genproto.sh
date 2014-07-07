#!/bin/sh
PATH=$PATH:$GOPATH/bin protoc --go_out=gauge spec.proto
PATH=$PATH:$GOPATH/bin protoc --go_out=gauge messages.proto
sed  -i.backup '/import main1 "spec.pb"/d' gauge/messages.pb.go && sed  -i.backup 's/main1.//g' gauge/messages.pb.go && rm gauge/messages.pb.go.backup
PATH=$PATH:$GOPATH/bin protoc --go_out=gauge api.proto
sed  -i.backup '/import main1 "spec.pb"/d' gauge/api.pb.go && sed  -i.backup 's/main1.//g' gauge/api.pb.go && rm gauge/api.pb.go.backup
cd gauge && go fmt && cd ..

protoc --java_out=gauge-java/src/main/java/ spec.proto
protoc --java_out=gauge-java/src/main/java/ messages.proto
protoc --java_out=gauge-java/src/main/java/ api.proto

ruby-protoc -o gauge-ruby/lib spec.proto
ruby-protoc -o gauge-ruby/lib messages.proto
ruby-protoc -o gauge-ruby/lib api.proto
