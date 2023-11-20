/*
	Copyright NetFoundry Inc.

	Licensed under the Apache License, Version 2.0 (the "License");
	you may not use this file except in compliance with the License.
	You may obtain a copy of the License at

	https://www.apache.org/licenses/LICENSE-2.0

	Unless required by applicable law or agreed to in writing, software
	distributed under the License is distributed on an "AS IS" BASIS,
	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
	See the License for the specific language governing permissions and
	limitations under the License.
*/

package handler_ctrl

import (
	"fmt"
	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/channel/v2"
	"github.com/openziti/ziti/common/pb/ctrl_pb"
	"github.com/openziti/ziti/common/sshpipe"
)

func newSshTunnelDataHandler(registry *sshpipe.Registry) *sshTunnelDataHandler {
	return &sshTunnelDataHandler{
		registry: registry,
	}
}

type sshTunnelDataHandler struct {
	registry *sshpipe.Registry
}

func (*sshTunnelDataHandler) ContentType() int32 {
	return int32(ctrl_pb.ContentType_SshTunnelDataType)
}

func (handler *sshTunnelDataHandler) HandleReceive(msg *channel.Message, ch channel.Channel) {
	connId, _ := msg.GetUint32Header(int32(ctrl_pb.Header_SshTunnelConnIdHeader))
	tunnel := handler.registry.Get(connId)

	if tunnel == nil {
		pfxlog.ContextLogger(ch.Label()).
			WithField("connId", connId).
			Error("no ssh tunnel found for given connection id")

		go func() {
			errorMsg := fmt.Sprintf("invalid conn id '%v", connId)
			replyMsg := channel.NewMessage(int32(ctrl_pb.ContentType_SshTunnelCloseType), []byte(errorMsg))
			replyMsg.PutUint32Header(int32(ctrl_pb.Header_SshTunnelConnIdHeader), connId)
			if sendErr := ch.Send(msg); sendErr != nil {
				pfxlog.ContextLogger(ch.Label()).
					WithField("connId", connId).
					WithError(sendErr).
					Error("failed sending ssh tunnel close message after data with invalid conn")
			}

			_ = ch.Close()
		}()
		return
	}

	if err := tunnel.WriteToClient(msg.Body); err != nil {
		tunnel.Close(err)
	}
}
