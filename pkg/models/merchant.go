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
	"strings"
	"time"

	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"

	"xorm.io/xorm"
)

// Merchant represents a merchant with business information
type Merchant struct {
	// The unique, numeric id of this merchant.
	ID int64 `xorm:"bigint autoincr not null unique pk" json:"id" param:"merchant"`
	// The title/name of the merchant.
	Title string `xorm:"varchar(250) not null" json:"title" valid:"required,runelength(1|250)" minLength:"1" maxLength:"250"`
	// The legal representative of the merchant.
	LegalRepresentative string `xorm:"varchar(250) null" json:"legal_representative" valid:"runelength(0|250)" maxLength:"250"`
	// The business address of the merchant.
	BusinessAddress string `xorm:"varchar(500) null" json:"business_address" valid:"runelength(0|500)" maxLength:"500"`
	// The business district of the merchant.
	BusinessDistrict string `xorm:"varchar(250) null" json:"business_district" valid:"runelength(0|250)" maxLength:"250"`
	// The valid time period for the merchant.
	ValidTime string `xorm:"varchar(250) null" json:"valid_time" valid:"runelength(0|250)" maxLength:"250"`
	// The traffic conditions description.
	TrafficConditions string `xorm:"longtext null" json:"traffic_conditions"`
	// Fixed events description.
	FixedEvents string `xorm:"longtext null" json:"fixed_events"`
	// Terminal type information.
	TerminalType string `xorm:"varchar(250) null" json:"terminal_type" valid:"runelength(0|250)" maxLength:"250"`
	// Special time periods description.
	SpecialTimePeriods string `xorm:"longtext null" json:"special_time_periods"`
	// Custom filter properties as JSON.
	CustomFilters string `xorm:"longtext null" json:"custom_filters"`

	// Geographic coordinates (longitude, latitude).
	Longitude float64 `xorm:"decimal(10,7) null" json:"longitude"`
	Latitude  float64 `xorm:"decimal(10,7) null" json:"latitude"`
	// Geocoding accuracy score (0.0 to 1.0).
	GeocodeAccuracy float64 `xorm:"decimal(3,2) default 0.0" json:"geocode_accuracy"`
	// Formatted address from geocoding service.
	GeocodeAddress string `xorm:"varchar(500) null" json:"geocode_address"`
	// Geocoding service used.
	GeocodeService string `xorm:"varchar(50) null" json:"geocode_service"`
	// Whether coordinates were manually set.
	IsManualLocation bool `xorm:"bool default false" json:"is_manual_location"`

	// A timestamp when this merchant was created. You cannot change this value.
	Created time.Time `xorm:"created not null" json:"created"`
	// A timestamp when this merchant was last updated. You cannot change this value.
	Updated time.Time `xorm:"updated not null" json:"updated"`

	// The user who created this merchant.
	Owner *user.User `xorm:"-" json:"owner" valid:"-"`

	// The user id of the user who created this merchant.
	OwnerID int64 `xorm:"bigint not null INDEX" json:"-"`

	// Associated tags (loaded separately).
	Tags []*MerchantTag `xorm:"-" json:"tags,omitempty"`
	// Primary geo point (loaded separately).
	PrimaryGeoPoint *GeoPoint `xorm:"-" json:"primary_geo_point,omitempty"`
	// All geo points (loaded separately).
	GeoPoints []*GeoPoint `xorm:"-" json:"geo_points,omitempty"`

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

	// Load associated tags
	err = m.LoadTags(s)
	if err != nil {
		return err
	}

	// Load primary geo point
	err = m.LoadPrimaryGeoPoint(s)
	if err != nil {
		return err
	}

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

	// Load owners, tags, geo points and apply label replacements
	for _, merchant := range merchants {
		merchant.Owner, _ = user.GetUserByID(s, merchant.OwnerID)
		
		// Load associated tags
		merchant.LoadTags(s)
		
		// Load primary geo point
		merchant.LoadPrimaryGeoPoint(s)
		
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

// FieldLabelMapping represents field-specific label mappings
type FieldLabelMapping struct {
	Field       string         `json:"field"`       // Field name (e.g., "validTime", "trafficConditions")
	Mappings    []LabelMapping `json:"mappings"`    // Array of placeholder to label mappings for this field
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

// LoadTags loads the associated tags for this merchant
func (m *Merchant) LoadTags(s *xorm.Session) error {
	tags, err := GetMerchantTagsByMerchantID(s, m.ID)
	if err != nil {
		return err
	}
	m.Tags = tags
	return nil
}

// LoadGeoPoints loads all geo points for this merchant
func (m *Merchant) LoadGeoPoints(s *xorm.Session) error {
	geoPoints, err := GetGeoPointsByMerchantID(s, m.ID)
	if err != nil {
		return err
	}
	m.GeoPoints = geoPoints
	
	// Set primary geo point
	for _, gp := range geoPoints {
		if gp.IsPrimary {
			m.PrimaryGeoPoint = gp
			break
		}
	}
	
	return nil
}

// LoadPrimaryGeoPoint loads only the primary geo point for this merchant
func (m *Merchant) LoadPrimaryGeoPoint(s *xorm.Session) error {
	geoPoint, err := GetPrimaryGeoPointByMerchantID(s, m.ID)
	if err != nil && !IsErrGeoPointDoesNotExist(err) {
		return err
	}
	if geoPoint != nil {
		m.PrimaryGeoPoint = geoPoint
	}
	return nil
}

// SetTags associates this merchant with the given tag IDs
func (m *Merchant) SetTags(s *xorm.Session, tagIDs []int64) error {
	return AssociateMerchantWithTags(s, m.ID, tagIDs)
}

// HasLocation checks if the merchant has valid coordinates
func (m *Merchant) HasLocation() bool {
	return m.Longitude != 0 && m.Latitude != 0 &&
		m.Longitude >= -180 && m.Longitude <= 180 &&
		m.Latitude >= -90 && m.Latitude <= 90
}

// GetDistance calculates the distance to another merchant in kilometers
func (m *Merchant) GetDistance(other *Merchant) float64 {
	if !m.HasLocation() || !other.HasLocation() {
		return 0
	}
	return CalculateDistance(m.Longitude, m.Latitude, other.Longitude, other.Latitude)
}

// UpdateLocation updates the merchant's location coordinates
func (m *Merchant) UpdateLocation(s *xorm.Session, lng, lat float64, accuracy float64, address, service string, isManual bool) error {
	m.Longitude = lng
	m.Latitude = lat
	m.GeocodeAccuracy = accuracy
	m.GeocodeAddress = address
	m.GeocodeService = service
	m.IsManualLocation = isManual

	// Update in database
	_, err := s.ID(m.ID).Cols(
		"longitude",
		"latitude", 
		"geocode_accuracy",
		"geocode_address",
		"geocode_service",
		"is_manual_location",
	).Update(m)
	
	if err != nil {
		return err
	}

	// Also create/update a geo point record
	geoPoint := &GeoPoint{
		MerchantID:       m.ID,
		Location:         WKBPoint{Lng: lng, Lat: lat},
		OriginalAddress:  m.BusinessAddress,
		FormattedAddress: address,
		AccuracyScore:    accuracy,
		GeocodingService: service,
		IsManual:         isManual,
		IsPrimary:        true,
	}

	// Check if primary geo point already exists
	existing, err := GetPrimaryGeoPointByMerchantID(s, m.ID)
	if err != nil && !IsErrGeoPointDoesNotExist(err) {
		return err
	}

	if existing != nil {
		// Update existing
		geoPoint.ID = existing.ID
		_, err = s.ID(existing.ID).Update(geoPoint)
	} else {
		// Create new
		_, err = s.Insert(geoPoint)
	}

	return err
}