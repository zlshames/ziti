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
	"github.com/openziti/foundation/v2/concurrenz"
	"github.com/openziti/foundation/v2/debugz"
	"github.com/openziti/ziti/common/pb/mgmt_pb"
	"github.com/openziti/ziti/controller/network"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type sshTunnelManager struct {
	lock        sync.Mutex
	nextId      atomic.Uint32
	connections concurrenz.CopyOnWriteMap[uint32, *sshTunnel]
}

func (self *sshTunnelManager) registerTunnel(tunnel *sshTunnel) {
	self.lock.Lock()
	defer self.lock.Unlock()

	for {
		nextId := self.nextId.Add(1)
		if val := self.connections.Get(nextId); val == nil {
			self.connections.Put(nextId, tunnel)
			tunnel.id = nextId
			return
		}
	}
}

type sshTunnelHandler struct {
	network   *network.Network
	tunnelMgr *sshTunnelManager
	tunnel    *sshTunnel
}

func newSshTunnelHandler(network *network.Network, tunnelManager *sshTunnelManager) *sshTunnelHandler {
	return &sshTunnelHandler{
		network:   network,
		tunnelMgr: tunnelManager,
	}
}

func (*sshTunnelHandler) ContentType() int32 {
	return int32(mgmt_pb.ContentType_SshTunnelRequestType)
}

func (handler *sshTunnelHandler) HandleReceive(msg *channel.Message, ch channel.Channel) {
	log := pfxlog.ContextLogger(ch.Label()).Entry

	conn, err := net.Dial("tcp", "localhost:22")
	if err != nil {
		log.WithError(err).Error("failed to dial ssh")
		response := &mgmt_pb.SshTunnelResponse{
			Success: false,
			ConnId:  0,
			Msg:     err.Error(),
		}
		if sendErr := protobufs.MarshalTyped(response).WithTimeout(5 * time.Second).Send(ch); sendErr != nil {
			log.WithError(sendErr).Error("unable to send ssh tunnel response for failed tunnel")
		}
		return
	}

	tunnel := &sshTunnel{
		conn: conn,
		ch:   ch,
	}
	handler.tunnel = tunnel
	handler.tunnelMgr.registerTunnel(tunnel)
	log = log.WithField("connId", handler.tunnel.id)
	log.Info("registered ssh tunnel connection")

	response := &mgmt_pb.SshTunnelResponse{
		Success: true,
		ConnId:  tunnel.id,
	}

	if sendErr := protobufs.MarshalTyped(response).ReplyTo(msg).WithTimeout(5 * time.Second).SendAndWaitForWire(ch); sendErr != nil {
		log.WithError(sendErr).Error("unable to send ssh tunnel response for successful tunnel")
		tunnel.close(sendErr)
		return
	}

	log.Info("started ssh tunnel")

	go tunnel.readLoop()
}

func (handler *sshTunnelHandler) HandleClose(channel.Channel) {
	debugz.DumpLocalStack()
	handler.tunnel.close(nil)
}

type sshTunnel struct {
	id   uint32
	conn net.Conn
	ch   channel.Channel
}

func (self *sshTunnel) readLoop() {
	for {
		buf := make([]byte, 10240)
		n, err := self.conn.Read(buf)
		if err != nil {
			self.close(err)
			return
		}
		buf = buf[:n]
		msg := channel.NewMessage(int32(mgmt_pb.ContentType_SshTunnelDataType), buf)
		msg.PutUint32Header(int32(mgmt_pb.Header_SshTunnelConnIdHeader), self.id)
		if err := self.ch.Send(msg); err != nil {
			self.close(err)
			return
		}
	}
}

func (self *sshTunnel) close(err error) {
	log := pfxlog.ContextLogger(self.ch.Label()).WithField("connId", self.id)

	log.WithError(err).Info("closing ssh tunnel connection")

	if closeErr := self.conn.Close(); closeErr != nil {
		log.WithError(closeErr).Error("failed closing ssh tunnel connection")
	}

	if err != io.EOF && err != nil {
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

func newSshTunnelDataHandler(tunnelManager *sshTunnelManager) *sshTunnelDataHandler {
	return &sshTunnelDataHandler{
		tunnelMgr: tunnelManager,
	}
}

type sshTunnelDataHandler struct {
	tunnelMgr *sshTunnelManager
}

func (*sshTunnelDataHandler) ContentType() int32 {
	return int32(mgmt_pb.ContentType_SshTunnelDataType)
}

func (handler *sshTunnelDataHandler) HandleReceive(msg *channel.Message, ch channel.Channel) {
	connId, _ := msg.GetUint32Header(int32(mgmt_pb.Header_SshTunnelConnIdHeader))
	tunnel := handler.tunnelMgr.connections.Get(connId)

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

	if _, err := tunnel.conn.Write(msg.Body); err != nil {
		tunnel.close(err)
	}
}
