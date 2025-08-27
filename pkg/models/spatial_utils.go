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
	"fmt"
	"math"

	"xorm.io/xorm"
)

// SpatialQueryOptions represents options for spatial queries
type SpatialQueryOptions struct {
	// Center point for distance-based queries
	CenterLng float64
	CenterLat float64
	// Radius in kilometers for distance-based queries
	RadiusKm float64
	// Bounding box for area-based queries
	MinLng, MinLat, MaxLng, MaxLat float64
	// Maximum number of results
	Limit int
	// Include distance in results
	IncludeDistance bool
}

// MerchantWithDistance represents a merchant with calculated distance
type MerchantWithDistance struct {
	*Merchant
	Distance float64 `json:"distance,omitempty"` // Distance in kilometers
}

// FindMerchantsWithinRadius finds merchants within a specified radius from a center point
func FindMerchantsWithinRadius(s *xorm.Session, options SpatialQueryOptions) ([]*MerchantWithDistance, error) {
	if options.Limit <= 0 {
		options.Limit = 100 // Default limit
	}

	// Convert radius from kilometers to meters for PostGIS
	radiusMeters := options.RadiusKm * 1000

	query := s.NewSession()
	
	if options.IncludeDistance {
		// Include distance calculation in the query
		query = query.Select(`
			merchants.*,
			ST_Distance(
				ST_SetSRID(ST_MakePoint(merchants.longitude, merchants.latitude), 4326)::geography,
				ST_SetSRID(ST_MakePoint(?, ?), 4326)::geography
			) / 1000 as distance
		`, options.CenterLng, options.CenterLat)
	} else {
		query = query.Select("merchants.*")
	}

	// Use PostGIS functions for spatial filtering
	query = query.
		Where("merchants.longitude IS NOT NULL AND merchants.latitude IS NOT NULL").
		Where(`ST_DWithin(
			ST_SetSRID(ST_MakePoint(merchants.longitude, merchants.latitude), 4326)::geography,
			ST_SetSRID(ST_MakePoint(?, ?), 4326)::geography,
			?
		)`, options.CenterLng, options.CenterLat, radiusMeters).
		Limit(options.Limit)

	if options.IncludeDistance {
		query = query.OrderBy("distance ASC")
	}

	// Execute query
	results := []map[string]interface{}{}
	err := query.Find(&results)
	if err != nil {
		return nil, err
	}

	// Convert results to MerchantWithDistance
	merchants := make([]*MerchantWithDistance, 0, len(results))
	for _, result := range results {
		merchant := &Merchant{}
		
		// Map basic fields
		if id, ok := result["id"].(int64); ok {
			merchant.ID = id
		}
		if title, ok := result["title"].(string); ok {
			merchant.Title = title
		}
		if legalRep, ok := result["legal_representative"].(string); ok {
			merchant.LegalRepresentative = legalRep
		}
		if address, ok := result["business_address"].(string); ok {
			merchant.BusinessAddress = address
		}
		if district, ok := result["business_district"].(string); ok {
			merchant.BusinessDistrict = district
		}
		if lng, ok := result["longitude"].(float64); ok {
			merchant.Longitude = lng
		}
		if lat, ok := result["latitude"].(float64); ok {
			merchant.Latitude = lat
		}

		merchantWithDistance := &MerchantWithDistance{
			Merchant: merchant,
		}

		// Add distance if included
		if options.IncludeDistance {
			if distance, ok := result["distance"].(float64); ok {
				merchantWithDistance.Distance = distance
			}
		}

		merchants = append(merchants, merchantWithDistance)
	}

	return merchants, nil
}

// FindMerchantsInBoundingBox finds merchants within a bounding box
func FindMerchantsInBoundingBox(s *xorm.Session, options SpatialQueryOptions) ([]*Merchant, error) {
	if options.Limit <= 0 {
		options.Limit = 100
	}

	merchants := []*Merchant{}
	err := s.
		Where("longitude IS NOT NULL AND latitude IS NOT NULL").
		Where("longitude BETWEEN ? AND ?", options.MinLng, options.MaxLng).
		Where("latitude BETWEEN ? AND ?", options.MinLat, options.MaxLat).
		Limit(options.Limit).
		Find(&merchants)

	return merchants, err
}

