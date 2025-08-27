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

package routes

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	auth2 "code.vikunja.io/api/pkg/modules/auth"
	"code.vikunja.io/api/pkg/web/handler"

	"github.com/labstack/echo/v4"
	"github.com/xuri/excelize/v2"
)

func registerMerchantRoutes(a *echo.Group) {
	merchantHandler := &handler.WebHandler{
		EmptyStruct: func() handler.CObject {
			return &models.Merchant{}
		},
	}

	// Basic CRUD operations
	a.GET("/merchants", merchantHandler.ReadAllWeb)
	a.PUT("/merchants", merchantHandler.CreateWeb)
	a.GET("/merchants/:merchant", merchantHandler.ReadOneWeb)
	a.POST("/merchants/:merchant", merchantHandler.UpdateWeb)
	a.DELETE("/merchants/:merchant", merchantHandler.DeleteWeb)
	
	// Import/Export operations
	a.PUT("/merchants/import", merchantImport)
	a.GET("/merchants/export", merchantExport)
	
	// Advanced search and filtering
	a.POST("/merchants/search", merchantSearch)
	
	// Batch operations
	a.POST("/merchants/batch", merchantBatchOperations)
	a.POST("/merchants/batch/geocode", merchantBatchGeocode)
	
	// Location-based operations
	a.GET("/merchants/nearby", merchantNearby)
	a.POST("/merchants/:merchant/geocode", merchantGeocode)
	a.PUT("/merchants/:merchant/location", merchantUpdateLocation)
	
	// Statistics and analytics
	a.GET("/merchants/statistics", merchantStatistics)
	
	// Duplicate operations
	a.POST("/merchants/:merchant/duplicate", merchantDuplicate)
}

// merchantImport handles XLSX import for merchants
// @Summary Import merchants from XLSX
// @Description Import merchants from an uploaded XLSX file
// @tags merchant
// @Accept multipart/form-data
// @Param file formData file true "XLSX file to import"
// @Produce json
// @Security JWTKeyAuth
// @Success 200 {array} models.Merchant "The imported merchants"
// @Failure 400 {object} web.HTTPError "Invalid file format"
// @Failure 500 {object} web.HTTPError "Internal server error"
// @Router /merchants/import [put]
func merchantImport(c echo.Context) error {
	// Get auth
	auth, err := auth2.GetAuthFromClaims(c)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	// Get database session
	s := db.NewSession()
	defer s.Close()

	// Get uploaded file
	file, err := c.FormFile("file")
	if err != nil {
		// Log the actual error for debugging
		fmt.Printf("Error getting form file: %v\n", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "No file uploaded: " + err.Error()})
	}

	// Log file information
	fmt.Printf("Received file: %s, Size: %d, Header: %v\n", file.Filename, file.Size, file.Header)

	// Open the uploaded file
	src, err := file.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Could not open file"})
	}
	defer src.Close()

	// Read Excel file
	xlsx, err := excelize.OpenReader(src)
	if err != nil {
		fmt.Printf("Error opening Excel file: %v\n", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid XLSX file: " + err.Error()})
	}
	defer xlsx.Close()

	// Get the first sheet
	sheets := xlsx.GetSheetList()
	if len(sheets) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "No sheets found in XLSX file"})
	}

	rows, err := xlsx.GetRows(sheets[0])
	if err != nil {
		fmt.Printf("Error reading sheet data: %v\n", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Could not read sheet data: " + err.Error()})
	}

	fmt.Printf("Found %d rows in Excel file\n", len(rows))
	if len(rows) > 0 {
		fmt.Printf("First row (header): %v\n", rows[0])
	}

	if len(rows) < 2 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "XLSX file must have header row and at least one data row. Found " + strconv.Itoa(len(rows)) + " rows."})
	}

	// Parse header row
	header := rows[0]
	expectedHeaders := []string{"法人", "经营地址", "商圈", "有效时间", "交通情况", "固定事件", "终端类型", "特殊时段", "自定义筛选"}
	
	// Create a map of header positions
	headerMap := make(map[string]int)
	for i, h := range header {
		// Normalize header by trimming spaces
		normalizedHeader := strings.TrimSpace(h)
		headerMap[normalizedHeader] = i
	}

	// Check if all required headers are present
	for _, expected := range expectedHeaders {
		if _, exists := headerMap[expected]; !exists {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": fmt.Sprintf("Missing required header column: %s", expected),
			})
		}
	}

	// Import merchants
	var importedMerchants []*models.Merchant

	for i, row := range rows[1:] { // Skip header row
		if len(row) == 0 {
			continue // Skip empty rows
		}

		// Create merchant with required fields
		merchant := &models.Merchant{
			OwnerID: auth.GetID(),
		}

		// Map columns to merchant fields based on header positions
		if pos, exists := headerMap["法人"]; exists && pos < len(row) {
			merchant.LegalRepresentative = row[pos]
		}
		if pos, exists := headerMap["经营地址"]; exists && pos < len(row) {
			merchant.BusinessAddress = row[pos]
		}
		if pos, exists := headerMap["商圈"]; exists && pos < len(row) {
			merchant.BusinessDistrict = row[pos]
		}
		if pos, exists := headerMap["有效时间"]; exists && pos < len(row) {
			merchant.ValidTime = row[pos]
		}
		if pos, exists := headerMap["交通情况"]; exists && pos < len(row) {
			merchant.TrafficConditions = row[pos]
		}
		if pos, exists := headerMap["固定事件"]; exists && pos < len(row) {
			merchant.FixedEvents = row[pos]
		}
		if pos, exists := headerMap["终端类型"]; exists && pos < len(row) {
			merchant.TerminalType = row[pos]
		}
		if pos, exists := headerMap["特殊时段"]; exists && pos < len(row) {
			merchant.SpecialTimePeriods = row[pos]
		}
		if pos, exists := headerMap["自定义筛选"]; exists && pos < len(row) {
			merchant.CustomFilters = row[pos]
		}

		// Set title as a combination of key fields or use a default
		title := merchant.LegalRepresentative
		if title == "" {
			title = fmt.Sprintf("商户 #%d", i+1)
		}
		merchant.Title = title

		// Skip if all fields are empty
		if merchant.LegalRepresentative == "" && 
		   merchant.BusinessAddress == "" && 
		   merchant.BusinessDistrict == "" {
			continue
		}

		// Create merchant
		err = merchant.Create(s, auth)
		if err != nil {
			_ = s.Rollback()
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Error creating merchant at row " + strconv.Itoa(i+2) + ": " + err.Error(),
			})
		}

		importedMerchants = append(importedMerchants, merchant)
	}

	if err = s.Commit(); err != nil {
		return handler.HandleHTTPError(err)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Import completed successfully",
		"count":   len(importedMerchants),
		"merchants": importedMerchants,
	})
}

