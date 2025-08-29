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
	"encoding/json"
	"mime/multipart"
	"strings"
	"time"

	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"

	"xorm.io/xorm"
)

// Merchant represents a merchant with business information
// swagger:model merchant
type Merchant struct {
	// The unique, numeric id of this merchant.
	// required: true
	// example: 1
	ID int64 `xorm:"bigint autoincr not null unique pk" json:"id" param:"merchant"`
	// The title/name of the merchant.
	// required: true
	// example: ABC Store
	Title string `xorm:"varchar(250) not null" json:"title" valid:"required,runelength(1|250)" minLength:"1" maxLength:"250"`
	// The legal representative of the merchant.
	// example: John Doe
	LegalRepresentative string `xorm:"varchar(250) null" json:"legal_representative" valid:"runelength(0|250)" maxLength:"250"`
	// The business address of the merchant.
	// example: 123 Main St, City, State
	BusinessAddress string `xorm:"varchar(500) null" json:"business_address" valid:"runelength(0|500)" maxLength:"500"`
	// The business district of the merchant.
	// example: commercial
	BusinessDistrict string `xorm:"varchar(250) null" json:"business_district" valid:"runelength(0|250)" maxLength:"250"`
	// The valid time period for the merchant.
	// example: Morning
	ValidTime string `xorm:"varchar(250) null" json:"valid_time" valid:"runelength(0|250)" maxLength:"250"`
	// The traffic conditions description.
	// example: High traffic during lunch hours
	TrafficConditions string `xorm:"longtext null" json:"traffic_conditions"`
	// Fixed events description.
	// example: Closed on Sundays
	FixedEvents string `xorm:"longtext null" json:"fixed_events"`
	// Terminal type information.
	// example: Desktop
	TerminalType string `xorm:"varchar(250) null" json:"terminal_type" valid:"runelength(0|250)" maxLength:"250"`
	// Special time periods description.
	// example: Extra busy during holidays
	SpecialTimePeriods string `xorm:"longtext null" json:"special_time_periods"`
	// Custom filter properties as JSON.
	// example: {"validTime": "Morning", "trafficConditions": "High"}
	CustomFilters string `xorm:"longtext null" json:"custom_filters"`

	// A timestamp when this merchant was created. You cannot change this value.
	// readOnly: true
	// example: 2023-01-01T00:00:00Z
	Created time.Time `xorm:"created not null" json:"created"`
	// A timestamp when this merchant was last updated. You cannot change this value.
	// readOnly: true
	// example: 2023-01-01T00:00:00Z
	Updated time.Time `xorm:"updated not null" json:"updated"`

	// The user who created this merchant.
	// readOnly: true
	Owner *user.User `xorm:"-" json:"owner" valid:"-"`

	// The user id of the user who created this merchant.
	// readOnly: true
	OwnerID int64 `xorm:"bigint not null INDEX" json:"-"`

	web.CRUDable `xorm:"-" json:"-"`
}

// TableName returns the table name for merchants
func (m *Merchant) TableName() string {
	return "merchants"
}

// GetID returns the ID of the merchant
func (m *Merchant) GetID() int64 {
	return m.ID
}

// Create creates a new merchant
func (m *Merchant) Create(s *xorm.Session, auth web.Auth) (err error) {
	m.OwnerID = auth.GetID()

	// Insert the merchant
	_, err = s.Insert(m)
	if err != nil {
		return err
	}

	m.Owner, _ = user.GetUserByID(s, m.OwnerID)

	return
}

// ReadOne returns a single merchant by its ID
func (m *Merchant) ReadOne(s *xorm.Session, auth web.Auth) (err error) {
	exists, err := s.Where("id = ?", m.ID).Get(m)
	if err != nil {
		return err
	}
	if !exists {
		return ErrMerchantDoesNotExist{MerchantID: m.ID}
	}

	m.Owner, _ = user.GetUserByID(s, m.OwnerID)

	// Apply label replacements
	err = m.ApplyLabelReplacements(s)
	if err != nil {
		return err
	}

	return
}