// FindNearestMerchants finds the N nearest merchants to a point
func FindNearestMerchants(s *xorm.Session, lng, lat float64, limit int) ([]*MerchantWithDistance, error) {
	if limit <= 0 {
		limit = 10
	}

	// Use a reasonable search radius (e.g., 50km) to limit the search space
	return FindMerchantsWithinRadius(s, SpatialQueryOptions{
		CenterLng:       lng,
		CenterLat:       lat,
		RadiusKm:        50, // 50km radius
		Limit:           limit,
		IncludeDistance: true,
	})
}

// ValidateCoordinates checks if coordinates are valid
func ValidateCoordinates(lng, lat float64) error {
	if lng < -180 || lng > 180 {
		return fmt.Errorf("longitude must be between -180 and 180, got %f", lng)
	}
	if lat < -90 || lat > 90 {
		return fmt.Errorf("latitude must be between -90 and 90, got %f", lat)
	}
	return nil
}

// NormalizeCoordinates ensures coordinates are within valid ranges
func NormalizeCoordinates(lng, lat float64) (float64, float64) {
	// Normalize longitude to [-180, 180]
	for lng > 180 {
		lng -= 360
	}
	for lng < -180 {
		lng += 360
	}

	// Clamp latitude to [-90, 90]
	if lat > 90 {
		lat = 90
	}
	if lat < -90 {
		lat = -90
	}

	return lng, lat
}

// CreateBoundingBox creates a bounding box around a center point with a given radius
func CreateBoundingBox(centerLng, centerLat, radiusKm float64) (minLng, minLat, maxLng, maxLat float64) {
	// Approximate degrees per kilometer
	// This is a rough approximation and becomes less accurate at extreme latitudes
	latDegreesPerKm := 1.0 / 111.0
	lngDegreesPerKm := 1.0 / (111.0 * math.Cos(centerLat*math.Pi/180.0))

	latOffset := radiusKm * latDegreesPerKm
	lngOffset := radiusKm * lngDegreesPerKm

	minLng = centerLng - lngOffset
	maxLng = centerLng + lngOffset
	minLat = centerLat - latOffset
	maxLat = centerLat + latOffset

	// Ensure bounds are within valid coordinate ranges
	minLng, minLat = NormalizeCoordinates(minLng, minLat)
	maxLng, maxLat = NormalizeCoordinates(maxLng, maxLat)

	return minLng, minLat, maxLng, maxLat
}

// CalculateCenter calculates the center point of a set of coordinates
func CalculateCenter(coordinates [][]float64) (centerLng, centerLat float64) {
	if len(coordinates) == 0 {
		return 0, 0
	}

	var sumLng, sumLat float64
	for _, coord := range coordinates {
		if len(coord) >= 2 {
			sumLng += coord[0]
			sumLat += coord[1]
		}
	}

	count := float64(len(coordinates))
	return sumLng / count, sumLat / count
}

// IsPointInPolygon checks if a point is inside a polygon using ray casting algorithm
func IsPointInPolygon(lng, lat float64, polygon [][]float64) bool {
	if len(polygon) < 3 {
		return false
	}

	inside := false
	j := len(polygon) - 1

	for i := 0; i < len(polygon); i++ {
		if len(polygon[i]) < 2 || len(polygon[j]) < 2 {
			j = i
			continue
		}

		xi, yi := polygon[i][0], polygon[i][1]
		xj, yj := polygon[j][0], polygon[j][1]

		if ((yi > lat) != (yj > lat)) && (lng < (xj-xi)*(lat-yi)/(yj-yi)+xi) {
			inside = !inside
		}
		j = i
	}

	return inside
}

