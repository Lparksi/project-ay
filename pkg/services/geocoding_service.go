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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
)

// AmapGeocodingService implements geocoding using Amap (高德地图) API
type AmapGeocodingService struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
}

// NewAmapGeocodingService creates a new Amap geocoding service
func NewAmapGeocodingService(apiKey string) *AmapGeocodingService {
	return &AmapGeocodingService{
		APIKey:  apiKey,
		BaseURL: "https://restapi.amap.com/v3",
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// AmapGeocodeResponse represents the response from Amap geocoding API
type AmapGeocodeResponse struct {
	Status   string `json:"status"`
	Info     string `json:"info"`
	InfoCode string `json:"infocode"`
	Count    string `json:"count"`
	Geocodes []struct {
		FormattedAddress string `json:"formatted_address"`
		Country          string `json:"country"`
		Province         string `json:"province"`
		City             string `json:"city"`
		Citycode         string `json:"citycode"`
		District         string `json:"district"`
		Township         string `json:"township"`
		Neighborhood     struct {
			Name string `json:"name"`
			Type string `json:"type"`
		} `json:"neighborhood"`
		Building struct {
			Name string `json:"name"`
			Type string `json:"type"`
		} `json:"building"`
		Adcode   string `json:"adcode"`
		Street   string `json:"street"`
		Number   string `json:"number"`
		Location string `json:"location"`
		Level    string `json:"level"`
	} `json:"geocodes"`
}

// AmapReverseGeocodeResponse represents the response from Amap reverse geocoding API
type AmapReverseGeocodeResponse struct {
	Status     string `json:"status"`
	Info       string `json:"info"`
	InfoCode   string `json:"infocode"`
	Regeocode  struct {
		FormattedAddress string `json:"formatted_address"`
		AddressComponent struct {
			Country      string `json:"country"`
			Province     string `json:"province"`
			City         string `json:"city"`
			Citycode     string `json:"citycode"`
			District     string `json:"district"`
			Township     string `json:"township"`
			Adcode       string `json:"adcode"`
			Street       string `json:"street"`
			Number       string `json:"number"`
			Neighborhood struct {
				Name string `json:"name"`
				Type string `json:"type"`
			} `json:"neighborhood"`
			Building struct {
				Name string `json:"name"`
				Type string `json:"type"`
			} `json:"building"`
		} `json:"addressComponent"`
	} `json:"regeocode"`
}

// Geocode converts an address to coordinates using Amap API
func (ags *AmapGeocodingService) Geocode(address string) (*models.GeocodeResult, error) {
	if ags.APIKey == "" {
		return nil, fmt.Errorf("Amap API key not configured")
	}

	// Build request URL
	params := url.Values{}
	params.Set("key", ags.APIKey)
	params.Set("address", address)
	params.Set("output", "json")
	params.Set("batch", "false")

	requestURL := fmt.Sprintf("%s/geocode/geo?%s", ags.BaseURL, params.Encode())

	// Make HTTP request
	resp, err := ags.HTTPClient.Get(requestURL)
	if err != nil {
		return nil, fmt.Errorf("failed to make geocoding request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse JSON response
	var geocodeResp AmapGeocodeResponse
	err = json.Unmarshal(body, &geocodeResp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse geocoding response: %w", err)
	}

	// Check response status
	if geocodeResp.Status != "1" {
		return nil, fmt.Errorf("geocoding failed: %s (code: %s)", geocodeResp.Info, geocodeResp.InfoCode)
	}

	// Check if we have results
	if len(geocodeResp.Geocodes) == 0 {
		return nil, fmt.Errorf("no geocoding results found for address: %s", address)
	}

	// Parse the first result
	geocode := geocodeResp.Geocodes[0]
	
	// Parse coordinates from location string (format: "lng,lat")
	locationParts := strings.Split(geocode.Location, ",")
	if len(locationParts) != 2 {
		return nil, fmt.Errorf("invalid location format: %s", geocode.Location)
	}

	lng, err := strconv.ParseFloat(locationParts[0], 64)
	if err != nil {
		return nil, fmt.Errorf("invalid longitude: %s", locationParts[0])
	}

	lat, err := strconv.ParseFloat(locationParts[1], 64)
	if err != nil {
		return nil, fmt.Errorf("invalid latitude: %s", locationParts[1])
	}

	// Calculate accuracy based on level
	accuracy := ags.calculateAccuracy(geocode.Level)

	// Build metadata
	metadata := map[string]interface{}{
		"country":    geocode.Country,
		"province":   geocode.Province,
		"city":       geocode.City,
		"district":   geocode.District,
		"township":   geocode.Township,
		"street":     geocode.Street,
		"number":     geocode.Number,
		"adcode":     geocode.Adcode,
		"level":      geocode.Level,
		"citycode":   geocode.Citycode,
	}

	return &models.GeocodeResult{
		Longitude:        lng,
		Latitude:         lat,
		FormattedAddress: geocode.FormattedAddress,
		Accuracy:         accuracy,
		Service:          ags.GetServiceName(),
		Metadata:         metadata,
		IsManual:         false,
	}, nil
}

// ReverseGeocode converts coordinates to an address using Amap API
func (ags *AmapGeocodingService) ReverseGeocode(lng, lat float64) (*models.ReverseGeocodeResult, error) {
	if ags.APIKey == "" {
		return nil, fmt.Errorf("Amap API key not configured")
	}

	// Build request URL
	params := url.Values{}
	params.Set("key", ags.APIKey)
	params.Set("location", fmt.Sprintf("%.6f,%.6f", lng, lat))
	params.Set("output", "json")
	params.Set("extensions", "base")

	requestURL := fmt.Sprintf("%s/geocode/regeo?%s", ags.BaseURL, params.Encode())

	// Make HTTP request
	resp, err := ags.HTTPClient.Get(requestURL)
	if err != nil {
		return nil, fmt.Errorf("failed to make reverse geocoding request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse JSON response
	var reverseResp AmapReverseGeocodeResponse
	err = json.Unmarshal(body, &reverseResp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse reverse geocoding response: %w", err)
	}

	// Check response status
	if reverseResp.Status != "1" {
		return nil, fmt.Errorf("reverse geocoding failed: %s (code: %s)", reverseResp.Info, reverseResp.InfoCode)
	}

	// Build components
	components := map[string]string{
		"country":   reverseResp.Regeocode.AddressComponent.Country,
		"province":  reverseResp.Regeocode.AddressComponent.Province,
		"city":      reverseResp.Regeocode.AddressComponent.City,
		"district":  reverseResp.Regeocode.AddressComponent.District,
		"township":  reverseResp.Regeocode.AddressComponent.Township,
		"street":    reverseResp.Regeocode.AddressComponent.Street,
		"number":    reverseResp.Regeocode.AddressComponent.Number,
		"adcode":    reverseResp.Regeocode.AddressComponent.Adcode,
		"citycode":  reverseResp.Regeocode.AddressComponent.Citycode,
	}

	// Build metadata
	metadata := map[string]interface{}{
		"address_component": reverseResp.Regeocode.AddressComponent,
	}

	return &models.ReverseGeocodeResult{
		FormattedAddress: reverseResp.Regeocode.FormattedAddress,
		Components:       components,
		Service:          ags.GetServiceName(),
		Metadata:         metadata,
	}, nil
}

// BatchGeocode processes multiple addresses at once
func (ags *AmapGeocodingService) BatchGeocode(addresses []string) ([]*models.GeocodeResult, error) {
	results := make([]*models.GeocodeResult, len(addresses))
	
	for i, address := range addresses {
		result, err := ags.Geocode(address)
		if err != nil {
			// For batch operations, we continue with other addresses
			log.Errorf("Failed to geocode address '%s': %v", address, err)
			results[i] = &models.GeocodeResult{
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
		
		// Add delay to avoid rate limiting
		time.Sleep(200 * time.Millisecond)
	}
	
	return results, nil
}

// GetServiceName returns the service name
func (ags *AmapGeocodingService) GetServiceName() string {
	return "amap"
}

// calculateAccuracy converts Amap level to accuracy score
func (ags *AmapGeocodingService) calculateAccuracy(level string) float64 {
	switch level {
	case "道路":
		return 0.95
	case "门牌号":
		return 0.98
	case "兴趣点":
		return 0.90
	case "区县":
		return 0.70
	case "城市":
		return 0.50
	case "省份":
		return 0.30
	default:
		return 0.80
	}
}

// GeocodingServiceManager manages multiple geocoding services with configuration
type GeocodingServiceManager struct {
	*models.GeocodingManager
	config map[string]interface{}
}

// NewGeocodingServiceManager creates a new geocoding service manager
func NewGeocodingServiceManager() *GeocodingServiceManager {
	manager := &GeocodingServiceManager{
		GeocodingManager: models.NewGeocodingManager(),
		config:           make(map[string]interface{}),
	}

	// Initialize services based on configuration
	manager.initializeServices()

	return manager
}

// initializeServices initializes geocoding services based on configuration
func (gsm *GeocodingServiceManager) initializeServices() {
	// Try to get Amap API key from config
	amapKey := config.ServicePublicURL.GetString() // This is a placeholder, replace with actual config key
	if amapKey != "" {
		amapService := NewAmapGeocodingService(amapKey)
		gsm.AddService(amapService)
		log.Info("Initialized Amap geocoding service")
	}

	// Always add mock service as fallback
	mockService := models.NewMockGeocodingService()
	gsm.AddService(mockService)
	log.Info("Initialized mock geocoding service as fallback")
}

// GeocodeWithRetry geocodes with retry logic and accuracy validation
func (gsm *GeocodingServiceManager) GeocodeWithRetry(address string, minAccuracy float64, maxRetries int) (*models.GeocodeResult, error) {
	var lastResult *models.GeocodeResult
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		result, err := gsm.Geocode(address)
		if err != nil {
			lastErr = err
			continue
		}

		lastResult = result

		// Check if accuracy meets minimum requirement
		if result.Accuracy >= minAccuracy {
			return result, nil
		}

		// If this is not the last attempt, wait before retrying
		if attempt < maxRetries-1 {
			time.Sleep(time.Duration(attempt+1) * time.Second)
		}
	}

	// If we have a result but it doesn't meet accuracy requirements
	if lastResult != nil {
		log.Warningf("Geocoding result for '%s' has low accuracy: %.2f (required: %.2f)", 
			address, lastResult.Accuracy, minAccuracy)
		return lastResult, nil
	}

	// If all attempts failed
	if lastErr != nil {
		return nil, fmt.Errorf("geocoding failed after %d attempts: %w", maxRetries, lastErr)
	}

	return nil, fmt.Errorf("geocoding failed after %d attempts with no results", maxRetries)
}

// ValidateAndNormalizeAddress validates and normalizes an address before geocoding
func (gsm *GeocodingServiceManager) ValidateAndNormalizeAddress(address string) (string, error) {
	// Trim whitespace
	address = strings.TrimSpace(address)
	
	if address == "" {
		return "", fmt.Errorf("address cannot be empty")
	}

	// Remove excessive whitespace
	address = strings.Join(strings.Fields(address), " ")

	// Basic validation for Chinese addresses
	if len(address) < 3 {
		return "", fmt.Errorf("address too short, minimum 3 characters required")
	}

	if len(address) > 500 {
		return "", fmt.Errorf("address too long, maximum 500 characters allowed")
	}

	// Add common suffixes if missing (for Chinese addresses)
	commonSuffixes := []string{"市", "区", "县", "镇", "街道", "路", "街", "巷", "号"}
	hasSuffix := false
	for _, suffix := range commonSuffixes {
		if strings.HasSuffix(address, suffix) {
			hasSuffix = true
			break
		}
	}

	// If no common suffix found, it might be an incomplete address
	if !hasSuffix {
		log.Warningf("Address '%s' may be incomplete (no common suffix found)", address)
	}

	return address, nil
}

// GetGeocodingStatistics returns statistics about geocoding operations
func (gsm *GeocodingServiceManager) GetGeocodingStatistics() map[string]interface{} {
	stats := map[string]interface{}{
		"cache_size": gsm.GetCacheSize(),
		"services":   []string{},
	}

	// This would need to be implemented to track service usage
	// For now, return basic info
	return stats
}

// ClearGeocodingCache clears the geocoding cache
func (gsm *GeocodingServiceManager) ClearGeocodingCache() {
	gsm.ClearCache()
	log.Info("Geocoding cache cleared")
}

// TestGeocodingService tests if geocoding services are working
func (gsm *GeocodingServiceManager) TestGeocodingService() error {
	testAddress := "北京市朝阳区"
	
	result, err := gsm.Geocode(testAddress)
	if err != nil {
		return fmt.Errorf("geocoding test failed: %w", err)
	}

	if result.Accuracy == 0 {
		return fmt.Errorf("geocoding test returned zero accuracy")
	}

	log.Infof("Geocoding test successful: %s -> %.6f,%.6f (accuracy: %.2f)", 
		testAddress, result.Longitude, result.Latitude, result.Accuracy)

	return nil
}

// GetDefaultGeocodingService returns a default geocoding service instance
func GetDefaultGeocodingService() models.GeocodingService {
	// Try to create Amap service first
	amapKey := "" // Get from config
	if amapKey != "" {
		return NewAmapGeocodingService(amapKey)
	}

	// Fallback to mock service
	return models.NewMockGeocodingService()
}