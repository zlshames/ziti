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

package sshpipe

import (
	"github.com/openziti/foundation/v2/concurrenz"
	"sync"
	"sync/atomic"
)

type Pipe interface {
	Id() uint32
	WriteToServer(data []byte) error
	WriteToClient(data []byte) error
	Close(err error)
}

type Registry struct {
	lock        sync.Mutex
	nextId      atomic.Uint32
	connections concurrenz.CopyOnWriteMap[uint32, Pipe]
}

func (self *Registry) Register(pipe Pipe) uint32 {
	self.lock.Lock()
	defer self.lock.Unlock()

	for {
		nextId := self.nextId.Add(1)
		if val := self.connections.Get(nextId); val == nil {
			self.connections.Put(nextId, pipe)
			return nextId
		}
	}
}

func (self *Registry) Unregister(id uint32) {
	self.connections.Delete(id)
}

func (self *Registry) Get(id uint32) Pipe {
	return self.connections.Get(id)
}
