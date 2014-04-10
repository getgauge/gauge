PATH=$PATH:$GOPATH/bin protoc --go_out=. messages.proto
protoc --java_out=java/src messages.proto 
ruby-protoc -o ruby messages.proto 