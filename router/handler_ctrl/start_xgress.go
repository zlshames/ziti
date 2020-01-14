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

package handler_ctrl

import (
	"github.com/michaelquigley/pfxlog"
	"github.com/netfoundry/ziti-fabric/pb/ctrl_pb"
	"github.com/netfoundry/ziti-fabric/router/forwarder"
	"github.com/netfoundry/ziti-foundation/channel2"
	"github.com/netfoundry/ziti-foundation/identity/identity"
)

type startXgressHandler struct {
	forwarder *forwarder.Forwarder
}

func newStartXgressHandler(forwarder *forwarder.Forwarder) *startXgressHandler {
	return &startXgressHandler{forwarder: forwarder}
}

func (h *startXgressHandler) ContentType() int32 {
	return int32(ctrl_pb.ContentType_StartXgressType)
}

func (h *startXgressHandler) HandleReceive(msg *channel2.Message, ch channel2.Channel) {
	token := string(msg.Body)
	pfxlog.Logger().Infof("attempting to start xgress for session %v", token)
	sessionId := &identity.TokenId{Token: token}
	h.forwarder.StartDestinations(sessionId)
}
