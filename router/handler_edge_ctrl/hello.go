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

package handler_edge_ctrl

import (
	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/channel/v2"
	"github.com/openziti/channel/v2/protobufs"
	"github.com/openziti/edge/controller/env"
	"github.com/openziti/edge/pb/edge_ctrl_pb"
	"github.com/openziti/fabric/build"
	"google.golang.org/protobuf/proto"
	"strconv"
)

type helloHandler struct {
	listeners []*edge_ctrl_pb.Listener

	//backwards compat for controllers v0.26.3 and older
	hostname           string
	supportedProtocols []string
	protocolPorts      []string
}

func NewHelloHandler(listeners []*edge_ctrl_pb.Listener) *helloHandler {
	//supportedProtocols, protocolPorts, and hostname is for backwards compatability with v0.26.3 and older controllers
	var supportedProtocols []string
	var protocolPorts []string
	hostname := ""

	for _, listener := range listeners {
		pfxlog.Logger().Debugf("HelloHandler will contain supportedProtocols address: %s advertise: %s", listener.Address.Value, listener.Advertise.Value)

		supportedProtocols = append(supportedProtocols, listener.Address.Protocol)
		protocolPorts = append(protocolPorts, strconv.Itoa(int(listener.Advertise.Port)))

		if hostname != "" && hostname != listener.Advertise.Hostname {
			pfxlog.Logger().Warnf("this router is configured to use different hostnames for different edge listeners. If the controller is v0.26.3 or earlier this is not supported. Advertise %s will be used for all protocls", listeners[0].Advertise.Value)
		}

		hostname = listener.Advertise.Hostname
	}

	return &helloHandler{
		listeners: listeners,

		//v0.26.3 and older used to check and ensure all advertise hostnames were the same which can't be done now
		//with the ability to report multiple advertise protocols on different hostnames
		hostname:           listeners[0].Advertise.Hostname,
		supportedProtocols: supportedProtocols,
		protocolPorts:      protocolPorts,
	}
}

func (h *helloHandler) ContentType() int32 {
	return env.ServerHelloType
}

func (h *helloHandler) HandleReceive(msg *channel.Message, ch channel.Channel) {
	go func() {
		serverHello := &edge_ctrl_pb.ServerHello{}
		if err := proto.Unmarshal(msg.Body, serverHello); err == nil {
			pfxlog.Logger().Info("received server hello, replying")

			clientHello := &edge_ctrl_pb.ClientHello{
				Version:   build.GetBuildInfo().Version(),
				Listeners: h.listeners,

				Hostname:      h.hostname,
				Protocols:     h.supportedProtocols,
				ProtocolPorts: h.protocolPorts,
			}
			if err := protobufs.MarshalTyped(clientHello).ReplyTo(msg).Send(ch); err != nil {
				pfxlog.Logger().WithError(err).Error("could not send client hello")
			}
			return
		} else {
			pfxlog.Logger().WithError(err).Error("could not unmarshal server hello")
		}
	}()
}
