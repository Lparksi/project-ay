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

package services

import (
	"fmt"
	"strings"

	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/web"
	"xorm.io/xorm"
)

// MerchantService provides business logic for merchant operations
type MerchantService struct {
	geocoder models.GeocodingService
}

// NewMerchantService creates a new merchant service
func NewMerchantService(geocoder models.GeocodingService) *MerchantService {
	if geocoder == nil {
		geocoder = models.NewMockGeocodingService()
	}
	return &MerchantService{
		geocoder: geocoder,
	}
}

// MerchantSearchOptions represents search and filter options for merchants
type MerchantSearchOptions struct {
	Search       string   `json:"search"`
	Tags         []int64  `json:"tags"`
	Categories   []string `json:"categories"`
	Districts    []string `json:"districts"`
	HasLocation  *bool    `json:"has_location"`
	CenterLng    float64  `json:"center_lng"`
	CenterLat    float64  `json:"center_lat"`
	RadiusKm     float64  `json:"radius_km"`
	MinAccuracy  float64  `json:"min_accuracy"`
	Page         int      `json:"page"`
	PerPage      int      `json:"per_page"`
	SortBy       string   `json:"sort_by"`
	SortOrder    string   `json:"sort_order"`
}

// MerchantSearchResult represents the result of a merchant search
type MerchantSearchResult struct {
	Merchants    []*models.Merchant `json:"merchants"`
	TotalCount   int64              `json:"total_count"`
	Page         int                `json:"page"`
	PerPage      int                `json:"per_page"`
	HasMore      bool               `json:"has_more"`
}

// SearchMerchants performs advanced search with filtering and sorting
func (ms *MerchantService) SearchMerchants(s *xorm.Session, auth web.Auth, options MerchantSearchOptions) (*MerchantSearchResult, error) {
	query := s.Where("owner_id = ?", auth.GetID())

	// Apply text search
	if options.Search != "" {
		searchTerm := "%" + strings.ToLower(options.Search) + "%"
		query = query.And("(LOWER(title) LIKE ? OR LOWER(legal_representative) LIKE ? OR LOWER(business_address) LIKE ? OR LOWER(business_district) LIKE ?)",
			searchTerm, searchTerm, searchTerm, searchTerm)
	}

	// Apply tag filters
	if len(options.Tags) > 0 {
		tagQuery := s.NewSession()
		tagQuery = tagQuery.
			Select("DISTINCT merchant_id").
			From("merchant_merchant_tags").
			Where("tag_id IN (?)", options.Tags)
		
		subQuery, args, err := tagQuery.ToSQL(&models.MerchantMerchantTag{})
		if err != nil {
			return nil, err
		}
		
		query = query.And(fmt.Sprintf("id IN (%s)", subQuery), args...)
	}

	// Apply district filters
	if len(options.Districts) > 0 {
		query = query.And("business_district IN (?)", options.Districts)
	}

	// Apply location filters
	if options.HasLocation != nil {
		if *options.HasLocation {
			query = query.And("longitude IS NOT NULL AND latitude IS NOT NULL")
		} else {
			query = query.And("(longitude IS NULL OR latitude IS NULL)")
		}
	}

	// Apply radius filter
	if options.CenterLng != 0 && options.CenterLat != 0 && options.RadiusKm > 0 {
		radiusMeters := options.RadiusKm * 1000
		query = query.And(`ST_DWithin(
			ST_SetSRID(ST_MakePoint(longitude, latitude), 4326)::geography,
			ST_SetSRID(ST_MakePoint(?, ?), 4326)::geography,
			?
		)`, options.CenterLng, options.CenterLat, radiusMeters)
	}

	// Apply accuracy filter
	if options.MinAccuracy > 0 {
		query = query.And("geocode_accuracy >= ?", options.MinAccuracy)
	}

	// Get total count before pagination
	totalCount, err := query.Count(&models.Merchant{})
	if err != nil {
		return nil, err
	}

	// Apply sorting
	orderBy := "created DESC" // Default sort
	if options.SortBy != "" {
		validSortFields := map[string]bool{
			"title":              true,
			"legal_representative": true,
			"business_district":   true,
			"created":            true,
			"updated":            true,
			"geocode_accuracy":   true,
		}
		
		if validSortFields[options.SortBy] {
			sortOrder := "ASC"
			if options.SortOrder == "desc" {
				sortOrder = "DESC"
			}
			orderBy = fmt.Sprintf("%s %s", options.SortBy, sortOrder)
		}
	}

	// Apply pagination
	if options.Page <= 0 {
		options.Page = 1
	}
	if options.PerPage <= 0 {
		options.PerPage = 20
	}
	
	limit := options.PerPage
	offset := (options.Page - 1) * options.PerPage

	// Execute query
	merchants := []*models.Merchant{}
	err = query.
		OrderBy(orderBy).
		Limit(limit, offset).
		Find(&merchants)
	if err != nil {
		return nil, err
	}

	// Load associated data for each merchant
	for _, merchant := range merchants {
		merchant.LoadTags(s)
		merchant.LoadPrimaryGeoPoint(s)
		merchant.ApplyLabelReplacements(s)
	}

	hasMore := int64(offset+len(merchants)) < totalCount

	return &MerchantSearchResult{
		Merchants:  merchants,
		TotalCount: totalCount,
		Page:       options.Page,
		PerPage:    options.PerPage,
		HasMore:    hasMore,
	}, nil
}

