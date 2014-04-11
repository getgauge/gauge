PATH=$PATH:$GOPATH/bin protoc --go_out=src/ messages.proto
protoc --java_out=../twist2-java/src/main/java/ messages.proto 
#ruby-protoc -o ruby messages.proto 