// GetMerchantsInPolygon finds merchants within a polygon area
func GetMerchantsInPolygon(s *xorm.Session, polygon [][]float64, limit int) ([]*Merchant, error) {
	if limit <= 0 {
		limit = 100
	}

	// First, get merchants in the bounding box of the polygon for efficiency
	minLng, minLat := polygon[0][0], polygon[0][1]
	maxLng, maxLat := minLng, minLat

	for _, coord := range polygon {
		if len(coord) >= 2 {
			if coord[0] < minLng {
				minLng = coord[0]
			}
			if coord[0] > maxLng {
				maxLng = coord[0]
			}
			if coord[1] < minLat {
				minLat = coord[1]
			}
			if coord[1] > maxLat {
				maxLat = coord[1]
			}
		}
	}

	// Get merchants in bounding box
	candidates, err := FindMerchantsInBoundingBox(s, SpatialQueryOptions{
		MinLng: minLng,
		MinLat: minLat,
		MaxLng: maxLng,
		MaxLat: maxLat,
		Limit:  limit * 2, // Get more candidates to filter
	})
	if err != nil {
		return nil, err
	}

	// Filter candidates that are actually inside the polygon
	result := make([]*Merchant, 0, len(candidates))
	for _, merchant := range candidates {
		if merchant.HasLocation() && IsPointInPolygon(merchant.Longitude, merchant.Latitude, polygon) {
			result = append(result, merchant)
			if len(result) >= limit {
				break
			}
		}
	}

	return result, nil
}

// SpatialIndex represents a simple spatial index for fast lookups
type SpatialIndex struct {
	GridSize float64                    // Size of each grid cell in degrees
	Grid     map[string][]*Merchant     // Grid cells containing merchants
}

// NewSpatialIndex creates a new spatial index
func NewSpatialIndex(gridSize float64) *SpatialIndex {
	return &SpatialIndex{
		GridSize: gridSize,
		Grid:     make(map[string][]*Merchant),
	}
}

// AddMerchant adds a merchant to the spatial index
func (si *SpatialIndex) AddMerchant(merchant *Merchant) {
	if !merchant.HasLocation() {
		return
	}

	key := si.getGridKey(merchant.Longitude, merchant.Latitude)
	si.Grid[key] = append(si.Grid[key], merchant)
}

// FindNearby finds merchants near a point using the spatial index
func (si *SpatialIndex) FindNearby(lng, lat, radiusKm float64) []*Merchant {
	// Calculate which grid cells to search
	cellsToSearch := si.getCellsInRadius(lng, lat, radiusKm)
	
	var candidates []*Merchant
	for _, key := range cellsToSearch {
		if merchants, exists := si.Grid[key]; exists {
			candidates = append(candidates, merchants...)
		}
	}

	// Filter by actual distance
	var result []*Merchant
	for _, merchant := range candidates {
		if merchant.HasLocation() {
			distance := CalculateDistance(lng, lat, merchant.Longitude, merchant.Latitude)
			if distance <= radiusKm {
				result = append(result, merchant)
			}
		}
	}

	return result
}

func (si *SpatialIndex) getGridKey(lng, lat float64) string {
	gridX := int(lng / si.GridSize)
	gridY := int(lat / si.GridSize)
	return fmt.Sprintf("%d,%d", gridX, gridY)
}

func (si *SpatialIndex) getCellsInRadius(lng, lat, radiusKm float64) []string {
	// Convert radius to degrees (approximate)
	radiusDegrees := radiusKm / 111.0 // Rough conversion

	minLng := lng - radiusDegrees
	maxLng := lng + radiusDegrees
	minLat := lat - radiusDegrees
	maxLat := lat + radiusDegrees

	var keys []string
	for gridLng := minLng; gridLng <= maxLng; gridLng += si.GridSize {
		for gridLat := minLat; gridLat <= maxLat; gridLat += si.GridSize {
			key := si.getGridKey(gridLng, gridLat)
			keys = append(keys, key)
		}
	}

	return keys
}