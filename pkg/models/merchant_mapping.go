// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package models

import (
	"net/http"
	"time"

	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"

	"xorm.io/xorm"
)

// MerchantMapping represents a mapping configuration for merchant fields
// swagger:model merchantMapping
type MerchantMapping struct {
	// The unique, numeric id of this mapping.
	// required: true
	// example: 1
	ID int64 `xorm:"bigint autoincr not null unique pk" json:"id" param:"merchantMapping"`
	// The field name this mapping applies to.
	// required: true
	// example: validTime
	FieldName string `xorm:"varchar(100) not null" json:"field_name" valid:"required,runelength(1|100)" minLength:"1" maxLength:"100"`
	// The placeholder text to be replaced.
	// required: true
	// example: A
	Placeholder string `xorm:"varchar(100) not null" json:"placeholder" valid:"required,runelength(1|100)" minLength:"1" maxLength:"100"`
	// The display text to replace the placeholder with.
	// required: true
	// example: 开门较晚，上午10点后拜访
	DisplayText string `xorm:"varchar(500) not null" json:"display_text" valid:"required,runelength(1|500)" minLength:"1" maxLength:"500"`
	// The label ID associated with this mapping.
	// required: true
	// example: 1000
	LabelID int64 `xorm:"bigint not null" json:"label_id"`
	// Whether this mapping is active.
	// required: false
	// example: true
	IsActive bool `xorm:"bool not null default true" json:"is_active"`

	// A timestamp when this mapping was created. You cannot change this value.
	// readOnly: true
	// example: 2023-01-01T00:00:00Z
	Created time.Time `xorm:"created not null" json:"created"`
	// A timestamp when this mapping was last updated. You cannot change this value.
	// readOnly: true
	// example: 2023-01-01T00:00:00Z
	Updated time.Time `xorm:"updated not null" json:"updated"`

	// The user who created this mapping.
	// readOnly: true
	Owner *user.User `xorm:"-" json:"owner" valid:"-"`

	// The user id of the user who created this mapping.
	// readOnly: true
	OwnerID int64 `xorm:"bigint not null INDEX" json:"-"`

	web.CRUDable `xorm:"-" json:"-"`
}

// TableName returns the table name for merchant mappings
func (mm *MerchantMapping) TableName() string {
	return "merchant_mappings"
}

// GetID returns the ID of the mapping
func (mm *MerchantMapping) GetID() int64 {
	return mm.ID
}

// Create creates a new merchant mapping
func (mm *MerchantMapping) Create(s *xorm.Session, auth web.Auth) (err error) {
	mm.OwnerID = auth.GetID()

	// Insert the mapping
	_, err = s.Insert(mm)
	if err != nil {
		return err
	}

	mm.Owner, _ = user.GetUserByID(s, mm.OwnerID)

	return
}

// ReadOne returns a single merchant mapping by its ID
func (mm *MerchantMapping) ReadOne(s *xorm.Session, auth web.Auth) (err error) {
	exists, err := s.Where("id = ? AND owner_id = ?", mm.ID, auth.GetID()).Get(mm)
	if err != nil {
		return err
	}
	if !exists {
		return ErrMerchantMappingDoesNotExist{MappingID: mm.ID}
	}

	mm.Owner, _ = user.GetUserByID(s, mm.OwnerID)

	return
}

// ReadAll returns all merchant mappings for a user
func (mm *MerchantMapping) ReadAll(s *xorm.Session, auth web.Auth, search string, page int, perPage int) (interface{}, int, int64, error) {
	query := s.Where("owner_id = ?", auth.GetID())

	// Enhanced search functionality - only apply if search is provided
	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.And(
			"(field_name LIKE ? OR placeholder LIKE ? OR display_text LIKE ?)",
			searchPattern, searchPattern, searchPattern,
		)
	}

	// Get total count before applying limit
	totalItems, err := query.Count(&MerchantMapping{})
	if err != nil {
		return nil, 0, 0, err
	}

	// Apply pagination only if page is not -1
	if page != -1 {
		limit, start := getLimitFromPageIndex(page, perPage)
		query = query.Limit(limit, start)
	}

	// Get mappings
	mappings := []*MerchantMapping{}
	err = query.Where("is_active = ?", true).OrderBy("field_name ASC, label_id ASC").Find(&mappings)
	if err != nil {
		return nil, 0, 0, err
	}

	// Load owners
	for _, mapping := range mappings {
		mapping.Owner, _ = user.GetUserByID(s, mapping.OwnerID)
	}

	numberOfTotalItems := len(mappings)

	return mappings, numberOfTotalItems, totalItems, nil
}

// Update updates a merchant mapping
func (mm *MerchantMapping) Update(s *xorm.Session, auth web.Auth) (err error) {
	// Check if we have at least an ID
	if mm.ID == 0 {
		return ErrMerchantMappingDoesNotExist{MappingID: mm.ID}
	}

	// Make sure the mapping exists
	exists, err := s.Where("id = ? AND owner_id = ?", mm.ID, auth.GetID()).Get(&MerchantMapping{})
	if err != nil {
		return
	}
	if !exists {
		return ErrMerchantMappingDoesNotExist{MappingID: mm.ID}
	}

	// Update the mapping
	_, err = s.ID(mm.ID).Update(mm)
	if err != nil {
		return err
	}

	mm.Owner, _ = user.GetUserByID(s, mm.OwnerID)

	return
}

