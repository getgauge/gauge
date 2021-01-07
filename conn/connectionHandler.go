/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package conn

import (
	"bytes"
	"fmt"
	"net"
	"time"

	"github.com/getgauge/gauge/logger"
	"github.com/golang/protobuf/proto"
)

type messageHandler interface {
	MessageBytesReceived([]byte, net.Conn)
}

type GaugeConnectionHandler struct {
	tcpListener    *net.TCPListener
	messageHandler messageHandler
}

func NewGaugeConnectionHandler(port int, messageHandler messageHandler) (*GaugeConnectionHandler, error) {
	// port = 0 means GO will find a unused port
	address, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return nil, err
	}
	listener, err := net.ListenTCP("tcp", address)
	if err != nil {
		return nil, err
	}

	return &GaugeConnectionHandler{tcpListener: listener, messageHandler: messageHandler}, nil
}

func (connectionHandler *GaugeConnectionHandler) AcceptConnection(connectionTimeOut time.Duration, errChannel chan error) (net.Conn, error) {
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
		return nil, fmt.Errorf("Timed out connecting to %v", connectionHandler.tcpListener.Addr())
	}
}

func (connectionHandler *GaugeConnectionHandler) acceptConnectionWithoutTimeout() (net.Conn, error) {
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

func (connectionHandler *GaugeConnectionHandler) handleConnectionMessages(conn net.Conn) {
	buffer := new(bytes.Buffer)
	data := make([]byte, 8192)
	for {
		n, err := conn.Read(data)
		if err != nil {
			e := conn.Close()
			if e != nil {
				logger.Debugf(false, "Connection already closed, %s", e.Error())
			}
			logger.Infof(false, "Closing connection [%s] cause: %s", conn.RemoteAddr(), err.Error())
			return
		}

		_, err = buffer.Write(data[0:n])
		if err != nil {
			logger.Infof(false, "Unable to write to buffer, %s", err.Error())
			return
		}
		connectionHandler.processMessage(buffer, conn)
	}
}

func (connectionHandler *GaugeConnectionHandler) processMessage(buffer *bytes.Buffer, conn net.Conn) {
	for {
		messageLength, bytesRead := proto.DecodeVarint(buffer.Bytes())
		if messageLength > 0 && messageLength < uint64(buffer.Len()) {
			messageBoundary := int(messageLength) + bytesRead
			receivedBytes := buffer.Bytes()[bytesRead:messageBoundary]
			connectionHandler.messageHandler.MessageBytesReceived(receivedBytes, conn)
			buffer.Next(messageBoundary)
			if buffer.Len() == 0 {
				return
			}
		} else {
			return
		}
	}
}

// HandleMultipleConnections accepts multiple connections and Handler responds to incoming messages
func (connectionHandler *GaugeConnectionHandler) HandleMultipleConnections() {
	for {
		_, err := connectionHandler.acceptConnectionWithoutTimeout()
		if err != nil {
			logger.Fatalf(true, "Unable to connect to runner: %s", err.Error())
		}
	}

}

func (connectionHandler *GaugeConnectionHandler) ConnectionPortNumber() int {
	if connectionHandler.tcpListener != nil {
		return connectionHandler.tcpListener.Addr().(*net.TCPAddr).Port
	}
	return 0
}
