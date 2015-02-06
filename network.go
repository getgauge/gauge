// Copyright 2014 ThoughtWorks, Inc.

// This file is part of Gauge.

// Gauge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// Gauge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with Gauge.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/getgauge/common"
	"github.com/getgauge/gauge/gauge_messages"
	"github.com/golang/protobuf/proto"
	"net"
	"os"
	"strconv"
	"time"
)

type messageHandler interface {
	messageBytesReceived([]byte, net.Conn)
}

type dataHandlerFn func(*gaugeConnectionHandler, []byte)

type gaugeConnectionHandler struct {
	tcpListener    *net.TCPListener
	messageHandler messageHandler
}

func newGaugeConnectionHandler(port int, messageHandler messageHandler) (*gaugeConnectionHandler, error) {
	// port = 0 means GO will find a unused port

	listener, err := net.ListenTCP("tcp", &net.TCPAddr{Port: port})
	if err != nil {
		return nil, err
	}

	return &gaugeConnectionHandler{tcpListener: listener, messageHandler: messageHandler}, nil
}

func (connectionHandler *gaugeConnectionHandler) acceptConnection(connectionTimeOut time.Duration, errChannel chan error) (net.Conn, error) {
	connectionChannel := make(chan net.Conn)

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
		return nil, err
	case conn := <-connectionChannel:
		if connectionHandler.messageHandler != nil {
			go connectionHandler.handleConnectionMessages(conn)
		}
		return conn, nil
	case <-time.After(connectionTimeOut):
		return nil, errors.New(fmt.Sprintf("Timed out connecting to %v", connectionHandler.tcpListener.Addr()))
	}
}

func (connectionHandler *gaugeConnectionHandler) acceptConnectionWithoutTimeout() (net.Conn, error) {
	errChannel := make(chan error)
	connectionChannel := make(chan net.Conn)

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
		return nil, err
	case conn := <-connectionChannel:
		if connectionHandler.messageHandler != nil {
			go connectionHandler.handleConnectionMessages(conn)
		}
		return conn, nil
	}
}

func (connectionHandler *gaugeConnectionHandler) handleConnectionMessages(conn net.Conn) {
	buffer := new(bytes.Buffer)
	data := make([]byte, 8192)
	for {
		n, err := conn.Read(data)
		if err != nil {
			conn.Close()
			//TODO: Move to file
			//			log.Println(fmt.Sprintf("Closing connection [%s] cause: %s", connectionHandler.conn.RemoteAddr(), err.Error()))
			return
		}

		buffer.Write(data[0:n])
		connectionHandler.processMessage(buffer, conn)
	}
}

func (connectionHandler *gaugeConnectionHandler) processMessage(buffer *bytes.Buffer, conn net.Conn) {
	for {
		messageLength, bytesRead := proto.DecodeVarint(buffer.Bytes())
		if messageLength > 0 && messageLength < uint64(buffer.Len()) {
			messageBoundary := int(messageLength) + bytesRead
			receivedBytes := buffer.Bytes()[bytesRead : messageLength+uint64(bytesRead)]
			connectionHandler.messageHandler.messageBytesReceived(receivedBytes, conn)
			buffer.Next(messageBoundary)
			if buffer.Len() == 0 {
				return
			}
		} else {
			return
		}
	}
}

func writeDataAndGetResponse(conn net.Conn, messageBytes []byte) ([]byte, error) {
	if err := write(conn, messageBytes); err != nil {
		return nil, err
	}

	return readResponse(conn)
}

func readResponse(conn net.Conn) ([]byte, error) {
	buffer := new(bytes.Buffer)
	data := make([]byte, 8192)
	for {
		n, err := conn.Read(data)
		if err != nil {
			conn.Close()
			return nil, errors.New(fmt.Sprintf("Connection closed [%s] cause: %s", conn.RemoteAddr(), err.Error()))
		}

		buffer.Write(data[0:n])

		messageLength, bytesRead := proto.DecodeVarint(buffer.Bytes())
		if messageLength > 0 && messageLength < uint64(buffer.Len()) {
			return buffer.Bytes()[bytesRead : messageLength+uint64(bytesRead)], nil
		}
	}
}

func write(conn net.Conn, messageBytes []byte) error {
	messageLen := proto.EncodeVarint(uint64(len(messageBytes)))
	data := append(messageLen, messageBytes...)
	_, err := conn.Write(data)
	return err
}

//accepts multiple connections and Handler responds to incoming messages
func (connectionHandler *gaugeConnectionHandler) handleMultipleConnections() {
	for {
		connectionHandler.acceptConnectionWithoutTimeout()
	}

}

func (connectionHandler *gaugeConnectionHandler) connectionPortNumber() int {
	if connectionHandler.tcpListener != nil {
		return connectionHandler.tcpListener.Addr().(*net.TCPAddr).Port
	} else {
		return 0
	}
}

func writeGaugeMessage(message *gauge_messages.Message, conn net.Conn) error {
	messageId := common.GetUniqueId()
	message.MessageId = &messageId

	data, err := proto.Marshal(message)
	if err != nil {
		return err
	}
	return write(conn, data)
}

func getResponseForGaugeMessage(message *gauge_messages.Message, conn net.Conn) (*gauge_messages.Message, error) {
	messageId := common.GetUniqueId()
	message.MessageId = &messageId

	data, err := proto.Marshal(message)
	if err != nil {
		return nil, err
	}
	responseBytes, err := writeDataAndGetResponse(conn, data)
	if err != nil {
		return nil, err
	}
	responseMessage := &gauge_messages.Message{}
	err = proto.Unmarshal(responseBytes, responseMessage)
	if err != nil {
		return nil, err
	}
	return responseMessage, err
}

func getResponseForMessageWithTimeout(message *gauge_messages.Message, conn net.Conn, t time.Duration) (*gauge_messages.Message, error) {
	timeout := make(chan bool, 1)
	received := make(chan bool, 1)
	go func() {
		time.Sleep(t)
		timeout <- true
	}()
	var response *gauge_messages.Message
	var error error
	go func() {
		response, error = getResponseForGaugeMessage(message, conn)
		received <- true
		close(received)
	}()
	select {
	case <-received:
		return response, error
	case <-timeout:
		return nil, errors.New("Request Timeout")
	}
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