// ReadAll returns all merchants for a user
func (m *Merchant) ReadAll(s *xorm.Session, auth web.Auth, search string, page int, perPage int) (interface{}, int, int64, error) {
	query := s.Where("owner_id = ?", auth.GetID())

	if search != "" {
		query = query.And("title LIKE ?", "%"+search+"%")
	}

	limit, start := getLimitFromPageIndex(page, perPage)

	// Get total count
	totalItems, err := query.Count(&Merchant{})
	if err != nil {
		return nil, 0, 0, err
	}

	// Get merchants
	merchants := []*Merchant{}
	err = query.Limit(limit, start).OrderBy("created DESC").Find(&merchants)
	if err != nil {
		return nil, 0, 0, err
	}

	// Load owners and apply label replacements
	for _, merchant := range merchants {
		merchant.Owner, _ = user.GetUserByID(s, merchant.OwnerID)
		// Apply label replacements
		merchant.ApplyLabelReplacements(s)
	}

	numberOfTotalItems := len(merchants)

	return merchants, numberOfTotalItems, totalItems, nil
}

// Update updates a merchant
func (m *Merchant) Update(s *xorm.Session, auth web.Auth) (err error) {
	// Check if we have at least an ID
	if m.ID == 0 {
		return ErrMerchantDoesNotExist{MerchantID: m.ID}
	}

	// Make sure the merchant exists
	exists, err := s.Where("id = ? AND owner_id = ?", m.ID, auth.GetID()).Get(&Merchant{})
	if err != nil {
		return
	}
	if !exists {
		return ErrMerchantDoesNotExist{MerchantID: m.ID}
	}

	// Update the merchant
	_, err = s.ID(m.ID).Update(m)
	if err != nil {
		return err
	}

	m.Owner, _ = user.GetUserByID(s, m.OwnerID)

	return
}

// Delete deletes a merchant
func (m *Merchant) Delete(s *xorm.Session, auth web.Auth) (err error) {
	// Check if the merchant exists
	exists, err := s.Where("id = ? AND owner_id = ?", m.ID, auth.GetID()).Get(&Merchant{})
	if err != nil {
		return
	}
	if !exists {
		return ErrMerchantDoesNotExist{MerchantID: m.ID}
	}

	// Delete the merchant
	_, err = s.ID(m.ID).Delete(&Merchant{})
	return
}

// CanWrite checks if the user can write to this merchant
func (m *Merchant) CanWrite(s *xorm.Session, auth web.Auth) (bool, error) {
	return m.checkPermission(s, auth.GetID(), PermissionWrite)
}

// CanRead checks if the user can read this merchant
func (m *Merchant) CanRead(s *xorm.Session, auth web.Auth) (bool, int, error) {
	canRead, err := m.checkPermission(s, auth.GetID(), PermissionRead)
	return canRead, int(PermissionRead), err
}

// CanDelete checks if the user can delete this merchant
func (m *Merchant) CanDelete(s *xorm.Session, auth web.Auth) (bool, error) {
	return m.checkPermission(s, auth.GetID(), PermissionAdmin)
}

// CanCreate checks if the user can create a merchant
func (m *Merchant) CanCreate(s *xorm.Session, auth web.Auth) (bool, error) {
	return true, nil
}

// CanUpdate checks if the user can update this merchant
func (m *Merchant) CanUpdate(s *xorm.Session, auth web.Auth) (bool, error) {
	return m.checkPermission(s, auth.GetID(), PermissionWrite)
}

// BusinessDistrict represents the business district of a merchant
// swagger:enum BusinessDistrict
type BusinessDistrict string

const (
	// BusinessDistrictResidential represents a residential area
	BusinessDistrictResidential BusinessDistrict = "residential"
	// BusinessDistrictCommercial represents a commercial (market) area
	BusinessDistrictCommercial BusinessDistrict = "commercial"
	// BusinessDistrictOther represents other areas
	BusinessDistrictOther BusinessDistrict = "other"
)

// FieldLabelMapping represents field-specific label mappings
type FieldLabelMapping struct {
	Field    string         `json:"field"`    // Field name (e.g., "validTime", "trafficConditions")
	Mappings []LabelMapping `json:"mappings"` // Array of placeholder to label mappings for this field
}

// LabelMapping represents a mapping between a placeholder and a label
type LabelMapping struct {
	Placeholder string `json:"placeholder"`
	LabelID     int64  `json:"labelId"`
}