// Delete deletes a merchant mapping
func (mm *MerchantMapping) Delete(s *xorm.Session, auth web.Auth) (err error) {
	// Check if the mapping exists
	exists, err := s.Where("id = ? AND owner_id = ?", mm.ID, auth.GetID()).Get(&MerchantMapping{})
	if err != nil {
		return
	}
	if !exists {
		return ErrMerchantMappingDoesNotExist{MappingID: mm.ID}
	}

	// Delete the mapping
	_, err = s.ID(mm.ID).Delete(&MerchantMapping{})
	return
}

// CanWrite checks if the user can write to this mapping
func (mm *MerchantMapping) CanWrite(s *xorm.Session, auth web.Auth) (bool, error) {
	return mm.checkPermission(s, auth.GetID(), PermissionWrite)
}

// CanRead checks if the user can read this mapping
func (mm *MerchantMapping) CanRead(s *xorm.Session, auth web.Auth) (bool, int, error) {
	canRead, err := mm.checkPermission(s, auth.GetID(), PermissionRead)
	return canRead, int(PermissionRead), err
}

// CanDelete checks if the user can delete this mapping
func (mm *MerchantMapping) CanDelete(s *xorm.Session, auth web.Auth) (bool, error) {
	return mm.checkPermission(s, auth.GetID(), PermissionAdmin)
}

// CanCreate checks if the user can create a mapping
func (mm *MerchantMapping) CanCreate(s *xorm.Session, auth web.Auth) (bool, error) {
	return true, nil
}

// CanUpdate checks if the user can update this mapping
func (mm *MerchantMapping) CanUpdate(s *xorm.Session, auth web.Auth) (bool, error) {
	return mm.checkPermission(s, auth.GetID(), PermissionWrite)
}

func (mm *MerchantMapping) checkPermission(s *xorm.Session, userID int64, permission Permission) (bool, error) {
	// Load the mapping if we don't have one
	if mm.ID != 0 {
		_, err := s.Where("id = ?", mm.ID).Get(mm)
		if err != nil {
			return false, err
		}
	}

	// Owner can do everything
	if mm.OwnerID == userID {
		return true, nil
	}

	return false, nil
}

// Error types
// ErrMerchantMappingDoesNotExist represents a "MerchantMappingDoesNotExist" kind of error.
type ErrMerchantMappingDoesNotExist struct {
	MappingID int64
}

// IsErrMerchantMappingDoesNotExist checks if an error is a ErrMerchantMappingDoesNotExist.
func IsErrMerchantMappingDoesNotExist(err error) bool {
	_, ok := err.(ErrMerchantMappingDoesNotExist)
	return ok
}

func (err ErrMerchantMappingDoesNotExist) Error() string {
	return "Merchant mapping does not exist."
}

// ErrCodeMerchantMappingDoesNotExist holds the unique world-error code of this error
const ErrCodeMerchantMappingDoesNotExist = 6020

// HTTPError holds the http error description
func (err ErrMerchantMappingDoesNotExist) HTTPError() web.HTTPError {
	return web.HTTPError{HTTPCode: http.StatusNotFound, Code: ErrCodeMerchantMappingDoesNotExist, Message: "This merchant mapping does not exist."}
}

// GetMerchantMappingsByField returns all active mappings for a specific field and user
func GetMerchantMappingsByField(s *xorm.Session, fieldName string, userID int64) ([]*MerchantMapping, error) {
	mappings := []*MerchantMapping{}
	err := s.Where("field_name = ? AND owner_id = ? AND is_active = ?", fieldName, userID, true).
		OrderBy("label_id ASC").
		Find(&mappings)
	return mappings, err
}

// GetAllMerchantMappingsByUser returns all active mappings for a user, grouped by field
func GetAllMerchantMappingsByUser(s *xorm.Session, userID int64) (map[string][]*MerchantMapping, error) {
	mappings := []*MerchantMapping{}
	err := s.Where("owner_id = ? AND is_active = ?", userID, true).
		OrderBy("field_name ASC, label_id ASC").
		Find(&mappings)
	if err != nil {
		return nil, err
	}

	// Group by field name
	result := make(map[string][]*MerchantMapping)
	for _, mapping := range mappings {
		result[mapping.FieldName] = append(result[mapping.FieldName], mapping)
	}

	return result, nil
}

// BulkCreateMerchantMappings creates multiple mappings at once
func BulkCreateMerchantMappings(s *xorm.Session, mappings []*MerchantMapping, auth web.Auth) error {
	for _, mapping := range mappings {
		mapping.OwnerID = auth.GetID()
	}

	_, err := s.Insert(mappings)
	return err
}

// BulkDeleteMerchantMappingsByField deletes all mappings for a specific field and user
func BulkDeleteMerchantMappingsByField(s *xorm.Session, fieldName string, userID int64) error {
	_, err := s.Where("field_name = ? AND owner_id = ?", fieldName, userID).Delete(&MerchantMapping{})
	return err
}
