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

package model

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/golang-jwt/jwt/v5"
	"github.com/openziti/identity"
	"github.com/openziti/metrics"
	"github.com/openziti/ziti/common"
	"github.com/openziti/ziti/common/cert"
	"github.com/openziti/ziti/controller/config"
	"github.com/openziti/ziti/controller/db"
	"github.com/openziti/ziti/controller/jwtsigner"
	"github.com/openziti/ziti/controller/network"
	"github.com/xeipuuv/gojsonschema"
)

type Env interface {
	GetManagers() *Managers
	GetConfig() *config.Config
	GetDbProvider() network.DbProvider
	GetStores() *db.Stores
	GetAuthRegistry() AuthRegistry
	GetEnrollRegistry() EnrollmentRegistry
	GetApiClientCsrSigner() cert.Signer
	GetApiServerCsrSigner() cert.Signer
	GetControlClientCsrSigner() cert.Signer
	GetHostController() HostController
	IsEdgeRouterOnline(id string) bool
	GetMetricsRegistry() metrics.Registry
	GetFingerprintGenerator() cert.FingerprintGenerator
	HandleServiceUpdatedEventForIdentityId(identityId string)

	GetServerJwtSigner() jwtsigner.Signer
	GetServerCert() (*tls.Certificate, string, jwt.SigningMethod)
	JwtSignerKeyFunc(token *jwt.Token) (interface{}, error)
	GetPeerControllerAddresses() []string

	ValidateAccessToken(token string) (*common.AccessClaims, error)
	ValidateServiceAccessToken(token string, apiSessionId *string) (*common.ServiceAccessClaims, error)

	OidcIssuer() string
	RootIssuer() string
}

type HostController interface {
	GetNetwork() *network.Network
	Shutdown()
	GetCloseNotifyChannel() <-chan struct{}
	IsRaftEnabled() bool
	Identity() identity.Identity
	GetPeerSigners() []*x509.Certificate
	GetRaftIndex() uint64
}

type Schemas interface {
	GetEnrollErPost() *gojsonschema.Schema
	GetEnrollUpdbPost() *gojsonschema.Schema
}