// ApplyLabelReplacements applies field-specific label replacements to merchant fields
func (m *Merchant) ApplyLabelReplacements(s *xorm.Session) error {
	if m.CustomFilters == "" {
		return nil
	}

	// Try to parse as new field-specific format first
	var fieldMappings []FieldLabelMapping
	err := json.Unmarshal([]byte(m.CustomFilters), &fieldMappings)
	if err != nil {
		// If new format fails, try legacy format for backward compatibility
		var legacyMappings []LabelMapping
		err = json.Unmarshal([]byte(m.CustomFilters), &legacyMappings)
		if err != nil {
			// If both formats fail, just return without error
			return nil
		}
		// Apply legacy format (same mappings to all fields)
		m.applyLegacyLabelReplacements(s, legacyMappings)
		return nil
	}

	// Apply field-specific mappings
	for _, fieldMapping := range fieldMappings {
		// Build replacement map for this field
		replacements := make(map[string]string)
		for _, mapping := range fieldMapping.Mappings {
			if mapping.Placeholder == "" || mapping.LabelID == 0 {
				continue
			}

			// Get label by ID
			label := &Label{}
			exists, err := s.Where("id = ?", mapping.LabelID).Get(label)
			if err != nil || !exists {
				continue
			}

			replacements[mapping.Placeholder] = label.Title
		}

		// Apply replacements to the specific field
		switch fieldMapping.Field {
		case "validTime":
			m.ValidTime = applyReplacements(m.ValidTime, replacements)
		case "trafficConditions":
			m.TrafficConditions = applyReplacements(m.TrafficConditions, replacements)
		case "fixedEvents":
			m.FixedEvents = applyReplacements(m.FixedEvents, replacements)
		case "specialTimePeriods":
			m.SpecialTimePeriods = applyReplacements(m.SpecialTimePeriods, replacements)
		}
	}

	return nil
}

// applyLegacyLabelReplacements applies legacy format (same mappings to all fields)
func (m *Merchant) applyLegacyLabelReplacements(s *xorm.Session, mappings []LabelMapping) {
	// Build replacement map from placeholder to label title
	replacements := make(map[string]string)
	for _, mapping := range mappings {
		if mapping.Placeholder == "" || mapping.LabelID == 0 {
			continue
		}

		// Get label by ID
		label := &Label{}
		exists, err := s.Where("id = ?", mapping.LabelID).Get(label)
		if err != nil || !exists {
			continue
		}

		replacements[mapping.Placeholder] = label.Title
	}

	// Apply replacements to relevant fields (legacy behavior)
	m.ValidTime = applyReplacements(m.ValidTime, replacements)
	m.TrafficConditions = applyReplacements(m.TrafficConditions, replacements)
	m.FixedEvents = applyReplacements(m.FixedEvents, replacements)
	m.SpecialTimePeriods = applyReplacements(m.SpecialTimePeriods, replacements)
}

// applyReplacements applies a set of string replacements to a text
func applyReplacements(text string, replacements map[string]string) string {
	result := text
	for placeholder, replacement := range replacements {
		result = strings.ReplaceAll(result, placeholder, replacement)
	}
	return result
}

func (m *Merchant) checkPermission(s *xorm.Session, userID int64, permission Permission) (bool, error) {
	// Load the merchant if we don't have one
	if m.ID != 0 {
		_, err := s.Where("id = ?", m.ID).Get(m)
		if err != nil {
			return false, err
		}
	}

	// Owner can do everything
	if m.OwnerID == userID {
		return true, nil
	}

	return false, nil
}

// swagger:operation GET /merchants merchant listMerchants
// ---
// summary: Get all merchants
// description: Get all merchants for the current user
// parameters:
// - name: page
//   in: query
//   description: The page number. Used for pagination.
//   type: integer
//   default: 1
// - name: per_page
//   in: query
//   description: The number of items per page. Used for pagination.
//   type: integer
//   default: 50
// - name: s
//   in: query
//   description: Search merchants by title
//   type: string
// produces:
// - application/json
// responses:
//   "200":
//     description: The merchants
//     schema:
//       type: array
//       items:
//         $ref: '#/definitions/merchant'
//   "401":
//     description: Unauthorized
//     schema:
//       $ref: '#/definitions/HTTPError'
//   "500":
//     description: Internal error
//     schema:
//       $ref: '#/definitions/HTTPError'

