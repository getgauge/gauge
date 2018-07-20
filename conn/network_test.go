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
	"errors"
	"fmt"
	"net"
	"reflect"
	"testing"
	"time"

	"github.com/getgauge/gauge/gauge_messages"
	"github.com/golang/protobuf/proto"
)

var id int64

type mockConn struct {
	sleepDuration time.Duration
}

var responseMessage *gauge_messages.Message

func (m mockConn) Read(b []byte) (n int, err error) {
	time.Sleep(m.sleepDuration)
	if responseMessage.MessageId == 0 {
		responseMessage.MessageId = id
	}
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
	id = 0
	responseMessage = &gauge_messages.Message{}
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
			Span:          &gauge_messages.Span{Start: 2, End: 6, StartChar: 0, EndChar: 2},
			StepName:      []string{"The word {} has {} vowels."},
		},
	}

	conn := mockConn{}

	res, err := GetResponseForMessageWithTimeout(message, conn, 3*time.Second)

	if err != nil {
		t.Errorf("expected err to be nil. got %v", err)
	}
	if !proto.Equal(res, responseMessage) {
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
			Span:          &gauge_messages.Span{Start: 2, End: 2, StartChar: 0, EndChar: 2},
			StepName:      []string{"The word {} has {} vowels."},
		},
	}

	conn := mockConn{}

	go getResponseForGaugeMessage(message, conn, response{}, 3*time.Second)

	response := <-r.result
	if !proto.Equal(response, responseMessage) {
		t.Errorf("expected : %v\ngot : %v", responseMessage, response)
	}
}

func TestGetResponseForGaugeMessageShoudGiveErrorForUnsupportedMessage(t *testing.T) {
	id = 0
	message := &gauge_messages.Message{
		MessageType: gauge_messages.Message_StepNameRequest,
		StepNameRequest: &gauge_messages.StepNameRequest{
			StepValue: "The worrd {} has {} vowels.",
		},
	}

	responseMessage = &gauge_messages.Message{
		MessageType:                gauge_messages.Message_UnsupportedMessageResponse,
		UnsupportedMessageResponse: &gauge_messages.UnsupportedMessageResponse{},
	}

	conn := mockConn{}

	_, err := GetResponseForMessageWithTimeout(message, conn, 1*time.Second)

	expected := errors.New("Unsupported Message response received. Message not supported.")

	if reflect.DeepEqual(err, expected) {
		t.Errorf("expected %v\n got %v", expected, err)
	}

}

func TestGetResponseForGaugeMessageShoudErrorWithTimeOut(t *testing.T) {
	id = 0
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
			Span:          &gauge_messages.Span{Start: 2, End: 2, StartChar: 0, EndChar: 2},
			StepName:      []string{"The word {} has {} vowels."},
		},
	}

	conn := mockConn{sleepDuration: 2 * time.Second}
	_, err := GetResponseForMessageWithTimeout(message, conn, 1*time.Second)

	expected := fmt.Errorf("Request timed out for Message with ID => %v and Type => StepNameRequest", id)
	if !reflect.DeepEqual(err, expected) {
		t.Errorf("expected %v\n got %v", expected, err)
	}
}

func TestGetResponseForGaugeMessageShoudNotErrorIfNoTimeoutIsSpecified(t *testing.T) {
	id = 0
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
			Span:          &gauge_messages.Span{Start: 2, End: 6, StartChar: 0, EndChar: 2},
			StepName:      []string{"The word {} has {} vowels."},
		},
	}

	conn := mockConn{}

	res, err := GetResponseForMessageWithTimeout(message, conn, 0)

	if err != nil {
		t.Errorf("expected err to be nil. got %v", err)
	}
	if !proto.Equal(res, responseMessage) {
		t.Errorf("expected : %v\ngot : %v", responseMessage, res)
	}
}
