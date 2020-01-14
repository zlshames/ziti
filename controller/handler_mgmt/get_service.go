/*
	Copyright 2019 NetFoundry, Inc.

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

package handler_mgmt

import (
	"github.com/golang/protobuf/proto"
	"github.com/michaelquigley/pfxlog"
	"github.com/netfoundry/ziti-fabric/controller/network"
	"github.com/netfoundry/ziti-fabric/pb/mgmt_pb"
	"github.com/netfoundry/ziti-foundation/channel2"
)

type getServiceHandler struct {
	network *network.Network
}

func newGetServiceHandler(network *network.Network) *getServiceHandler {
	return &getServiceHandler{network: network}
}

func (h *getServiceHandler) ContentType() int32 {
	return int32(mgmt_pb.ContentType_GetServiceRequestType)
}

func (h *getServiceHandler) HandleReceive(msg *channel2.Message, ch channel2.Channel) {
	rs := &mgmt_pb.GetServiceRequest{}
	err := proto.Unmarshal(msg.Body, rs)
	if err == nil {
		response := &mgmt_pb.GetServiceResponse{}
		if svc, found := h.network.GetService(rs.ServiceId); found {
			response.Service = &mgmt_pb.Service{
				Id:              svc.Id,
				Binding:         svc.Binding,
				EndpointAddress: svc.EndpointAddress,
				Egress:          svc.Egress,
			}

			body, err := proto.Marshal(response)
			if err == nil {
				responseMsg := channel2.NewMessage(int32(mgmt_pb.ContentType_GetServiceResponseType), body)
				responseMsg.ReplyTo(msg)
				ch.Send(responseMsg)

			} else {
				pfxlog.ContextLogger(ch.Label()).Errorf("unexpected error (%s)", err)
			}

		} else {
			sendFailure(msg, ch, "no such service")
		}
	} else {
		sendFailure(msg, ch, err.Error())
	}
}
