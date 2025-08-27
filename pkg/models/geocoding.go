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
	"fmt"
	"time"
)

// GeocodeResult represents the result of a geocoding operation
type GeocodeResult struct {
	// The geocoded coordinates
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
	
	// The formatted address returned by the service
	FormattedAddress string `json:"formatted_address"`
	
	// Accuracy/confidence score (0.0 to 1.0)
	Accuracy float64 `json:"accuracy"`
	
	// The service that provided this result
	Service string `json:"service"`
	
	// Additional metadata from the geocoding service
	Metadata map[string]interface{} `json:"metadata,omitempty"`
	
	// Whether this result was manually verified/corrected
	IsManual bool `json:"is_manual"`
}

// ReverseGeocodeResult represents the result of a reverse geocoding operation
type ReverseGeocodeResult struct {
	// The formatted address
	FormattedAddress string `json:"formatted_address"`
	
	// Address components
	Components map[string]string `json:"components,omitempty"`
	
	// The service that provided this result
	Service string `json:"service"`
	
	// Additional metadata
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// GeocodingService interface defines methods for geocoding operations
type GeocodingService interface {
	// Geocode converts an address to coordinates
	Geocode(address string) (*GeocodeResult, error)
	
	// ReverseGeocode converts coordinates to an address
	ReverseGeocode(lng, lat float64) (*ReverseGeocodeResult, error)
	
	// BatchGeocode processes multiple addresses at once
	BatchGeocode(addresses []string) ([]*GeocodeResult, error)
	
	// GetServiceName returns the name of the geocoding service
	GetServiceName() string
}

// MockGeocodingService provides a mock implementation for testing
type MockGeocodingService struct {
	ServiceName string
}

// NewMockGeocodingService creates a new mock geocoding service
func NewMockGeocodingService() *MockGeocodingService {
	return &MockGeocodingService{
		ServiceName: "mock",
	}
}

// Geocode implements a mock geocoding operation
func (m *MockGeocodingService) Geocode(address string) (*GeocodeResult, error) {
	// Simple mock: generate coordinates based on address hash
	hash := 0
	for _, char := range address {
		hash = hash*31 + int(char)
	}
	
	// Generate coordinates within Beijing area (rough bounds)
	lng := 116.0 + float64(hash%1000)/10000.0  // 116.0 to 116.1
	lat := 39.8 + float64((hash/1000)%1000)/10000.0  // 39.8 to 39.9
	
	return &GeocodeResult{
		Longitude:        lng,
		Latitude:         lat,
		FormattedAddress: fmt.Sprintf("Mock Address for: %s", address),
		Accuracy:         0.8,
		Service:          m.ServiceName,
		Metadata: map[string]interface{}{
			"mock": true,
			"hash": hash,
		},
		IsManual: false,
	}, nil
}

// ReverseGeocode implements a mock reverse geocoding operation
func (m *MockGeocodingService) ReverseGeocode(lng, lat float64) (*ReverseGeocodeResult, error) {
	return &ReverseGeocodeResult{
		FormattedAddress: fmt.Sprintf("Mock Address at %.6f, %.6f", lng, lat),
		Components: map[string]string{
			"country":  "中国",
			"province": "北京市",
			"city":     "北京市",
			"district": "朝阳区",
		},
		Service: m.ServiceName,
		Metadata: map[string]interface{}{
			"mock": true,
		},
	}, nil
}

// BatchGeocode implements mock batch geocoding
func (m *MockGeocodingService) BatchGeocode(addresses []string) ([]*GeocodeResult, error) {
	results := make([]*GeocodeResult, len(addresses))
	for i, address := range addresses {
		result, err := m.Geocode(address)
		if err != nil {
			return nil, err
		}
		results[i] = result
	}
	return results, nil
}

// GetServiceName returns the service name
func (m *MockGeocodingService) GetServiceName() string {
	return m.ServiceName
}

// GeocodingManager manages multiple geocoding services with fallback
type GeocodingManager struct {
	services []GeocodingService
	cache    map[string]*GeocodeResult
}

// NewGeocodingManager creates a new geocoding manager
func NewGeocodingManager() *GeocodingManager {
	return &GeocodingManager{
		services: make([]GeocodingService, 0),
		cache:    make(map[string]*GeocodeResult),
	}
}

// AddService adds a geocoding service to the manager
func (gm *GeocodingManager) AddService(service GeocodingService) {
	gm.services = append(gm.services, service)
}

// Geocode tries geocoding with each service until one succeeds
func (gm *GeocodingManager) Geocode(address string) (*GeocodeResult, error) {
	// Check cache first
	if cached, exists := gm.cache[address]; exists {
		return cached, nil
	}
	
	var lastErr error
	for _, service := range gm.services {
		result, err := service.Geocode(address)
		if err != nil {
			lastErr = err
			continue
		}
		
		// Cache successful result
		gm.cache[address] = result
		return result, nil
	}
	
	if lastErr != nil {
		return nil, fmt.Errorf("all geocoding services failed, last error: %w", lastErr)
	}
	
	return nil, fmt.Errorf("no geocoding services available")
}

// ReverseGeocode tries reverse geocoding with each service until one succeeds
func (gm *GeocodingManager) ReverseGeocode(lng, lat float64) (*ReverseGeocodeResult, error) {
	var lastErr error
	for _, service := range gm.services {
		result, err := service.ReverseGeocode(lng, lat)
		if err != nil {
			lastErr = err
			continue
		}
		return result, nil
	}
	
	if lastErr != nil {
		return nil, fmt.Errorf("all reverse geocoding services failed, last error: %w", lastErr)
	}
	
	return nil, fmt.Errorf("no reverse geocoding services available")
}

// BatchGeocode processes multiple addresses with retry logic
func (gm *GeocodingManager) BatchGeocode(addresses []string) ([]*GeocodeResult, error) {
	results := make([]*GeocodeResult, len(addresses))
	
	for i, address := range addresses {
		result, err := gm.Geocode(address)
		if err != nil {
			// For batch operations, we might want to continue with other addresses
			// and return partial results
			results[i] = &GeocodeResult{
				FormattedAddress: address,
				Accuracy:         0.0,
				Service:          "failed",
				Metadata: map[string]interface{}{
					"error": err.Error(),
				},
			}
		} else {
			results[i] = result
		}
		
		// Add small delay to avoid rate limiting
		time.Sleep(100 * time.Millisecond)
	}
	
	return results, nil
}

// ClearCache clears the geocoding cache
func (gm *GeocodingManager) ClearCache() {
	gm.cache = make(map[string]*GeocodeResult)
}

// GetCacheSize returns the number of cached results
func (gm *GeocodingManager) GetCacheSize() int {
	return len(gm.cache)
}

// GeocodeAndSave geocodes an address and saves the result to a merchant
func GeocodeAndSave(merchant *Merchant, geocoder GeocodingService) error {
	if merchant.BusinessAddress == "" {
		return fmt.Errorf("merchant has no business address to geocode")
	}
	
	result, err := geocoder.Geocode(merchant.BusinessAddress)
	if err != nil {
		return fmt.Errorf("geocoding failed: %w", err)
	}
	
	// Validate coordinates
	err = ValidateCoordinates(result.Longitude, result.Latitude)
	if err != nil {
		return fmt.Errorf("invalid coordinates from geocoding: %w", err)
	}
	
	// Update merchant with geocoding results
	merchant.Longitude = result.Longitude
	merchant.Latitude = result.Latitude
	merchant.GeocodeAccuracy = result.Accuracy
	merchant.GeocodeAddress = result.FormattedAddress
	merchant.GeocodeService = result.Service
	merchant.IsManualLocation = result.IsManual
	
	return nil
}

// BatchGeocodeAndSave geocodes multiple merchants
func BatchGeocodeAndSave(merchants []*Merchant, geocoder GeocodingService) error {
	addresses := make([]string, len(merchants))
	for i, merchant := range merchants {
		addresses[i] = merchant.BusinessAddress
	}
	
	results, err := geocoder.BatchGeocode(addresses)
	if err != nil {
		return fmt.Errorf("batch geocoding failed: %w", err)
	}
	
	if len(results) != len(merchants) {
		return fmt.Errorf("geocoding results count mismatch: expected %d, got %d", len(merchants), len(results))
	}
	
	for i, result := range results {
		if result.Accuracy > 0 { // Only update if geocoding was successful
			err = ValidateCoordinates(result.Longitude, result.Latitude)
			if err != nil {
				continue // Skip invalid coordinates
			}
			
			merchants[i].Longitude = result.Longitude
			merchants[i].Latitude = result.Latitude
			merchants[i].GeocodeAccuracy = result.Accuracy
			merchants[i].GeocodeAddress = result.FormattedAddress
			merchants[i].GeocodeService = result.Service
			merchants[i].IsManualLocation = result.IsManual
		}
	}
	
	return nil
}

// SerializeMetadata converts metadata to JSON string
func SerializeMetadata(metadata map[string]interface{}) string {
	if metadata == nil {
		return ""
	}
	
	data, err := json.Marshal(metadata)
	if err != nil {
		return ""
	}
	
	return string(data)
}

// DeserializeMetadata converts JSON string to metadata map
func DeserializeMetadata(data string) map[string]interface{} {
	if data == "" {
		return nil
	}
	
	var metadata map[string]interface{}
	err := json.Unmarshal([]byte(data), &metadata)
	if err != nil {
		return nil
	}
	
	return metadata
}