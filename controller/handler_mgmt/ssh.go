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

package handler_mgmt

import (
	"fmt"
	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/channel/v2"
	"github.com/openziti/channel/v2/protobufs"
	"github.com/openziti/ziti/common/pb/ctrl_pb"
	"github.com/openziti/ziti/common/pb/mgmt_pb"
	"github.com/openziti/ziti/common/sshpipe"
	"github.com/openziti/ziti/controller/network"
	"google.golang.org/protobuf/proto"
	"io"
	"net"
	"sync/atomic"
	"time"
)

type sshTunnelHandler struct {
	network  *network.Network
	registry *sshpipe.Registry
	tunnel   sshpipe.Pipe
	ch       channel.Channel
}

func newSshTunnelHandler(network *network.Network, registry *sshpipe.Registry, ch channel.Channel) *sshTunnelHandler {
	return &sshTunnelHandler{
		network:  network,
		registry: registry,
		ch:       ch,
	}
}

func (*sshTunnelHandler) ContentType() int32 {
	return int32(mgmt_pb.ContentType_SshTunnelRequestType)
}

func (handler *sshTunnelHandler) HandleReceive(msg *channel.Message, ch channel.Channel) {
	log := pfxlog.ContextLogger(ch.Label()).Entry

	request := &mgmt_pb.SshTunnelRequest{}
	if err := proto.Unmarshal(msg.Body, request); err != nil {
		log.WithError(err).Error("unable to unmarshall ssh tunnel request")
		return
	}

	if request.DestinationType.CheckControllers() {
		if request.Destination == handler.network.GetAppId() {
			handler.tunnelToLocalhost(msg)
		}
		if request.DestinationType == mgmt_pb.DestinationType_Controller {
			handler.respondError(msg, fmt.Sprintf("no controllers found with id '%s'", request.Destination))
			return
		}
	}

	if request.DestinationType.CheckRouters() {
		r := handler.network.GetConnectedRouter(request.Destination)
		if r != nil {
			handler.tunnelToRouter(msg, request, r)
			return
		}
		if request.DestinationType == mgmt_pb.DestinationType_Router {
			r, _ = handler.network.GetRouter(request.Destination)
			if r == nil {
				handler.respondError(msg, fmt.Sprintf("no router found with id '%s'", request.Destination))
			} else {
				handler.respondError(msg, fmt.Sprintf("router '%s' not connected to controller", request.Destination))
			}
			return
		}
	}

	handler.respondError(msg, fmt.Sprintf("no destination found with with id '%s'", request.Destination))
}

func (handler *sshTunnelHandler) tunnelToRouter(msg *channel.Message, mgmtReq *mgmt_pb.SshTunnelRequest, r *network.Router) {
	tunnel := &routerSshTunnel{
		ch: handler.ch,
		r:  r,
	}

	handler.tunnel = tunnel
	tunnel.id = handler.registry.Register(tunnel)

	log := pfxlog.ContextLogger(handler.ch.Label()).WithField("connId", tunnel.id)

	req := &ctrl_pb.SshTunnelRequest{
		Destination:   r.Id,
		TimeoutMillis: mgmtReq.TimeoutMillis,
		ConnId:        tunnel.id,
	}

	envelope := protobufs.MarshalTyped(req).WithTimeout(time.Duration(mgmtReq.TimeoutMillis) * time.Millisecond)

	routerResp := &ctrl_pb.SshTunnelResponse{}
	err := protobufs.TypedResponse(routerResp).Unmarshall(envelope.SendForReply(r.Control))
	if err != nil {
		handler.respondError(msg, fmt.Sprintf("router error: %s", err.Error()))
		return
	}

	response := &mgmt_pb.SshTunnelResponse{
		Success: true,
		ConnId:  tunnel.id,
	}

	if sendErr := protobufs.MarshalTyped(response).ReplyTo(msg).WithTimeout(5 * time.Second).SendAndWaitForWire(handler.ch); sendErr != nil {
		log.WithError(sendErr).Error("unable to send ssh tunnel response for successful tunnel")
		tunnel.Close(sendErr)
		return
	}

	log.Info("started ssh tunnel")
}

