PATH=$PATH:$GOPATH/bin protoc --go_out=src/ messages.proto
PATH=$PATH:$GOPATH/bin protoc --go_out=src/ api.proto
protoc --java_out=../gauge-java/src/main/java/ messages.proto 
protoc --java_out=../gauge-java/src/main/java/ api.proto 
#ruby-protoc -o ruby messages.proto 
