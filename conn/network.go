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

func WriteDataAndGetResponse(conn net.Conn, messageBytes []byte) ([]byte, error) {
	if err := Write(conn, messageBytes); err != nil {
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
			return nil, fmt.Errorf("Connection closed [%s] cause: %s", conn.RemoteAddr(), err.Error())
		}

		buffer.Write(data[0:n])

		messageLength, bytesRead := proto.DecodeVarint(buffer.Bytes())
		if messageLength > 0 && messageLength < uint64(buffer.Len()) {
			return buffer.Bytes()[bytesRead : messageLength+uint64(bytesRead)], nil
		}
	}
}

func Write(conn net.Conn, messageBytes []byte) error {
	messageLen := proto.EncodeVarint(uint64(len(messageBytes)))
	data := append(messageLen, messageBytes...)
	_, err := conn.Write(data)
	return err
}

func WriteGaugeMessage(message *gauge_messages.Message, conn net.Conn) error {
	messageId := common.GetUniqueID()
	message.MessageId = &messageId

	data, err := proto.Marshal(message)
	if err != nil {
		return err
	}
	return Write(conn, data)
}

func GetResponseForGaugeMessage(message *gauge_messages.Message, conn net.Conn) (*gauge_messages.Message, error) {
	messageId := common.GetUniqueID()
	message.MessageId = &messageId

	data, err := proto.Marshal(message)
	if err != nil {
		return nil, err
	}
	responseBytes, err := WriteDataAndGetResponse(conn, data)
	if err != nil {
		return nil, err
	}
	responseMessage := &gauge_messages.Message{}
	if err := proto.Unmarshal(responseBytes, responseMessage); err != nil {
		return nil, err
	}

	if err := checkUnsupportedResponseMessage(responseMessage); err != nil {
		return responseMessage, err
	}
	return responseMessage, err
}

func checkUnsupportedResponseMessage(message *gauge_messages.Message) error {
	if message.GetMessageType() == gauge_messages.Message_UnsupportedMessageResponse {
		return fmt.Errorf("Unsupported Message response received. Message not supported. %s", message.GetUnsupportedMessageResponse().GetMessage())
	}
	return nil
}

func GetResponseForMessageWithTimeout(message *gauge_messages.Message, conn net.Conn, t time.Duration) (*gauge_messages.Message, error) {
	responseChan := make(chan bool, 1)
	errChan := make(chan bool, 1)

	var response *gauge_messages.Message
	var err error
	go func() {
		response, err = GetResponseForGaugeMessage(message, conn)
		if err != nil {
			errChan <- true
			close(errChan)
		} else {
			responseChan <- true
			close(responseChan)
		}
	}()
	select {
	case <-errChan:
		return nil, err
	case <-responseChan:
		return response, nil
	case <-time.After(t):
		return nil, errors.New("Request Timeout")
	}
}

func GetPortFromEnvironmentVariable(portEnvVariable string) (int, error) {
	if port := os.Getenv(portEnvVariable); port != "" {
		gport, err := strconv.Atoi(port)
		if err != nil {
			return 0, fmt.Errorf("%s is not a valid port", port)
		}
		return gport, nil
	}
	return 0, fmt.Errorf("%s Environment variable not set", portEnvVariable)
}