// CreateMerchantWithGeocoding creates a merchant and geocodes its address
func (ms *MerchantService) CreateMerchantWithGeocoding(s *xorm.Session, auth web.Auth, merchant *models.Merchant, tagIDs []int64) error {
	// Create the merchant first
	err := merchant.Create(s, auth)
	if err != nil {
		return err
	}

	// Associate tags if provided
	if len(tagIDs) > 0 {
		err = merchant.SetTags(s, tagIDs)
		if err != nil {
			return err
		}
	}

	// Geocode the address if provided
	if merchant.BusinessAddress != "" && !merchant.IsManualLocation {
		err = ms.geocodeMerchant(s, merchant)
		if err != nil {
			// Don't fail the creation if geocoding fails, just log it
			// The merchant can be geocoded later
		}
	}

	return nil
}

// UpdateMerchantWithGeocoding updates a merchant and re-geocodes if address changed
func (ms *MerchantService) UpdateMerchantWithGeocoding(s *xorm.Session, auth web.Auth, merchant *models.Merchant, tagIDs []int64) error {
	// Get the existing merchant to check if address changed
	existing := &models.Merchant{ID: merchant.ID}
	err := existing.ReadOne(s, auth)
	if err != nil {
		return err
	}

	addressChanged := existing.BusinessAddress != merchant.BusinessAddress

	// Update the merchant
	err = merchant.Update(s, auth)
	if err != nil {
		return err
	}

	// Update tags if provided
	if tagIDs != nil {
		err = merchant.SetTags(s, tagIDs)
		if err != nil {
			return err
		}
	}

	// Re-geocode if address changed and not manually set
	if addressChanged && merchant.BusinessAddress != "" && !merchant.IsManualLocation {
		err = ms.geocodeMerchant(s, merchant)
		if err != nil {
			// Don't fail the update if geocoding fails
		}
	}

	return nil
}

