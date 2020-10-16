// Code generated by go-swagger; DO NOT EDIT.

//
// Copyright NetFoundry, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// __          __              _
// \ \        / /             (_)
//  \ \  /\  / /_ _ _ __ _ __  _ _ __   __ _
//   \ \/  \/ / _` | '__| '_ \| | '_ \ / _` |
//    \  /\  / (_| | |  | | | | | | | | (_| | : This file is generated, do not edit it.
//     \/  \/ \__,_|_|  |_| |_|_|_| |_|\__, |
//                                      __/ |
//                                     |___/

package rest_model

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/validate"
)

// PostureCheckDetail posture check detail
//
// swagger:discriminator PostureCheckDetail typeId
type PostureCheckDetail interface {
	runtime.Validatable

	// links
	// Required: true
	Links() Links
	SetLinks(Links)

	// created at
	// Required: true
	// Format: date-time
	CreatedAt() *strfmt.DateTime
	SetCreatedAt(*strfmt.DateTime)

	// id
	// Required: true
	ID() *string
	SetID(*string)

	// name
	// Required: true
	Name() *string
	SetName(*string)

	// role attributes
	// Required: true
	RoleAttributes() Attributes
	SetRoleAttributes(Attributes)

	// tags
	// Required: true
	Tags() Tags
	SetTags(Tags)

	// type Id
	// Required: true
	TypeID() string
	SetTypeID(string)

	// updated at
	// Required: true
	// Format: date-time
	UpdatedAt() *strfmt.DateTime
	SetUpdatedAt(*strfmt.DateTime)

	// version
	// Required: true
	Version() *int64
	SetVersion(*int64)

	// AdditionalProperties in base type shoud be handled just like regular properties
	// At this moment, the base type property is pushed down to the subtype
}

type postureCheckDetail struct {
	linksField Links

	createdAtField *strfmt.DateTime

	idField *string

	nameField *string

	roleAttributesField Attributes

	tagsField Tags

	typeIdField string

	updatedAtField *strfmt.DateTime

	versionField *int64
}

// Links gets the links of this polymorphic type
func (m *postureCheckDetail) Links() Links {
	return m.linksField
}

// SetLinks sets the links of this polymorphic type
func (m *postureCheckDetail) SetLinks(val Links) {
	m.linksField = val
}

// CreatedAt gets the created at of this polymorphic type
func (m *postureCheckDetail) CreatedAt() *strfmt.DateTime {
	return m.createdAtField
}

// SetCreatedAt sets the created at of this polymorphic type
func (m *postureCheckDetail) SetCreatedAt(val *strfmt.DateTime) {
	m.createdAtField = val
}

// ID gets the id of this polymorphic type
func (m *postureCheckDetail) ID() *string {
	return m.idField
}

// SetID sets the id of this polymorphic type
func (m *postureCheckDetail) SetID(val *string) {
	m.idField = val
}

// Name gets the name of this polymorphic type
func (m *postureCheckDetail) Name() *string {
	return m.nameField
}

// SetName sets the name of this polymorphic type
func (m *postureCheckDetail) SetName(val *string) {
	m.nameField = val
}

// RoleAttributes gets the role attributes of this polymorphic type
func (m *postureCheckDetail) RoleAttributes() Attributes {
	return m.roleAttributesField
}

// SetRoleAttributes sets the role attributes of this polymorphic type
func (m *postureCheckDetail) SetRoleAttributes(val Attributes) {
	m.roleAttributesField = val
}

// Tags gets the tags of this polymorphic type
func (m *postureCheckDetail) Tags() Tags {
	return m.tagsField
}

// SetTags sets the tags of this polymorphic type
func (m *postureCheckDetail) SetTags(val Tags) {
	m.tagsField = val
}

// TypeID gets the type Id of this polymorphic type
func (m *postureCheckDetail) TypeID() string {
	return "PostureCheckDetail"
}

// SetTypeID sets the type Id of this polymorphic type
func (m *postureCheckDetail) SetTypeID(val string) {
}

// UpdatedAt gets the updated at of this polymorphic type
func (m *postureCheckDetail) UpdatedAt() *strfmt.DateTime {
	return m.updatedAtField
}

// SetUpdatedAt sets the updated at of this polymorphic type
func (m *postureCheckDetail) SetUpdatedAt(val *strfmt.DateTime) {
	m.updatedAtField = val
}

// Version gets the version of this polymorphic type
func (m *postureCheckDetail) Version() *int64 {
	return m.versionField
}

// SetVersion sets the version of this polymorphic type
func (m *postureCheckDetail) SetVersion(val *int64) {
	m.versionField = val
}