// merchantExport handles merchant export to various formats
// @Summary Export merchants
// @Description Export merchants to XLSX format with filtering options
// @tags merchant
// @Accept json
// @Produce application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Security JWTKeyAuth
// @Param search query string false "Search term"
// @Param tags query string false "Comma-separated tag IDs"
// @Param districts query string false "Comma-separated districts"
// @Success 200 {file} file "XLSX file"
// @Failure 400 {object} web.HTTPError "Invalid parameters"
// @Failure 500 {object} web.HTTPError "Internal server error"
// @Router /merchants/export [get]
func merchantExport(c echo.Context) error {
	// Implementation would go here
	return c.JSON(http.StatusNotImplemented, map[string]string{
		"message": "Export functionality not yet implemented",
	})
}

// merchantSearch handles advanced merchant search
// @Summary Advanced merchant search
// @Description Search merchants with advanced filtering and sorting options
// @tags merchant
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param search body services.MerchantSearchOptions true "Search options"
// @Success 200 {object} services.MerchantSearchResult "Search results"
// @Failure 400 {object} web.HTTPError "Invalid search parameters"
// @Failure 500 {object} web.HTTPError "Internal server error"
// @Router /merchants/search [post]
func merchantSearch(c echo.Context) error {
	// Implementation would go here
	return c.JSON(http.StatusNotImplemented, map[string]string{
		"message": "Advanced search functionality not yet implemented",
	})
}

// merchantBatchOperations handles batch operations on merchants
// @Summary Batch merchant operations
// @Description Perform batch operations (update, delete) on multiple merchants
// @tags merchant
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param operation body object true "Batch operation details"
// @Success 200 {object} object "Operation results"
// @Failure 400 {object} web.HTTPError "Invalid operation"
// @Failure 500 {object} web.HTTPError "Internal server error"
// @Router /merchants/batch [post]
func merchantBatchOperations(c echo.Context) error {
	// Implementation would go here
	return c.JSON(http.StatusNotImplemented, map[string]string{
		"message": "Batch operations functionality not yet implemented",
	})
}

