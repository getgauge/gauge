#!/bin/sh
PATH=$PATH:$GOPATH/bin protoc --go_out=gauge spec.proto
PATH=$PATH:$GOPATH/bin protoc --go_out=gauge messages.proto
PATH=$PATH:$GOPATH/bin protoc --go_out=gauge api.proto
protoc --java_out=gauge-java/src/main/java/ spec.proto
protoc --java_out=gauge-java/src/main/java/ messages.proto
protoc --java_out=gauge-java/src/main/java/ api.proto
ruby-protoc -o gauge-ruby/lib spec.proto
ruby-protoc -o gauge-ruby/lib messages.proto
ruby-protoc -o gauge-ruby/lib api.proto