// BatchGeocodeMerchants geocodes multiple merchants
func (ms *MerchantService) BatchGeocodeMerchants(s *xorm.Session, auth web.Auth, merchantIDs []int64) error {
	if len(merchantIDs) == 0 {
		return nil
	}

	// Get merchants
	merchants := []*models.Merchant{}
	err := s.Where("owner_id = ?", auth.GetID()).In("id", merchantIDs).Find(&merchants)
	if err != nil {
		return err
	}

	// Filter merchants that need geocoding
	toGeocode := []*models.Merchant{}
	for _, merchant := range merchants {
		if merchant.BusinessAddress != "" && !merchant.IsManualLocation {
			toGeocode = append(toGeocode, merchant)
		}
	}

	if len(toGeocode) == 0 {
		return nil
	}

	// Batch geocode
	err = models.BatchGeocodeAndSave(toGeocode, ms.geocoder)
	if err != nil {
		return err
	}

	// Update merchants in database
	for _, merchant := range toGeocode {
		_, err = s.ID(merchant.ID).Cols(
			"longitude",
			"latitude",
			"geocode_accuracy",
			"geocode_address",
			"geocode_service",
		).Update(merchant)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetMerchantsByLocation finds merchants near a location
func (ms *MerchantService) GetMerchantsByLocation(s *xorm.Session, auth web.Auth, lng, lat, radiusKm float64, limit int) ([]*models.MerchantWithDistance, error) {
	return models.FindMerchantsWithinRadius(s, models.SpatialQueryOptions{
		CenterLng:       lng,
		CenterLat:       lat,
		RadiusKm:        radiusKm,
		Limit:           limit,
		IncludeDistance: true,
	})
}

// GetMerchantStatistics returns statistics about merchants
func (ms *MerchantService) GetMerchantStatistics(s *xorm.Session, auth web.Auth) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total merchants
	totalCount, err := s.Where("owner_id = ?", auth.GetID()).Count(&models.Merchant{})
	if err != nil {
		return nil, err
	}
	stats["total_merchants"] = totalCount

	// Merchants with location
	withLocationCount, err := s.Where("owner_id = ? AND longitude IS NOT NULL AND latitude IS NOT NULL", auth.GetID()).Count(&models.Merchant{})
	if err != nil {
		return nil, err
	}
	stats["merchants_with_location"] = withLocationCount

	// Merchants by district
	type DistrictCount struct {
		District string `json:"district"`
		Count    int    `json:"count"`
	}
	
	districtCounts := []DistrictCount{}
	err = s.
		Select("business_district as district, COUNT(*) as count").
		Where("owner_id = ? AND business_district != ''", auth.GetID()).
		GroupBy("business_district").
		OrderBy("count DESC").
		Limit(10).
		Find(&districtCounts)
	if err != nil {
		return nil, err
	}
	stats["top_districts"] = districtCounts

	// Geocoding accuracy distribution
	type AccuracyRange struct {
		Range string `json:"range"`
		Count int    `json:"count"`
	}
	
	accuracyRanges := []AccuracyRange{}
	err = s.SQL(`
		SELECT 
			CASE 
				WHEN geocode_accuracy >= 0.9 THEN 'High (0.9-1.0)'
				WHEN geocode_accuracy >= 0.7 THEN 'Medium (0.7-0.9)'
				WHEN geocode_accuracy >= 0.5 THEN 'Low (0.5-0.7)'
				ELSE 'Very Low (0-0.5)'
			END as range,
			COUNT(*) as count
		FROM merchants 
		WHERE owner_id = ? AND geocode_accuracy > 0
		GROUP BY range
		ORDER BY MIN(geocode_accuracy) DESC
	`, auth.GetID()).Find(&accuracyRanges)
	if err != nil {
		return nil, err
	}
	stats["accuracy_distribution"] = accuracyRanges

	return stats, nil
}

// ValidateMerchantData validates merchant data before creation/update
func (ms *MerchantService) ValidateMerchantData(merchant *models.Merchant) error {
	if strings.TrimSpace(merchant.Title) == "" {
		return fmt.Errorf("merchant title is required")
	}

	if merchant.HasLocation() {
		err := models.ValidateCoordinates(merchant.Longitude, merchant.Latitude)
		if err != nil {
			return fmt.Errorf("invalid coordinates: %w", err)
		}
	}

	return nil
}

// geocodeMerchant geocodes a single merchant
func (ms *MerchantService) geocodeMerchant(s *xorm.Session, merchant *models.Merchant) error {
	result, err := ms.geocoder.Geocode(merchant.BusinessAddress)
	if err != nil {
		return err
	}

	return merchant.UpdateLocation(s, 
		result.Longitude, 
		result.Latitude, 
		result.Accuracy, 
		result.FormattedAddress, 
		result.Service, 
		result.IsManual)
}

// ExportMerchants exports merchants to a structured format
func (ms *MerchantService) ExportMerchants(s *xorm.Session, auth web.Auth, options MerchantSearchOptions) ([]*models.Merchant, error) {
	// Remove pagination for export
	options.Page = 0
	options.PerPage = 0

	result, err := ms.SearchMerchants(s, auth, options)
	if err != nil {
		return nil, err
	}

	return result.Merchants, nil
}

// DuplicateMerchant creates a copy of an existing merchant
func (ms *MerchantService) DuplicateMerchant(s *xorm.Session, auth web.Auth, merchantID int64) (*models.Merchant, error) {
	// Get the original merchant
	original := &models.Merchant{ID: merchantID}
	err := original.ReadOne(s, auth)
	if err != nil {
		return nil, err
	}

	// Create a copy
	duplicate := &models.Merchant{
		Title:               original.Title + " (Copy)",
		LegalRepresentative: original.LegalRepresentative,
		BusinessAddress:     original.BusinessAddress,
		BusinessDistrict:    original.BusinessDistrict,
		ValidTime:           original.ValidTime,
		TrafficConditions:   original.TrafficConditions,
		FixedEvents:         original.FixedEvents,
		TerminalType:        original.TerminalType,
		SpecialTimePeriods:  original.SpecialTimePeriods,
		CustomFilters:       original.CustomFilters,
		Longitude:           original.Longitude,
		Latitude:            original.Latitude,
		GeocodeAccuracy:     original.GeocodeAccuracy,
		GeocodeAddress:      original.GeocodeAddress,
		GeocodeService:      original.GeocodeService,
		IsManualLocation:    original.IsManualLocation,
	}

	// Create the duplicate
	err = duplicate.Create(s, auth)
	if err != nil {
		return nil, err
	}

	// Copy tags
	if len(original.Tags) > 0 {
		tagIDs := make([]int64, len(original.Tags))
		for i, tag := range original.Tags {
			tagIDs[i] = tag.ID
		}
		err = duplicate.SetTags(s, tagIDs)
		if err != nil {
			return nil, err
		}
	}

	return duplicate, nil
}