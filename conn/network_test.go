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
	"net"
	"reflect"
	"testing"
	"time"

	"github.com/getgauge/gauge/gauge_messages"
	"github.com/golang/protobuf/proto"
)

var id int64

type mockConn struct {
}

var responseMessage *gauge_messages.Message

func (m mockConn) Read(b []byte) (n int, err error) {
	responseMessage.MessageId = id
	messageBytes, err := proto.Marshal(responseMessage)
	data := append(proto.EncodeVarint(uint64(len(messageBytes))), messageBytes...)
	for i := 0; i < len(data); i++ {
		b[i] = data[i]
	}
	return len(data), nil
}

func (m mockConn) Write(b []byte) (n int, err error) {
	message := &gauge_messages.Message{}
	messageLength, bytesRead := proto.DecodeVarint(b)
	b = b[bytesRead : messageLength+uint64(bytesRead)]
	proto.Unmarshal(b, message)
	if id == 0 {
		id = message.MessageId
	}
	return 0, nil
}

func (m mockConn) Close() error {
	return nil
}

func (m mockConn) LocalAddr() net.Addr {
	return nil
}

func (m mockConn) RemoteAddr() net.Addr {
	return nil
}

func (m mockConn) SetDeadline(t time.Time) error {
	return nil
}

func (m mockConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (m mockConn) SetWriteDeadline(t time.Time) error {
	return nil
}

func TestGetResponseForGaugeMessageWithTimeout(t *testing.T) {
	message := &gauge_messages.Message{
		MessageType: gauge_messages.Message_StepNameRequest,
		StepNameRequest: &gauge_messages.StepNameRequest{
			StepValue: "The worrd {} has {} vowels.",
		},
	}

	responseMessage = &gauge_messages.Message{
		MessageType: gauge_messages.Message_StepNameResponse,
		StepNameResponse: &gauge_messages.StepNameResponse{
			FileName:      "foo.js",
			HasAlias:      false,
			IsStepPresent: true,
			LineNumber:    2,
			StepName:      []string{"The word {} has {} vowels."},
		},
	}

	conn := mockConn{}

	res, err := GetResponseForMessageWithTimeout(message, conn, 3*time.Second)

	if err != nil {
		t.Errorf("expected err to be nil. got %v", err)
	}
	if !reflect.DeepEqual(res, responseMessage) {
		t.Errorf("expected : %v\ngot : %v", responseMessage, res)
	}
}

func TestGetResponseForGaugeMessageShoudGiveTheRightResponse(t *testing.T) {
	id = 1234
	r := response{
		err:    make(chan error),
		result: make(chan *gauge_messages.Message),
	}

	m.put(id, r)

	message := &gauge_messages.Message{
		MessageType: gauge_messages.Message_StepNameRequest,
		StepNameRequest: &gauge_messages.StepNameRequest{
			StepValue: "The worrd {} has {} vowels.",
		},
	}

	responseMessage = &gauge_messages.Message{
		MessageType: gauge_messages.Message_StepNameResponse,
		StepNameResponse: &gauge_messages.StepNameResponse{
			FileName:      "foo.js",
			HasAlias:      false,
			IsStepPresent: true,
			LineNumber:    2,
			StepName:      []string{"The word {} has {} vowels."},
		},
	}

	conn := mockConn{}

	go getResponseForGaugeMessage(message, conn, response{})

	response := <-r.result
	if !reflect.DeepEqual(response, responseMessage) {
		t.Errorf("expected : %v\ngot : %v", responseMessage, response)
	}
}
