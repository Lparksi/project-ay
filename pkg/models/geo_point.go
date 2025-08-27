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
	"database/sql/driver"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
	"time"

	"code.vikunja.io/api/pkg/web"
	"xorm.io/xorm"
)

// WKBPoint represents a PostGIS POINT geometry in Well-Known Binary format
type WKBPoint struct {
	Lng float64 `json:"lng"`
	Lat float64 `json:"lat"`
}

// Scan implements the sql.Scanner interface for reading from database
func (p *WKBPoint) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	var wkb []byte
	switch v := value.(type) {
	case []byte:
		wkb = v
	case string:
		var err error
		wkb, err = hex.DecodeString(v)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("cannot scan %T into WKBPoint", value)
	}

	if len(wkb) < 21 {
		return fmt.Errorf("invalid WKB data length: %d", len(wkb))
	}

	// Parse WKB format:
	// Byte order (1 byte) + Geometry type (4 bytes) + X (8 bytes) + Y (8 bytes)
	byteOrder := wkb[0]
	
	var x, y float64
	if byteOrder == 1 { // Little endian
		x = math.Float64frombits(binary.LittleEndian.Uint64(wkb[5:13]))
		y = math.Float64frombits(binary.LittleEndian.Uint64(wkb[13:21]))
	} else { // Big endian
		x = math.Float64frombits(binary.BigEndian.Uint64(wkb[5:13]))
		y = math.Float64frombits(binary.BigEndian.Uint64(wkb[13:21]))
	}

	p.Lng = x
	p.Lat = y
	return nil
}

// Value implements the driver.Valuer interface for writing to database
func (p WKBPoint) Value() (driver.Value, error) {
	if p.Lng == 0 && p.Lat == 0 {
		return nil, nil
	}

	// Create WKB format: Little endian byte order + Point type + X + Y
	wkb := make([]byte, 21)
	
	// Byte order (little endian)
	wkb[0] = 1
	
	// Geometry type (Point = 1)
	binary.LittleEndian.PutUint32(wkb[1:5], 1)
	
	// X coordinate (longitude)
	binary.LittleEndian.PutUint64(wkb[5:13], math.Float64bits(p.Lng))
	
	// Y coordinate (latitude)
	binary.LittleEndian.PutUint64(wkb[13:21], math.Float64bits(p.Lat))
	
	return wkb, nil
}

// String returns a string representation of the point
func (p WKBPoint) String() string {
	return fmt.Sprintf("POINT(%f %f)", p.Lng, p.Lat)
}

// IsValid checks if the point has valid coordinates
func (p WKBPoint) IsValid() bool {
	return p.Lng >= -180 && p.Lng <= 180 && p.Lat >= -90 && p.Lat <= 90
}

// IsEmpty checks if the point is empty (0,0)
func (p WKBPoint) IsEmpty() bool {
	return p.Lng == 0 && p.Lat == 0
}

// GeoPoint represents a geographical point with metadata
type GeoPoint struct {
	// The unique, numeric id of this geo point.
	ID int64 `xorm:"bigint autoincr not null unique pk" json:"id"`
	
	// The merchant this geo point belongs to.
	MerchantID int64 `xorm:"bigint not null INDEX" json:"merchant_id"`
	
	// The geographical coordinates as PostGIS POINT.
	Location WKBPoint `xorm:"geometry(POINT,4326) null" json:"location"`
	
	// The original address used for geocoding.
	OriginalAddress string `xorm:"varchar(500) null" json:"original_address"`
	
	// The formatted address returned by the geocoding service.
	FormattedAddress string `xorm:"varchar(500) null" json:"formatted_address"`
	
	// The accuracy/confidence score of the geocoding (0.0 to 1.0).
	AccuracyScore float64 `xorm:"decimal(3,2) default 0.0" json:"accuracy_score"`
	
	// The geocoding service used (e.g., "amap", "google", "manual").
	GeocodingService string `xorm:"varchar(50) null" json:"geocoding_service"`
	
	// Whether this point was manually set by user.
	IsManual bool `xorm:"bool default false" json:"is_manual"`
	
	// Whether this is the primary/active location for the merchant.
	IsPrimary bool `xorm:"bool default true" json:"is_primary"`
	
	// Additional metadata from geocoding service (JSON format).
	Metadata string `xorm:"longtext null" json:"metadata"`
	
	// A timestamp when this geo point was created.
	Created time.Time `xorm:"created not null" json:"created"`
	// A timestamp when this geo point was last updated.
	Updated time.Time `xorm:"updated not null" json:"updated"`

	web.CRUDable `xorm:"-" json:"-"`
}

