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
	"net/http"
	"strconv"

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
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "No file uploaded"})
	}

	// Open the uploaded file
	src, err := file.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Could not open file"})
	}
	defer src.Close()

	// Read Excel file
	xlsx, err := excelize.OpenReader(src)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid XLSX file"})
	}
	defer xlsx.Close()

	// Get the first sheet
	sheets := xlsx.GetSheetList()
	if len(sheets) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "No sheets found in XLSX file"})
	}

	rows, err := xlsx.GetRows(sheets[0])
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Could not read sheet data"})
	}

	if len(rows) < 2 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "XLSX file must have header row and at least one data row"})
	}

	// Assume first row is header
	// Expected columns: Title, LegalRepresentative, BusinessAddress, BusinessDistrict, ValidTime, 
	//                  TrafficConditions, FixedEvents, TerminalType, SpecialTimePeriods, CustomFilters
	var importedMerchants []*models.Merchant

	for i, row := range rows[1:] { // Skip header row
		if len(row) < 1 {
			continue // Skip empty rows
		}

		merchant := &models.Merchant{
			OwnerID: auth.GetID(),
		}

		// Map columns to merchant fields
		if len(row) > 0 {
			merchant.Title = row[0]
		}
		if len(row) > 1 {
			merchant.LegalRepresentative = row[1]
		}
		if len(row) > 2 {
			merchant.BusinessAddress = row[2]
		}
		if len(row) > 3 {
			merchant.BusinessDistrict = row[3]
		}
		if len(row) > 4 {
			merchant.ValidTime = row[4]
		}
		if len(row) > 5 {
			merchant.TrafficConditions = row[5]
		}
		if len(row) > 6 {
			merchant.FixedEvents = row[6]
		}
		if len(row) > 7 {
			merchant.TerminalType = row[7]
		}
		if len(row) > 8 {
			merchant.SpecialTimePeriods = row[8]
		}
		if len(row) > 9 {
			merchant.CustomFilters = row[9]
		}

		// Skip if no title (required field)
		if merchant.Title == "" {
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