// swagger:operation POST /merchants merchant createMerchant
// ---
// summary: Create a merchant
// description: Create a new merchant
// parameters:
// - name: merchant
//   in: body
//   description: The merchant to create
//   required: true
//   schema:
//     $ref: '#/definitions/merchant'
// produces:
// - application/json
// responses:
//   "200":
//     description: The created merchant
//     schema:
//       $ref: '#/definitions/merchant'
//   "400":
//     description: Bad request
//     schema:
//       $ref: '#/definitions/HTTPError'
//   "401":
//     description: Unauthorized
//     schema:
//       $ref: '#/definitions/HTTPError'
//   "500":
//     description: Internal error
//     schema:
//       $ref: '#/definitions/HTTPError'

// swagger:operation GET /merchants/{merchant} merchant getMerchant
// ---
// summary: Get a merchant
// description: Get a merchant by its ID
// parameters:
// - name: merchant
//   in: path
//   description: The merchant ID
//   required: true
//   type: integer
// produces:
// - application/json
// responses:
//   "200":
//     description: The merchant
//     schema:
//       $ref: '#/definitions/merchant'
//   "401":
//     description: Unauthorized
//     schema:
//       $ref: '#/definitions/HTTPError'
//   "404":
//     description: Merchant not found
//     schema:
//       $ref: '#/definitions/HTTPError'
//   "500":
//     description: Internal error
//     schema:
//       $ref: '#/definitions/HTTPError'

// swagger:operation POST /merchants/{merchant} merchant updateMerchant
// ---
// summary: Update a merchant
// description: Update a merchant by its ID
// parameters:
// - name: merchant
//   in: path
//   description: The merchant ID
//   required: true
//   type: integer
// - name: merchantUpdate
//   in: body
//   description: The merchant update
//   required: true
//   schema:
//     $ref: '#/definitions/merchant'
// produces:
// - application/json
// responses:
//   "200":
//     description: The updated merchant
//     schema:
//       $ref: '#/definitions/merchant'
//   "400":
//     description: Bad request
//     schema:
//       $ref: '#/definitions/HTTPError'
//   "401":
//     description: Unauthorized
//     schema:
//       $ref: '#/definitions/HTTPError'
//   "404":
//     description: Merchant not found
//     schema:
//       $ref: '#/definitions/HTTPError'
//   "500":
//     description: Internal error
//     schema:
//       $ref: '#/definitions/HTTPError'

// swagger:operation DELETE /merchants/{merchant} merchant deleteMerchant
// ---
// summary: Delete a merchant
// description: Delete a merchant by its ID
// parameters:
// - name: merchant
//   in: path
//   description: The merchant ID
//   required: true
//   type: integer
// produces:
// - application/json
// responses:
//   "200":
//     description: Merchant deleted
//     schema:
//       $ref: '#/definitions/HTTPSuccess'
//   "401":
//     description: Unauthorized
//     schema:
//       $ref: '#/definitions/HTTPError'
//   "404":
//     description: Merchant not found
//     schema:
//       $ref: '#/definitions/HTTPError'
//   "500":
//     description: Internal error
//     schema:
//       $ref: '#/definitions/HTTPError'

// swagger:parameters listMerchants
type listMerchantsParams struct {
	// The page number. Used for pagination.
	// in: query
	// required: false
	// default: 1
	Page int `json:"page"`
	// The number of items per page. Used for pagination.
	// in: query
	// required: false
	// default: 50
	PerPage int `json:"per_page"`
	// Search merchants by title
	// in: query
	// required: false
	Search string `json:"s"`
}

// swagger:parameters createMerchant
type createMerchantParams struct {
	// The merchant to create
	// in: body
	// required: true
	Body Merchant
}

// swagger:parameters getMerchant updateMerchant deleteMerchant
type merchantParams struct {
	// The merchant ID
	// in: path
	// required: true
	MerchantID int64 `json:"merchant"`
}

// swagger:parameters updateMerchant
type updateMerchantParams struct {
	// The merchant update
	// in: body
	// required: true
	Body Merchant
}

// swagger:parameters importMerchants
type importMerchantsParams struct {
	// XLSX file to import
	// in: formData
	// required: true
	File *multipart.FileHeader `json:"file"`
}

// swagger:response merchantList
type merchantListResponse struct {
	// in: body
	Body []Merchant
}

// swagger:response merchant
type merchantResponse struct {
	// in: body
	Body Merchant
}
