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

	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/web/handler"

	"github.com/labstack/echo/v4"
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
}

// MerchantImport handles XLSX import for merchants
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
func MerchantImport(c echo.Context) error {
	// This will be implemented with the XLSX import functionality
	return c.JSON(http.StatusNotImplemented, map[string]string{
		"message": "XLSX import will be implemented",
	})
}