func (handler *sshTunnelHandler) tunnelToLocalhost(msg *channel.Message) {
	log := pfxlog.ContextLogger(handler.ch.Label()).Entry

	conn, err := net.Dial("tcp", "localhost:22")
	if err != nil {
		log.WithError(err).Error("failed to dial ssh")
		handler.respondError(msg, err.Error())
		return
	}

	tunnel := &localSshTunnel{
		conn: conn,
		ch:   handler.ch,
	}

	handler.tunnel = tunnel
	tunnel.id = handler.registry.Register(tunnel)
	log = log.WithField("connId", handler.tunnel.Id())
	log.Info("registered ssh tunnel connection")

	response := &mgmt_pb.SshTunnelResponse{
		Success: true,
		ConnId:  tunnel.id,
	}

	if sendErr := protobufs.MarshalTyped(response).ReplyTo(msg).WithTimeout(5 * time.Second).SendAndWaitForWire(handler.ch); sendErr != nil {
		log.WithError(sendErr).Error("unable to send ssh tunnel response for successful tunnel")
		tunnel.Close(sendErr)
		return
	}

	log.Info("started ssh tunnel to local controller")

	go tunnel.readLoop()
}

func (handler *sshTunnelHandler) respondError(request *channel.Message, msg string) {
	response := &mgmt_pb.SshTunnelResponse{
		Success: false,
		Msg:     msg,
	}

	if sendErr := protobufs.MarshalTyped(response).ReplyTo(request).WithTimeout(5 * time.Second).SendAndWaitForWire(handler.ch); sendErr != nil {
		log := pfxlog.ContextLogger(handler.ch.Label()).Entry
		log.WithError(sendErr).Error("unable to send ssh tunnel response for failed tunnel")
	}
}

func (handler *sshTunnelHandler) HandleClose(channel.Channel) {
	if handler.tunnel != nil {
		handler.tunnel.Close(nil)
		handler.registry.Unregister(handler.tunnel.Id())
	}
}

type localSshTunnel struct {
	id     uint32
	conn   net.Conn
	ch     channel.Channel
	closed atomic.Bool
}

func (self *localSshTunnel) Id() uint32 {
	return self.id
}

func (self *localSshTunnel) WriteToServer(data []byte) error {
	_, err := self.conn.Write(data)
	return err
}

func (self *localSshTunnel) WriteToClient(data []byte) error {
	msg := channel.NewMessage(int32(mgmt_pb.ContentType_SshTunnelDataType), data)
	msg.PutUint32Header(int32(mgmt_pb.Header_SshTunnelConnIdHeader), self.id)
	return msg.WithTimeout(time.Second).SendAndWaitForWire(self.ch)
}

func (self *localSshTunnel) readLoop() {
	for {
		buf := make([]byte, 10240)
		n, err := self.conn.Read(buf)
		if err != nil {
			self.Close(err)
			return
		}
		buf = buf[:n]
		if err := self.WriteToClient(buf); err != nil {
			self.Close(err)
			return
		}
	}
}

func (self *localSshTunnel) Close(err error) {
	if self.closed.CompareAndSwap(false, true) {
		log := pfxlog.ContextLogger(self.ch.Label()).WithField("connId", self.id)

		log.WithError(err).Info("closing ssh tunnel connection")

		if closeErr := self.conn.Close(); closeErr != nil {
			log.WithError(closeErr).Error("failed closing ssh tunnel connection")
		}

		if !self.ch.IsClosed() && err != io.EOF && err != nil {
			msg := channel.NewMessage(int32(mgmt_pb.ContentType_SshTunnelCloseType), []byte(err.Error()))
			msg.PutUint32Header(int32(mgmt_pb.Header_SshTunnelConnIdHeader), self.id)
			if sendErr := self.ch.Send(msg); sendErr != nil {
				log.WithError(sendErr).Error("failed sending ssh tunnel close message")
			}
		}

		if closeErr := self.ch.Close(); closeErr != nil {
			log.WithError(closeErr).Error("failed closing ssh tunnel client channel")
		}
	}
}

