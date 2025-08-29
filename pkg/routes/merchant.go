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

	a.GET("/merchants", merchantHandler.ReadAllWeb)
	a.PUT("/merchants", merchantHandler.CreateWeb)
	a.GET("/merchants/:merchant", merchantHandler.ReadOneWeb)
	a.POST("/merchants/:merchant", merchantHandler.UpdateWeb)
	a.DELETE("/merchants/:merchant", merchantHandler.DeleteWeb)
	a.PUT("/merchants/import", merchantImport)
	a.POST("/merchants/bulk_delete", merchantBulkDelete)
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
		"message":   "Import completed successfully",
		"count":     len(importedMerchants),
		"merchants": importedMerchants,
	})
}

// merchantBulkDelete handles bulk deletion of merchants
// @Summary Bulk delete merchants
// @Description Deletes multiple merchants by ID
// @tags merchant
// @Accept json
// @Param ids body object true "{\"ids\": [1,2,3]}"
// @Produce json
// @Security JWTKeyAuth
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} web.HTTPError
// @Failure 500 {object} web.HTTPError
// @Router /merchants/bulk_delete [post]
func merchantBulkDelete(c echo.Context) error {
	// Get auth
	auth, err := auth2.GetAuthFromClaims(c)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	// Get database session
	s := db.NewSession()
	defer s.Close()

	// Parse body
	var payload struct {
		IDs []int64 `json:"ids"`
	}
	if err := c.Bind(&payload); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid payload"})
	}

	if len(payload.IDs) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "No ids provided"})
	}

	deleted := 0
	for _, id := range payload.IDs {
		m := &models.Merchant{ID: id}
		if err := m.Delete(s, auth); err != nil {
			// rollback and return error
			_ = s.Rollback()
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Error deleting merchant: " + err.Error()})
		}
		deleted++
	}

	if err = s.Commit(); err != nil {
		return handler.HandleHTTPError(err)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"deleted": deleted,
	})
}

// swagger:route GET /merchants merchant listMerchants
// Lists all merchants for the current user.
// responses:
//   200: merchantList
//   401: HTTPError
//   500: HTTPError

// swagger:route POST /merchants merchant createMerchant
// Creates a new merchant.
// responses:
//   200: merchant
//   400: HTTPError
//   401: HTTPError
//   500: HTTPError

// swagger:route GET /merchants/{merchant} merchant getMerchant
// Gets a merchant by its ID.
// responses:
//   200: merchant
//   401: HTTPError
//   404: HTTPError
//   500: HTTPError

// swagger:route POST /merchants/{merchant} merchant updateMerchant
// Updates a merchant by its ID.
// responses:
//   200: merchant
//   400: HTTPError
//   401: HTTPError
//   404: HTTPError
//   500: HTTPError

// swagger:route DELETE /merchants/{merchant} merchant deleteMerchant
// Deletes a merchant by its ID.
// responses:
//   200: HTTPSuccess
//   401: HTTPError
//   404: HTTPError
//   500: HTTPError

// swagger:route PUT /merchants/import merchant importMerchants
// Imports merchants from an XLSX file.
// responses:
//   200: merchantList
//   400: HTTPError
//   401: HTTPError
//   500: HTTPError
