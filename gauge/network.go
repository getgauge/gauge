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

type messageHandler interface {
	messageBytesReceived([]byte, *gaugeConnectionHandler)
}

type dataHandlerFn func(*gaugeConnectionHandler, []byte)

type gaugeConnectionHandler struct {
	tcpListener    *net.TCPListener
	messageHandler messageHandler
	conn           net.Conn
}

func newGaugeConnectionHandler(port int, messageHandler messageHandler) (*gaugeConnectionHandler, error) {
	// port = 0 means GO will find a unused port

	listener, err := net.ListenTCP("tcp", &net.TCPAddr{Port: port})
	if err != nil {
		return nil, err
	}

	return &gaugeConnectionHandler{tcpListener: listener, messageHandler: messageHandler}, nil
}

func (connectionHandler *gaugeConnectionHandler) acceptConnection(connectionTimeOut time.Duration) error {
	errChannel := make(chan error, 1)
	connectionChannel := make(chan net.Conn, 1)

	go func() {
		connection, err := connectionHandler.tcpListener.Accept()
		if err != nil {
			errChannel <- err
		}
		if connection != nil {
			connectionChannel <- connection
		}
	}()

	select {
	case err := <-errChannel:
		return err
	case conn := <-connectionChannel:
		if connectionHandler.messageHandler != nil {
			go connectionHandler.handleConnectionMessages()
		}
		connectionHandler.conn = conn
		return nil
	case <-time.After(connectionTimeOut):
		return errors.New(fmt.Sprintf("Timed out connecting to %v", connectionHandler.tcpListener.Addr()))
	}
}

func (connectionHandler *gaugeConnectionHandler) handleConnectionMessages() {
	buffer := new(bytes.Buffer)
	data := make([]byte, 8192)
	for {
		n, err := connectionHandler.conn.Read(data)
		if err != nil {
			connectionHandler.conn.Close()
			//TODO: Move to file
			log.Println(fmt.Sprintf("Closing connection [%s] cause: %s", connectionHandler.conn.RemoteAddr(), err.Error()))
			return
		}

		buffer.Write(data[0:n])

		messageLength, bytesRead := proto.DecodeVarint(buffer.Bytes())
		if messageLength > 0 && messageLength < uint64(buffer.Len()) {
			receivedBytes := buffer.Bytes()[bytesRead : messageLength+uint64(bytesRead)]
			connectionHandler.messageHandler.messageBytesReceived(receivedBytes, connectionHandler)
			buffer.Reset()
		}
	}
}

func (connectionHandler *gaugeConnectionHandler) writeDataAndGetResponse(messageBytes []byte) ([]byte, error) {
	if err := connectionHandler.write(messageBytes); err != nil {
		return nil, err
	}

	return connectionHandler.readResponse()
}

func (connectionHandler *gaugeConnectionHandler) readResponse() ([]byte, error) {
	buffer := new(bytes.Buffer)
	data := make([]byte, 8192)
	for {
		n, err := connectionHandler.conn.Read(data)
		if err != nil {
			connectionHandler.conn.Close()
			return nil, errors.New(fmt.Sprintf("Connection closed [%s] cause: %s", connectionHandler.conn.RemoteAddr(), err.Error()))
		}

		buffer.Write(data[0:n])

		messageLength, bytesRead := proto.DecodeVarint(buffer.Bytes())
		if messageLength > 0 && messageLength < uint64(buffer.Len()) {
			return buffer.Bytes()[bytesRead : messageLength+uint64(bytesRead)], nil
		}
	}
}

func (connectionHandler *gaugeConnectionHandler) write(messageBytes []byte) error {
	messageLen := proto.EncodeVarint(uint64(len(messageBytes)))
	data := append(messageLen, messageBytes...)
	_, err := connectionHandler.conn.Write(data)
	return err
}

//accepts multiple connections and Handler responds to incoming messages
func (connectionHandler *gaugeConnectionHandler) handleMultipleConnections() {
	for {
		connectionHandler.acceptConnection(30 * time.Second)
	}

}

func (connectionHandler *gaugeConnectionHandler) connectionPortNumber() int {
	if connectionHandler.tcpListener != nil {
		return connectionHandler.tcpListener.Addr().(*net.TCPAddr).Port
	} else {
		return 0
	}
}

func writeGaugeMessage(message *Message, connectionHandler *gaugeConnectionHandler) error {
	messageId := common.GetUniqueId()
	message.MessageId = &messageId

	data, err := proto.Marshal(message)
	if err != nil {
		return err
	}
	return connectionHandler.write(data)
}

func getResponseForGaugeMessage(message *Message, connectionHandler *gaugeConnectionHandler) (*Message, error) {
	messageId := common.GetUniqueId()
	message.MessageId = &messageId

	data, err := proto.Marshal(message)
	if err != nil {
		return nil, err
	}
	responseBytes, err := connectionHandler.writeDataAndGetResponse(data)
	if err != nil {
		return nil, err
	}
	responseMessage := &Message{}
	err = proto.Unmarshal(responseBytes, responseMessage)
	if err != nil {
		return nil, err
	}
	return responseMessage, err
}

func getPortFromEnvironmentVariable(portEnvVariable string) (int, error) {
	if port := os.Getenv(portEnvVariable); port != "" {
		gport, err := strconv.Atoi(port)
		if err != nil {
			return 0, errors.New(fmt.Sprintf("%s is not a valid port", port))
		}
		return gport, nil
	}
	return 0, errors.New(fmt.Sprintf("%s Environment variable not set", portEnvVariable))
}
