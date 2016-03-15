// Copyright 2015 ThoughtWorks, Inc.

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

package conn

import (
	"bytes"
	"fmt"
	"net"
	"time"

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
	listener, err := net.ListenTCP("tcp", &net.TCPAddr{Port: port})
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
			conn.Close()
			//TODO: Move to file
			//			logger.Log.Println(fmt.Sprintf("Closing connection [%s] cause: %s", connectionHandler.conn.RemoteAddr(), err.Error()))
			return
		}

		buffer.Write(data[0:n])
		connectionHandler.processMessage(buffer, conn)
	}
}

func (connectionHandler *GaugeConnectionHandler) processMessage(buffer *bytes.Buffer, conn net.Conn) {
	for {
		messageLength, bytesRead := proto.DecodeVarint(buffer.Bytes())
		if messageLength > 0 && messageLength < uint64(buffer.Len()) {
			messageBoundary := int(messageLength) + bytesRead
			receivedBytes := buffer.Bytes()[bytesRead : messageLength+uint64(bytesRead)]
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
		connectionHandler.acceptConnectionWithoutTimeout()
	}

}

func (connectionHandler *GaugeConnectionHandler) ConnectionPortNumber() int {
	if connectionHandler.tcpListener != nil {
		return connectionHandler.tcpListener.Addr().(*net.TCPAddr).Port
	}
	return 0
}