// UnmarshalPostureCheckDetailSlice unmarshals polymorphic slices of PostureCheckDetail
func UnmarshalPostureCheckDetailSlice(reader io.Reader, consumer runtime.Consumer) ([]PostureCheckDetail, error) {
	var elements []json.RawMessage
	if err := consumer.Consume(reader, &elements); err != nil {
		return nil, err
	}

	var result []PostureCheckDetail
	for _, element := range elements {
		obj, err := unmarshalPostureCheckDetail(element, consumer)
		if err != nil {
			return nil, err
		}
		result = append(result, obj)
	}
	return result, nil
}

// UnmarshalPostureCheckDetail unmarshals polymorphic PostureCheckDetail
func UnmarshalPostureCheckDetail(reader io.Reader, consumer runtime.Consumer) (PostureCheckDetail, error) {
	// we need to read this twice, so first into a buffer
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	return unmarshalPostureCheckDetail(data, consumer)
}

func unmarshalPostureCheckDetail(data []byte, consumer runtime.Consumer) (PostureCheckDetail, error) {
	buf := bytes.NewBuffer(data)
	buf2 := bytes.NewBuffer(data)

	// the first time this is read is to fetch the value of the typeId property.
	var getType struct {
		TypeID string `json:"typeId"`
	}
	if err := consumer.Consume(buf, &getType); err != nil {
		return nil, err
	}

	if err := validate.RequiredString("typeId", "body", getType.TypeID); err != nil {
		return nil, err
	}

	// The value of typeId is used to determine which type to create and unmarshal the data into
	switch getType.TypeID {
	case "DOMAIN":
		var result PostureCheckDomainDetail
		if err := consumer.Consume(buf2, &result); err != nil {
			return nil, err
		}
		return &result, nil
	case "MAC":
		var result PostureCheckMacAddressDetail
		if err := consumer.Consume(buf2, &result); err != nil {
			return nil, err
		}
		return &result, nil
	case "OS":
		var result PostureCheckOperatingSystemDetail
		if err := consumer.Consume(buf2, &result); err != nil {
			return nil, err
		}
		return &result, nil
	case "PROCESS":
		var result PostureCheckProcessDetail
		if err := consumer.Consume(buf2, &result); err != nil {
			return nil, err
		}
		return &result, nil
	case "PostureCheckDetail":
		var result postureCheckDetail
		if err := consumer.Consume(buf2, &result); err != nil {
			return nil, err
		}
		return &result, nil
	}
	return nil, errors.New(422, "invalid typeId value: %q", getType.TypeID)
}

// Validate validates this posture check detail
func (m *postureCheckDetail) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateLinks(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateCreatedAt(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateID(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateName(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateRoleAttributes(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateTags(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateUpdatedAt(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateVersion(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *postureCheckDetail) validateLinks(formats strfmt.Registry) error {

	if err := m.Links().Validate(formats); err != nil {
		if ve, ok := err.(*errors.Validation); ok {
			return ve.ValidateName("_links")
		}
		return err
	}

	return nil
}

func (m *postureCheckDetail) validateCreatedAt(formats strfmt.Registry) error {

	if err := validate.Required("createdAt", "body", m.CreatedAt()); err != nil {
		return err
	}

	if err := validate.FormatOf("createdAt", "body", "date-time", m.CreatedAt().String(), formats); err != nil {
		return err
	}

	return nil
}

func (m *postureCheckDetail) validateID(formats strfmt.Registry) error {

	if err := validate.Required("id", "body", m.ID()); err != nil {
		return err
	}

	return nil
}

func (m *postureCheckDetail) validateName(formats strfmt.Registry) error {

	if err := validate.Required("name", "body", m.Name()); err != nil {
		return err
	}

	return nil
}

func (m *postureCheckDetail) validateRoleAttributes(formats strfmt.Registry) error {

	if err := validate.Required("roleAttributes", "body", m.RoleAttributes()); err != nil {
		return err
	}

	if err := m.RoleAttributes().Validate(formats); err != nil {
		if ve, ok := err.(*errors.Validation); ok {
			return ve.ValidateName("roleAttributes")
		}
		return err
	}

	return nil
}

func (m *postureCheckDetail) validateTags(formats strfmt.Registry) error {

	if err := m.Tags().Validate(formats); err != nil {
		if ve, ok := err.(*errors.Validation); ok {
			return ve.ValidateName("tags")
		}
		return err
	}

	return nil
}

func (m *postureCheckDetail) validateUpdatedAt(formats strfmt.Registry) error {

	if err := validate.Required("updatedAt", "body", m.UpdatedAt()); err != nil {
		return err
	}

	if err := validate.FormatOf("updatedAt", "body", "date-time", m.UpdatedAt().String(), formats); err != nil {
		return err
	}

	return nil
}

func (m *postureCheckDetail) validateVersion(formats strfmt.Registry) error {

	if err := validate.Required("version", "body", m.Version()); err != nil {
		return err
	}

	return nil
}