// TableName returns the table name for geo points
func (gp *GeoPoint) TableName() string {
	return "geo_points"
}

// GetID returns the ID of the geo point
func (gp *GeoPoint) GetID() int64 {
	return gp.ID
}

// Create creates a new geo point
func (gp *GeoPoint) Create(s *xorm.Session, auth web.Auth) (err error) {
	// If this is set as primary, unset other primary points for the same merchant
	if gp.IsPrimary {
		_, err = s.Where("merchant_id = ?", gp.MerchantID).Cols("is_primary").Update(&GeoPoint{IsPrimary: false})
		if err != nil {
			return err
		}
	}

	// Insert the geo point
	_, err = s.Insert(gp)
	return err
}

// ReadOne returns a single geo point by its ID
func (gp *GeoPoint) ReadOne(s *xorm.Session, auth web.Auth) (err error) {
	exists, err := s.Where("id = ?", gp.ID).Get(gp)
	if err != nil {
		return err
	}
	if !exists {
		return ErrGeoPointDoesNotExist{GeoPointID: gp.ID}
	}

	return nil
}

// ReadAll returns all geo points for a merchant
func (gp *GeoPoint) ReadAll(s *xorm.Session, auth web.Auth, search string, page int, perPage int) (interface{}, int, int64, error) {
	query := s.Where("merchant_id = ?", gp.MerchantID)

	if search != "" {
		query = query.And("(original_address LIKE ? OR formatted_address LIKE ?)", "%"+search+"%", "%"+search+"%")
	}

	limit, start := getLimitFromPageIndex(page, perPage)

	// Get total count
	totalItems, err := query.Count(&GeoPoint{})
	if err != nil {
		return nil, 0, 0, err
	}

	// Get geo points
	geoPoints := []*GeoPoint{}
	err = query.Limit(limit, start).OrderBy("is_primary DESC, created DESC").Find(&geoPoints)
	if err != nil {
		return nil, 0, 0, err
	}

	numberOfTotalItems := len(geoPoints)

	return geoPoints, numberOfTotalItems, totalItems, nil
}

// Update updates a geo point
func (gp *GeoPoint) Update(s *xorm.Session, auth web.Auth) (err error) {
	// Check if we have at least an ID
	if gp.ID == 0 {
		return ErrGeoPointDoesNotExist{GeoPointID: gp.ID}
	}

	// If this is set as primary, unset other primary points for the same merchant
	if gp.IsPrimary {
		_, err = s.Where("merchant_id = ? AND id != ?", gp.MerchantID, gp.ID).Cols("is_primary").Update(&GeoPoint{IsPrimary: false})
		if err != nil {
			return err
		}
	}

	// Update the geo point
	_, err = s.ID(gp.ID).Update(gp)
	return err
}

// Delete deletes a geo point
func (gp *GeoPoint) Delete(s *xorm.Session, auth web.Auth) (err error) {
	// Check if the geo point exists
	exists, err := s.Where("id = ?", gp.ID).Get(&GeoPoint{})
	if err != nil {
		return
	}
	if !exists {
		return ErrGeoPointDoesNotExist{GeoPointID: gp.ID}
	}

	// Delete the geo point
	_, err = s.ID(gp.ID).Delete(&GeoPoint{})
	return
}