type routerSshTunnel struct {
	id     uint32
	ch     channel.Channel
	r      *network.Router
	closed atomic.Bool
}

func (self *routerSshTunnel) Id() uint32 {
	return self.id
}

func (self *routerSshTunnel) WriteToServer(data []byte) error {
	msg := channel.NewMessage(int32(ctrl_pb.ContentType_SshTunnelDataType), data)
	msg.PutUint32Header(int32(ctrl_pb.Header_SshTunnelConnIdHeader), self.id)
	return msg.WithTimeout(time.Second).SendAndWaitForWire(self.r.Control)
}

func (self *routerSshTunnel) WriteToClient(data []byte) error {
	msg := channel.NewMessage(int32(mgmt_pb.ContentType_SshTunnelDataType), data)
	msg.PutUint32Header(int32(mgmt_pb.Header_SshTunnelConnIdHeader), self.id)
	return msg.WithTimeout(time.Second).SendAndWaitForWire(self.ch)
}

func (self *routerSshTunnel) Close(err error) {
	if self.closed.CompareAndSwap(false, true) {
		log := pfxlog.ContextLogger(self.ch.Label()).WithField("connId", self.id)

		log.WithError(err).Info("closing ssh tunnel connection")

		if !self.r.Control.IsClosed() {
			msg := channel.NewMessage(int32(ctrl_pb.ContentType_SshTunnelCloseType), func() []byte {
				if err != nil {
					return []byte(err.Error())
				}
				return []byte("closing")
			}())
			msg.PutUint32Header(int32(ctrl_pb.Header_SshTunnelConnIdHeader), self.id)
			if sendErr := self.ch.Send(msg); sendErr != nil {
				log.WithError(sendErr).Error("failed sending ssh tunnel close message")
			}
		}

		if !self.ch.IsClosed() && err != io.EOF && err != nil {
			msg := channel.NewMessage(int32(mgmt_pb.ContentType_SshTunnelCloseType), []byte(err.Error()))
			msg.PutUint32Header(int32(mgmt_pb.Header_SshTunnelConnIdHeader), self.id)
			if sendErr := self.ch.Send(msg); sendErr != nil {
				log.WithError(sendErr).Error("failed sending ssh tunnel close message")
			}
		}

		if closeErr := self.ch.Close(); closeErr != nil {
			log.WithError(closeErr).Error("failed closing ssh tunnel client channel")
		}
	}
}

func newSshTunnelDataHandler(registry *sshpipe.Registry) *sshTunnelDataHandler {
	return &sshTunnelDataHandler{
		registry: registry,
	}
}

type sshTunnelDataHandler struct {
	registry *sshpipe.Registry
}

func (*sshTunnelDataHandler) ContentType() int32 {
	return int32(mgmt_pb.ContentType_SshTunnelDataType)
}

func (handler *sshTunnelDataHandler) HandleReceive(msg *channel.Message, ch channel.Channel) {
	connId, _ := msg.GetUint32Header(int32(mgmt_pb.Header_SshTunnelConnIdHeader))
	tunnel := handler.registry.Get(connId)

	if tunnel == nil {
		pfxlog.ContextLogger(ch.Label()).
			WithField("connId", connId).
			Error("no ssh tunnel found for given connection id")

		go func() {
			errorMsg := fmt.Sprintf("invalid conn id '%v", connId)
			replyMsg := channel.NewMessage(int32(mgmt_pb.ContentType_SshTunnelCloseType), []byte(errorMsg))
			replyMsg.PutUint32Header(int32(mgmt_pb.Header_SshTunnelConnIdHeader), connId)
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

	if err := tunnel.WriteToServer(msg.Body); err != nil {
		tunnel.Close(err)
	}
}
