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

package fabric

import (
	"errors"
	"fmt"
	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/channel/v2"
	"github.com/openziti/channel/v2/protobufs"
	"github.com/openziti/ziti/common/pb/mgmt_pb"
	"github.com/openziti/ziti/ziti/cmd/api"
	"github.com/openziti/ziti/ziti/cmd/common"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/term"
	"io"
	"net"
	"os"
	"os/user"
	"time"
)

type sshAction struct {
	api.Options

	incomingData chan []byte
	user         string
	keyPath      string
}

func NewSshCmd(p common.OptionsProvider) *cobra.Command {
	action := sshAction{
		Options: api.Options{
			CommonOptions: p(),
		},
		incomingData: make(chan []byte, 4),
	}

	sshCmd := &cobra.Command{
		Use:     "ssh",
		Short:   "ssh to ziti comopnents",
		Example: "ziti fabric ssh ctrl",
		Args:    cobra.ExactArgs(0),
		RunE:    action.ssh,
	}

	action.AddCommonFlags(sshCmd)
	sshCmd.Flags().StringVarP(&action.user, "user", "u", "", "SSH username")
	sshCmd.Flags().StringVarP(&action.keyPath, "key", "k", "", "SSH key path")
	return sshCmd
}

func (self *sshAction) ssh(cmd *cobra.Command, _ []string) error {
	closeNotify := make(chan struct{})

	bindHandler := func(binding channel.Binding) error {
		binding.AddReceiveHandlerF(int32(mgmt_pb.ContentType_SshTunnelDataType), self.receiveData)
		binding.AddCloseHandler(channel.CloseHandlerF(func(ch channel.Channel) {
			close(closeNotify)
		}))
		return nil
	}

	ch, err := api.NewWsMgmtChannel(channel.BindHandlerF(bindHandler))
	if err != nil {
		return err
	}

	sshRequest := &mgmt_pb.SshTunnelRequest{}
	resp, err := protobufs.MarshalTyped(sshRequest).WithTimeout(time.Duration(self.Timeout) * time.Second).SendForReply(ch)
	sshResp := &mgmt_pb.SshTunnelResponse{}
	err = protobufs.TypedResponse(sshResp).Unmarshall(resp, err)
	if err != nil {
		return err
	}

	if !sshResp.Success {
		return errors.New(sshResp.Msg)
	}

	conn := &sshConn{
		id:    sshResp.ConnId,
		ch:    ch,
		dataC: self.incomingData,
	}

	return self.remoteShell(conn)
}

func sshAuthMethodFromFile(keyPath string) (ssh.AuthMethod, error) {
	content, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("could not read ssh file [%s]: %w", keyPath, err)
	}

	if signer, err := ssh.ParsePrivateKey(content); err == nil {
		return ssh.PublicKeys(signer), nil
	} else {
		if err.Error() == "ssh: no key found" {
			return nil, fmt.Errorf("no private key found in [%s]: %w", keyPath, err)
		} else if err.(*ssh.PassphraseMissingError) != nil {
			return nil, fmt.Errorf("file is password protected [%s] %w", keyPath, err)
		} else {
			return nil, fmt.Errorf("error parsing private key from [%s]L %w", keyPath, err)
		}
	}
}

func sshAuthMethodAgent() ssh.AuthMethod {
	if sshAgent, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
		return ssh.PublicKeysCallback(agent.NewClient(sshAgent).Signers)
	}
	return nil
}

func (self *sshAction) newSshConfig() *ssh.ClientConfig {
	var methods []ssh.AuthMethod

	if fileMethod, err := sshAuthMethodFromFile(self.keyPath); err == nil {
		methods = append(methods, fileMethod)
	} else {
		logrus.Error(err)
	}

	if agentMethod := sshAuthMethodAgent(); agentMethod != nil {
		methods = append(methods, sshAuthMethodAgent())
	}

	return &ssh.ClientConfig{
		User:            self.user,
		Auth:            methods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
}

func (self *sshAction) remoteShell(conn net.Conn) error {
	if self.user == "" {
		current, err := user.Current()
		if err != nil {
			return fmt.Errorf("unable to get current user: %w", err)
		}
		self.user = current.Name
	}

	clientConfig := self.newSshConfig()
	c, chans, reqs, err := ssh.NewClientConn(conn, "localhost:22", clientConfig)
	if err != nil {
		return err
	}
	client := ssh.NewClient(c, chans, reqs)

	session, err := client.NewSession()
	if err != nil {
		return err
	}

	fd := int(os.Stdout.Fd())

	oldState, err := term.MakeRaw(fd)
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = session.Close()
		_ = term.Restore(fd, oldState)
	}()

	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	session.Stdin = os.Stdin

	termWidth, termHeight, err := term.GetSize(fd)
	if err != nil {
		panic(err)
	}

	if err := session.RequestPty("xterm", termHeight, termWidth, ssh.TerminalModes{ssh.ECHO: 1}); err != nil {
		return err
	}

	return session.Run("/bin/bash")
}

func (self *sshAction) receiveData(msg *channel.Message, _ channel.Channel) {
	self.incomingData <- msg.Body
}

type sshConn struct {
	id       uint32
	ch       channel.Channel
	leftover []byte
	dataC    chan []byte
}

func (self *sshConn) Read(b []byte) (n int, err error) {
	log := pfxlog.Logger().WithField("connId", self.id)
	if self.ch.IsClosed() {
		return 0, io.EOF
	}

	log.Tracef("read buffer = %d bytes", len(b))
	if len(self.leftover) > 0 {
		log.Tracef("found %d leftover bytes", len(self.leftover))
		n := copy(b, self.leftover)
		self.leftover = self.leftover[n:]
		return n, nil
	}

	for {
		d := <-self.dataC
		log.Tracef("got buffer from sequencer %d bytes", len(d))

		n := copy(b, d)
		self.leftover = d[n:]

		log.Tracef("saving %d bytes for leftover", len(self.leftover))
		log.Debugf("reading %v bytes", n)
		return n, nil
	}
}

func (self *sshConn) Write(b []byte) (n int, err error) {
	msg := channel.NewMessage(int32(mgmt_pb.ContentType_SshTunnelDataType), b)
	msg.PutUint32Header(int32(mgmt_pb.Header_SshTunnelConnIdHeader), self.id)
	if err = msg.WithTimeout(5 * time.Second).SendAndWaitForWire(self.ch); err != nil {
		return 0, err
	}
	return len(b), err
}

func (self *sshConn) Close() error {
	return self.ch.Close()
}

func (self *sshConn) LocalAddr() net.Addr {
	return self.ch.Underlay().GetLocalAddr()
}

func (self *sshConn) RemoteAddr() net.Addr {
	return sshAddr{
		destination: fmt.Sprintf("%v", self.id),
	}
}

func (self *sshConn) SetDeadline(t time.Time) error {
	//TODO implement me
	panic("implement me")
}

func (self *sshConn) SetReadDeadline(t time.Time) error {
	//TODO implement me
	panic("implement me")
}

func (self *sshConn) SetWriteDeadline(t time.Time) error {
	//TODO implement me
	panic("implement me")
}

type sshAddr struct {
	destination string
}

func (self sshAddr) Network() string {
	return "ziti"
}

func (self sshAddr) String() string {
	return "ziti:" + self.destination
}