// merchantBatchGeocode handles batch geocoding of merchants
// @Summary Batch geocode merchants
// @Description Geocode multiple merchants at once
// @tags merchant
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param merchant_ids body []int64 true "Array of merchant IDs to geocode"
// @Success 200 {object} object "Geocoding results"
// @Failure 400 {object} web.HTTPError "Invalid merchant IDs"
// @Failure 500 {object} web.HTTPError "Internal server error"
// @Router /merchants/batch/geocode [post]
func merchantBatchGeocode(c echo.Context) error {
	// Implementation would go here
	return c.JSON(http.StatusNotImplemented, map[string]string{
		"message": "Batch geocoding functionality not yet implemented",
	})
}

// merchantNearby finds merchants near a location
// @Summary Find nearby merchants
// @Description Find merchants within a specified radius of a location
// @tags merchant
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param lng query float64 true "Longitude"
// @Param lat query float64 true "Latitude"
// @Param radius query float64 false "Radius in kilometers (default: 5)"
// @Param limit query int false "Maximum number of results (default: 20)"
// @Success 200 {array} models.MerchantWithDistance "Nearby merchants with distances"
// @Failure 400 {object} web.HTTPError "Invalid coordinates"
// @Failure 500 {object} web.HTTPError "Internal server error"
// @Router /merchants/nearby [get]
func merchantNearby(c echo.Context) error {
	// Implementation would go here
	return c.JSON(http.StatusNotImplemented, map[string]string{
		"message": "Nearby merchants functionality not yet implemented",
	})
}

// merchantGeocode geocodes a specific merchant
// @Summary Geocode merchant address
// @Description Geocode the address of a specific merchant
// @tags merchant
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param merchant path int true "Merchant ID"
// @Success 200 {object} models.GeocodeResult "Geocoding result"
// @Failure 400 {object} web.HTTPError "Invalid merchant ID"
// @Failure 404 {object} web.HTTPError "Merchant not found"
// @Failure 500 {object} web.HTTPError "Internal server error"
// @Router /merchants/{merchant}/geocode [post]
func merchantGeocode(c echo.Context) error {
	// Implementation would go here
	return c.JSON(http.StatusNotImplemented, map[string]string{
		"message": "Merchant geocoding functionality not yet implemented",
	})
}

// merchantUpdateLocation updates merchant location manually
// @Summary Update merchant location
// @Description Manually update the geographic coordinates of a merchant
// @tags merchant
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param merchant path int true "Merchant ID"
// @Param location body object true "Location data (lng, lat)"
// @Success 200 {object} models.Merchant "Updated merchant"
// @Failure 400 {object} web.HTTPError "Invalid location data"
// @Failure 404 {object} web.HTTPError "Merchant not found"
// @Failure 500 {object} web.HTTPError "Internal server error"
// @Router /merchants/{merchant}/location [put]
func merchantUpdateLocation(c echo.Context) error {
	// Implementation would go here
	return c.JSON(http.StatusNotImplemented, map[string]string{
		"message": "Location update functionality not yet implemented",
	})
}

// merchantStatistics returns merchant statistics
// @Summary Get merchant statistics
// @Description Get various statistics about merchants
// @tags merchant
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Success 200 {object} object "Statistics data"
// @Failure 500 {object} web.HTTPError "Internal server error"
// @Router /merchants/statistics [get]
func merchantStatistics(c echo.Context) error {
	// Implementation would go here
	return c.JSON(http.StatusNotImplemented, map[string]string{
		"message": "Statistics functionality not yet implemented",
	})
}

// merchantDuplicate creates a duplicate of an existing merchant
// @Summary Duplicate merchant
// @Description Create a copy of an existing merchant
// @tags merchant
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param merchant path int true "Merchant ID to duplicate"
// @Success 201 {object} models.Merchant "Duplicated merchant"
// @Failure 400 {object} web.HTTPError "Invalid merchant ID"
// @Failure 404 {object} web.HTTPError "Merchant not found"
// @Failure 500 {object} web.HTTPError "Internal server error"
// @Router /merchants/{merchant}/duplicate [post]
func merchantDuplicate(c echo.Context) error {
	// Implementation would go here
	return c.JSON(http.StatusNotImplemented, map[string]string{
		"message": "Duplicate functionality not yet implemented",
	})
}