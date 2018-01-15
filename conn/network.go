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
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/getgauge/common"
	"github.com/getgauge/gauge/gauge_messages"
	"github.com/golang/protobuf/proto"
)

type response struct {
	result chan *gauge_messages.Message
	err    chan error
	timer  *time.Timer
}

func (r *response) stopTimer() {
	if r.timer != nil {
		r.timer.Stop()
	}
}

func (r *response) addTimer(timeout time.Duration, message *gauge_messages.Message) {
	if timeout > 0 {
		r.timer = time.AfterFunc(timeout, func() {
			r.err <- fmt.Errorf("Request timed out for Message with ID => %v and Type => %s", message.GetMessageId(), message.GetMessageType().String())
		})
	}
}

type messages struct {
	m map[int64]response
	sync.Mutex
}

func (m *messages) get(k int64) response {
	m.Lock()
	defer m.Unlock()
	return m.m[k]
}

func (m *messages) put(k int64, res response) {
	m.Lock()
	defer m.Unlock()
	m.m[k] = res
}

func (m *messages) delete(k int64) {
	m.Lock()
	defer m.Unlock()
	delete(m.m, k)
}

var m = &messages{m: make(map[int64]response)}

func writeDataAndGetResponse(conn net.Conn, messageBytes []byte) ([]byte, error) {
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
	messageID := common.GetUniqueID()
	message.MessageId = messageID

	data, err := proto.Marshal(message)
	if err != nil {
		return err
	}
	return Write(conn, data)
}

func getResponseForGaugeMessage(message *gauge_messages.Message, conn net.Conn, res response, timeout time.Duration) {
	message.MessageId = common.GetUniqueID()
	res.addTimer(timeout, message)
	handle := func(err error) {
		if err != nil {
			res.stopTimer()
			res.err <- err
		}
	}

	data, err := proto.Marshal(message)

	handle(err)
	m.put(message.GetMessageId(), res)

	responseBytes, err := writeDataAndGetResponse(conn, data)
	handle(err)

	responseMessage := &gauge_messages.Message{}
	err = proto.Unmarshal(responseBytes, responseMessage)
	handle(err)

	err = checkUnsupportedResponseMessage(responseMessage)
	handle(err)

	responseRes := m.get(responseMessage.GetMessageId())
	responseRes.stopTimer()
	responseRes.result <- responseMessage
	m.delete(responseMessage.GetMessageId())
}

func checkUnsupportedResponseMessage(message *gauge_messages.Message) error {
	if message.GetMessageType() == gauge_messages.Message_UnsupportedMessageResponse {
		return fmt.Errorf("Unsupported Message response received. Message not supported. %s", message.GetUnsupportedMessageResponse().GetMessage())
	}
	return nil
}

// Sends request to plugin for a message. If response is not received for the given message within the configured timeout, an error is thrown
// To wait indefinitely for the response from the plugin, set timeout value as 0.
func GetResponseForMessageWithTimeout(message *gauge_messages.Message, conn net.Conn, timeout time.Duration) (*gauge_messages.Message, error) {
	res := response{result: make(chan *gauge_messages.Message), err: make(chan error)}
	go getResponseForGaugeMessage(message, conn, res, timeout)
	select {
	case err := <-res.err:
		return nil, err
	case res := <-res.result:
		return res, nil
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

// SendProcessKillMessage sends a KillProcessRequest message through the connection.
func SendProcessKillMessage(connection net.Conn) {
	id := common.GetUniqueID()
	message := &gauge_messages.Message{MessageId: id, MessageType: gauge_messages.Message_KillProcessRequest,
		KillProcessRequest: &gauge_messages.KillProcessRequest{}}

	WriteGaugeMessage(message, connection)
}
