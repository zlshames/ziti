/*
	Copyright NetFoundry, Inc.

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

package model

import (
	"fmt"
	"github.com/openziti/edge/controller/persistence"
	"go.etcd.io/bbolt"
)

type PostureCheckMacAddresses struct {
	MacAddresses []string
}

func newPostureCheckMacAddresses() PostureCheckSubType {
	return &PostureCheckMacAddresses{}
}

func (p *PostureCheckMacAddresses) fillFrom(handler Handler, tx *bbolt.Tx, check *persistence.PostureCheck, subType persistence.PostureCheckSubType) error {
	subCheck := subType.(*persistence.PostureCheckMacAddresses)

	if subCheck == nil {
		return fmt.Errorf("could not covert mac address check to bolt mac address cheeck")
	}

	p.MacAddresses = subCheck.MacAddresses
	return nil
}

func (p *PostureCheckMacAddresses) toBoltEntityForCreate(tx *bbolt.Tx, handler Handler) (persistence.PostureCheckSubType, error) {
	return &persistence.PostureCheckMacAddresses{
		MacAddresses: p.MacAddresses,
	}, nil
}

func (p *PostureCheckMacAddresses) toBoltEntityForUpdate(tx *bbolt.Tx, handler Handler) (persistence.PostureCheckSubType, error) {
	return &persistence.PostureCheckMacAddresses{
		MacAddresses: p.MacAddresses,
	}, nil
}

func (p *PostureCheckMacAddresses) toBoltEntityForPatch(tx *bbolt.Tx, handler Handler) (persistence.PostureCheckSubType, error) {
	return &persistence.PostureCheckMacAddresses{
		MacAddresses: p.MacAddresses,
	}, nil
}
