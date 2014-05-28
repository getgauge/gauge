package main

import (
	"bytes"
	"code.google.com/p/goprotobuf/proto"
	"errors"
	"fmt"
	"github.com/getgauge/common"
	"log"
	"net"
	"os"
	"strconv"
	"time"
)

const (
	runnerConnectionPort    = ":8888"
	runnerConnectionTimeOut = time.Second * 10
)

type MessageHandler interface {
	messageReceived([]byte, net.Conn)
}

// MessageId -> Callback
var pendingRequests = make(map[int64]chan<- *Message)

func handleConnection(conn net.Conn) {
	buffer := new(bytes.Buffer)
	data := make([]byte, 8192)
	for {
		n, err := conn.Read(data)
		if err != nil {
			conn.Close()
			//TODO: Move to file
			//log.Println(fmt.Sprintf("Closing connection [%s] cause: %s", conn.RemoteAddr(), err.Error()))
			return
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

func newGaugeListener(portEnvVariable string, staticBackupPort int) (*gaugeListener, error) {
	// if portEnvVariable is set, use that. Else try backup port. or Finally ListenTCP will assign a free port
	// port = 0 means GO will find a unused port
	port, err := getPortFromEnvironmentVariable(portEnvVariable)
	if err != nil {
		return nil, err
	}
	if port == 0 {
		port = staticBackupPort
	}

	listener, err := net.ListenTCP("tcp", &net.TCPAddr{Port: port})
	if err != nil {
		return nil, err
	}
	return &gaugeListener{tcpListener: listener}, nil
}

func getPortFromEnvironmentVariable(portEnvVariable string) (int, error) {
	if port := os.Getenv(portEnvVariable); port != "" {
		gport, err := strconv.Atoi(port)
		if err != nil {
			return 0, errors.New(fmt.Sprintf("%s is not a valid port", port))
		}
		return gport, nil
	}
	return 0, nil
}

func (listener *gaugeListener) acceptConnection(connectionTimeOut time.Duration) (net.Conn, error) {
	errChannel := make(chan error, 1)
	connectionChannel := make(chan net.Conn, 1)

	go func() {
		connection, err := listener.tcpListener.Accept()
		if err != nil {
			errChannel <- err
		}
		if connection != nil {
			connectionChannel <- connection
		}
	}()

	select {
	case err := <-errChannel:
		return nil, err
	case conn := <-connectionChannel:
		go handleConnection(conn)
		return conn, nil
	case <-time.After(connectionTimeOut):
		return nil, errors.New(fmt.Sprintf("Timed out connecting to %v", listener.tcpListener.Addr()))
	}

}

func (listener *gaugeListener) acceptAndHandleMultipleConnections(messageHandler MessageHandler) {
	connectionChannel := make(chan net.Conn, 1)
	go listener.acceptConnections(connectionChannel)
	go listener.handleConnections(connectionChannel, messageHandler)
}

func (listener *gaugeListener) acceptConnections(connectionChannel chan<- net.Conn) {
	go func() {
		for {
			connection, err := listener.tcpListener.Accept()
			if err == nil {
				connectionChannel <- connection
			}
		}
	}()
}

func (listener *gaugeListener) handleConnections(connectionChannel <-chan net.Conn, messageHandler MessageHandler) {
	for {
		connection := <-connectionChannel
		go listener.handleConnection(connection, messageHandler)
	}
}

func (gaugeListener *gaugeListener) handleConnection(conn net.Conn, messageHandler MessageHandler) {
	buffer := new(bytes.Buffer)
	data := make([]byte, 8192)
	for {
		n, err := conn.Read(data)
		if err != nil {
			conn.Close()
			//TODO: Move to file
			//log.Println(fmt.Sprintf("Closing connection [%s] cause: %s", conn.RemoteAddr(), err.Error()))
			return
		}

		buffer.Write(data[0:n])

		messageLength, bytesRead := proto.DecodeVarint(buffer.Bytes())
		if messageLength > 0 && messageLength < uint64(buffer.Len()) {
			messageHandler.messageReceived(buffer.Bytes()[bytesRead:messageLength+uint64(bytesRead)], conn)
			buffer.Reset()
		}
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
func writeMessage(conn net.Conn, message proto.Message) error {
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