// CanWrite checks if the user can write to this geo point
func (gp *GeoPoint) CanWrite(s *xorm.Session, auth web.Auth) (bool, error) {
	return gp.checkMerchantPermission(s, auth.GetID(), PermissionWrite)
}

// CanRead checks if the user can read this geo point
func (gp *GeoPoint) CanRead(s *xorm.Session, auth web.Auth) (bool, int, error) {
	canRead, err := gp.checkMerchantPermission(s, auth.GetID(), PermissionRead)
	return canRead, int(PermissionRead), err
}

// CanDelete checks if the user can delete this geo point
func (gp *GeoPoint) CanDelete(s *xorm.Session, auth web.Auth) (bool, error) {
	return gp.checkMerchantPermission(s, auth.GetID(), PermissionAdmin)
}

// CanCreate checks if the user can create a geo point
func (gp *GeoPoint) CanCreate(s *xorm.Session, auth web.Auth) (bool, error) {
	return gp.checkMerchantPermission(s, auth.GetID(), PermissionWrite)
}

// CanUpdate checks if the user can update this geo point
func (gp *GeoPoint) CanUpdate(s *xorm.Session, auth web.Auth) (bool, error) {
	return gp.checkMerchantPermission(s, auth.GetID(), PermissionWrite)
}

func (gp *GeoPoint) checkMerchantPermission(s *xorm.Session, userID int64, permission Permission) (bool, error) {
	// Load the geo point if we don't have one
	if gp.ID != 0 {
		_, err := s.Where("id = ?", gp.ID).Get(gp)
		if err != nil {
			return false, err
		}
	}

	// Check merchant permissions
	merchant := &Merchant{ID: gp.MerchantID}
	return merchant.checkPermission(s, userID, permission)
}

// GetPrimaryGeoPointByMerchantID returns the primary geo point for a merchant
func GetPrimaryGeoPointByMerchantID(s *xorm.Session, merchantID int64) (*GeoPoint, error) {
	geoPoint := &GeoPoint{}
	exists, err := s.Where("merchant_id = ? AND is_primary = ?", merchantID, true).Get(geoPoint)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrGeoPointDoesNotExist{MerchantID: merchantID}
	}
	return geoPoint, nil
}

// GetGeoPointsByMerchantID returns all geo points for a merchant
func GetGeoPointsByMerchantID(s *xorm.Session, merchantID int64) ([]*GeoPoint, error) {
	geoPoints := []*GeoPoint{}
	err := s.Where("merchant_id = ?", merchantID).OrderBy("is_primary DESC, created DESC").Find(&geoPoints)
	return geoPoints, err
}

// FindNearbyMerchants finds merchants within a certain distance from a point
func FindNearbyMerchants(s *xorm.Session, centerLng, centerLat, radiusKm float64, limit int) ([]*Merchant, error) {
	// Use PostGIS ST_DWithin function to find nearby points
	// ST_DWithin uses meters, so convert km to meters
	radiusMeters := radiusKm * 1000

	merchants := []*Merchant{}
	err := s.
		Join("INNER", "geo_points", "merchants.id = geo_points.merchant_id").
		Where("geo_points.is_primary = ?", true).
		Where("ST_DWithin(geo_points.location, ST_SetSRID(ST_MakePoint(?, ?), 4326), ?)", centerLng, centerLat, radiusMeters).
		Limit(limit).
		Find(&merchants)

	return merchants, err
}

// CalculateDistance calculates the distance between two points in kilometers
func CalculateDistance(lng1, lat1, lng2, lat2 float64) float64 {
	const earthRadius = 6371 // Earth's radius in kilometers

	// Convert degrees to radians
	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	deltaLat := (lat2 - lat1) * math.Pi / 180
	deltaLng := (lng2 - lng1) * math.Pi / 180

	// Haversine formula
	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLng/2)*math.Sin(deltaLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}