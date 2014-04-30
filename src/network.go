package main

import (
	"bytes"
	"code.google.com/p/goprotobuf/proto"
	"errors"
	"fmt"
	"github.com/twist2/common"
	"io"
	"log"
	"net"
	"time"
)

const (
	runnerConnectionPort    = ":8888"
	runnerConnectionTimeOut = time.Second * 5
)

// MessageId -> Callback
var pendingRequests = make(map[int64]chan<- *Message)

func handleConnection(conn net.Conn) {
	buffer := new(bytes.Buffer)
	data := make([]byte, 8192)
	for {
		n, err := conn.Read(data)
		if err != nil {
			if err == io.EOF {
				return
			}
			log.Println(err.Error())
		}

		buffer.Write(data[0:n])

		messageLength, bytesRead := proto.DecodeVarint(buffer.Bytes())
		if messageLength > 0 && messageLength < uint64(buffer.Len()) {
			message := &Message{}
			err = proto.Unmarshal(buffer.Bytes()[bytesRead:messageLength+uint64(bytesRead)], message)
			if err != nil {
				log.Printf("Failed to read proto message: %s\n", err.Error())
			} else {
				responseChannel := pendingRequests[*message.MessageId]
				responseChannel <- message
				delete(pendingRequests, *message.MessageId)
				buffer.Reset()
			}
		}
	}
}

func acceptConnection(portNo string, connectionTimeOut time.Duration) (net.Conn, error) {
	listener, err := net.Listen("tcp", portNo)
	if err != nil {
		return nil, err
	}
	errChannel := make(chan error, 1)
	connectionChannel := make(chan net.Conn, 1)

	go func() {
		connection, err := listener.Accept()
		errChannel <- err
		connectionChannel <- connection

	}()

	select {
	case err := <-errChannel:
		return nil, err
	case conn := <-connectionChannel:
		go handleConnection(conn)
		return conn, nil
	case <-time.After(connectionTimeOut):
		return nil, errors.New(fmt.Sprintf("Timed out connecting to port %s", portNo))
	}

}

// Sends the specified message and waits for a response
// This function blocks till it gets a response
// Each message gets a unique id and messages are prefixed with it's length
// encoded using protobuf'd varint format
func getResponse(conn net.Conn, message *Message) (*Message, error) {
	responseChan := make(chan *Message)
	messageId := common.GetUniqueId()
	message.MessageId = &messageId
	pendingRequests[*message.MessageId] = responseChan

	data, err := proto.Marshal(message)
	if err != nil {
		delete(pendingRequests, *message.MessageId)
		return nil, err
	}
	dataLength := proto.EncodeVarint(uint64(len(data)))
	data = append(dataLength, data...)

	_, err = conn.Write(data)
	if err != nil {
		delete(pendingRequests, *message.MessageId)
		return nil, err
	}

	select {
	case response := <-responseChan:
		return response, nil
	}
}

//Sends a specified message and does not wait for any response
func writeMessage(conn net.Conn, message *Message) error {
	messageId := common.GetUniqueId()
	message.MessageId = &messageId

	data, err := proto.Marshal(message)
	if err != nil {
		return err
	}
	dataLength := proto.EncodeVarint(uint64(len(data)))
	data = append(dataLength, data...)

	_, err = conn.Write(data)
	if err != nil {
		return err
	}
	return nil
}
