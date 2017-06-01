package debug

import (
	"fmt"
	"net"

	"github.com/getgauge/common"
	"github.com/getgauge/gauge/conn"
	m "github.com/getgauge/gauge/gauge_messages"
	"github.com/golang/protobuf/proto"
)

type api struct {
	connection net.Conn
}

func newAPI(host string, port string) (*api, error) {
	c, err := net.Dial("tcp", fmt.Sprintf("%s:%s", host, port))
	return &api{connection: c}, err
}

func (a *api) getResponse(message *m.APIMessage) (*m.APIMessage, error) {
	messageId := common.GetUniqueID()
	message.MessageId = messageId

	data, err := proto.Marshal(message)
	if err != nil {
		return nil, err
	}
	responseBytes, err := conn.WriteDataAndGetResponse(a.connection, data)
	if err != nil {
		return nil, err
	}
	responseMessage := &m.APIMessage{}
	if err := proto.Unmarshal(responseBytes, responseMessage); err != nil {
		return nil, err
	}
	return responseMessage, err
}

func (a *api) close() {
	a.connection.Close()
}
