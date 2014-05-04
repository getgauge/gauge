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
	"os"
	"strconv"
)

const (
	timeoutInSeconds = 30
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

type gaugeListener struct {
	tcpListener *net.TCPListener
}

func newListener() (*gaugeListener, error) {
	// if GAUGE_PORT is set, use that. Else ListenTCP will assign a free port and set that to GAUGE_ROOT
	// port = 0 means GO will find a unused port
	port := 0
	if gaugePort := os.Getenv(common.GaugePortEnvName); gaugePort != "" {
		gport, err := strconv.Atoi(gaugePort)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("%s is not a valid port", gaugePort))
		}
		port = gport
	}

	listener, err := net.ListenTCP("tcp", &net.TCPAddr{Port: port})
	if err != nil {
		return nil, err
	}

	if err := common.SetEnvVariable(common.GaugeInternalPortEnvName, strconv.Itoa(listener.Addr().(*net.TCPAddr).Port)); err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to set %s. %s", common.GaugePortEnvName, err.Error()))
	}

	return &gaugeListener{tcpListener: listener}, nil
}

func (listener *gaugeListener) acceptConnection() (net.Conn, error) {
	conn, err := listener.tcpListener.Accept()
	if err != nil {
		return nil, err
	}

	go handleConnection(conn)
	return conn, nil
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
