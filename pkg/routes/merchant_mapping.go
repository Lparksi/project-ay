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

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	auth2 "code.vikunja.io/api/pkg/modules/auth"
	"code.vikunja.io/api/pkg/web/handler"

	"github.com/labstack/echo/v4"
)

func registerMerchantMappingRoutes(a *echo.Group) {
	mappingHandler := &handler.WebHandler{
		EmptyStruct: func() handler.CObject {
			return &models.MerchantMapping{}
		},
	}

	a.GET("/merchant-mappings", mappingHandler.ReadAllWeb)
	a.PUT("/merchant-mappings", mappingHandler.CreateWeb)
	a.GET("/merchant-mappings/:merchantMapping", mappingHandler.ReadOneWeb)
	a.POST("/merchant-mappings/:merchantMapping", mappingHandler.UpdateWeb)
	a.DELETE("/merchant-mappings/:merchantMapping", mappingHandler.DeleteWeb)
	a.POST("/merchant-mappings/bulk_save", merchantMappingBulkSave)
	a.DELETE("/merchant-mappings/bulk_delete", merchantMappingBulkDelete)
}

// merchantMappingBulkSave handles bulk save/update of merchant mappings
// @Summary Bulk save merchant mappings
// @Description Saves multiple merchant mappings at once, replacing all existing mappings for the specified fields
// @tags merchantMapping
// @Accept json
// @Param mappings body object true "Field mappings object"
// @Produce json
// @Security JWTKeyAuth
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} web.HTTPError
// @Failure 500 {object} web.HTTPError
// @Router /merchant-mappings/bulk_save [post]
func merchantMappingBulkSave(c echo.Context) error {
	// Get auth
	auth, err := auth2.GetAuthFromClaims(c)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	// Get database session
	s := db.NewSession()
	defer s.Close()

	// Begin transaction
	if err := s.Begin(); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to begin transaction: " + err.Error(),
		})
	}

	// Parse body - expecting field mappings format from frontend
	var payload struct {
		FieldMappings []struct {
			Field    string `json:"field"`
			Mappings []struct {
				Placeholder string `json:"placeholder"`
				DisplayText string `json:"displayText"`
				LabelID     int64  `json:"labelId"`
			} `json:"mappings"`
		} `json:"fieldMappings"`
	}

	if err := c.Bind(&payload); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid payload"})
	}

	// Process each field's mappings
	totalSaved := 0
	for _, fieldMapping := range payload.FieldMappings {
		// First, delete existing mappings for this field
		err := models.BulkDeleteMerchantMappingsByField(s, fieldMapping.Field, auth.GetID())
		if err != nil {
			_ = s.Rollback()
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Error deleting existing mappings: " + err.Error(),
			})
		}

		// Create new mappings
		var newMappings []*models.MerchantMapping
		for _, mapping := range fieldMapping.Mappings {
			// Skip empty mappings
			if mapping.Placeholder == "" || mapping.DisplayText == "" {
				continue
			}

			newMapping := &models.MerchantMapping{
				FieldName:   fieldMapping.Field,
				Placeholder: mapping.Placeholder,
				DisplayText: mapping.DisplayText,
				LabelID:     mapping.LabelID,
				IsActive:    true,
				OwnerID:     auth.GetID(),
			}
			newMappings = append(newMappings, newMapping)
		}

		// Bulk insert new mappings
		if len(newMappings) > 0 {
			err = models.BulkCreateMerchantMappings(s, newMappings, auth)
			if err != nil {
				_ = s.Rollback()
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"error": "Error creating new mappings: " + err.Error(),
				})
			}
			totalSaved += len(newMappings)
		}
	}

	if err = s.Commit(); err != nil {
		return handler.HandleHTTPError(err)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Mappings saved successfully",
		"count":   totalSaved,
	})
}

// merchantMappingBulkDelete handles bulk deletion of merchant mappings by field
// @Summary Bulk delete merchant mappings
// @Description Deletes all mappings for specified fields
// @tags merchantMapping
// @Accept json
// @Param payload body object true "{\"fields\": [\"validTime\", \"trafficConditions\"]}"
// @Produce json
// @Security JWTKeyAuth
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} web.HTTPError
// @Failure 500 {object} web.HTTPError
// @Router /merchant-mappings/bulk_delete [delete]
func merchantMappingBulkDelete(c echo.Context) error {
	// Get auth
	auth, err := auth2.GetAuthFromClaims(c)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	// Get database session
	s := db.NewSession()
	defer s.Close()

	// Begin transaction
	if err := s.Begin(); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to begin transaction: " + err.Error(),
		})
	}

	// Parse body
	var payload struct {
		Fields []string `json:"fields"`
	}
	if err := c.Bind(&payload); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid payload"})
	}

	if len(payload.Fields) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "No fields provided"})
	}

	totalDeleted := 0
	for _, field := range payload.Fields {
		err := models.BulkDeleteMerchantMappingsByField(s, field, auth.GetID())
		if err != nil {
			_ = s.Rollback()
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Error deleting mappings: " + err.Error(),
			})
		}
		totalDeleted++
	}

	if err = s.Commit(); err != nil {
		return handler.HandleHTTPError(err)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"deleted": totalDeleted,
	})
}

// swagger:route GET /merchant-mappings merchantMapping listMerchantMappings
// Lists all merchant mappings for the current user.
// responses:
//   200: merchantMappingList
//   401: HTTPError
//   500: HTTPError

// swagger:route POST /merchant-mappings merchantMapping createMerchantMapping
// Creates a new merchant mapping.
// responses:
//   200: merchantMapping
//   400: HTTPError
//   401: HTTPError
//   500: HTTPError

// swagger:route GET /merchant-mappings/{merchantMapping} merchantMapping getMerchantMapping
// Gets a merchant mapping by its ID.
// responses:
//   200: merchantMapping
//   401: HTTPError
//   404: HTTPError
//   500: HTTPError

// swagger:route POST /merchant-mappings/{merchantMapping} merchantMapping updateMerchantMapping
// Updates a merchant mapping by its ID.
// responses:
//   200: merchantMapping
//   401: HTTPError
//   404: HTTPError
//   500: HTTPError

// swagger:route DELETE /merchant-mappings/{merchantMapping} merchantMapping deleteMerchantMapping
// Deletes a merchant mapping by its ID.
// responses:
//   200: HTTPSuccess
//   401: HTTPError
//   404: HTTPError
//   500: HTTPError
