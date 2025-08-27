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

func registerMerchantTagRoutes(a *echo.Group) {
	merchantTagHandler := &handler.WebHandler{
		EmptyStruct: func() handler.CObject {
			return &models.MerchantTag{}
		},
	}

	// Basic CRUD operations
	a.GET("/merchant-tags", merchantTagHandler.ReadAllWeb)
	a.PUT("/merchant-tags", merchantTagHandler.CreateWeb)
	a.GET("/merchant-tags/:merchanttag", merchantTagHandler.ReadOneWeb)
	a.POST("/merchant-tags/:merchanttag", merchantTagHandler.UpdateWeb)
	a.DELETE("/merchant-tags/:merchanttag", merchantTagHandler.DeleteWeb)

	// Category and classification operations
	a.GET("/merchant-tags/categories", merchantTagCategories)
	a.GET("/merchant-tags/category/:category", merchantTagsByCategory)

	// Batch operations
	a.POST("/merchant-tags/batch", merchantTagBatchCreate)

	// Usage and statistics
	a.GET("/merchant-tags/statistics", merchantTagStatistics)
	a.GET("/merchant-tags/:merchanttag/merchants", merchantTagUsage)

	// Suggestions
	a.POST("/merchant-tags/suggest", merchantTagSuggestions)

	// System tags management
	a.POST("/merchant-tags/system/initialize", merchantTagInitializeSystem)
}

// merchantTagCategories returns all tag categories
// @Summary Get tag categories
// @Description Get all merchant tag categories with their tags
// @tags merchant-tags
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param include_system query bool false "Include system tags (default: true)"
// @Success 200 {object} map[string]services.TagCategoryInfo "Categories with tags"
// @Failure 500 {object} web.HTTPError "Internal server error"
// @Router /merchant-tags/categories [get]
func merchantTagCategories(c echo.Context) error {
	return c.JSON(http.StatusNotImplemented, map[string]string{
		"message": "Tag categories functionality not yet implemented",
	})
}

// merchantTagsByCategory returns tags in a specific category
// @Summary Get tags by category
// @Description Get all merchant tags in a specific category
// @tags merchant-tags
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param category path string true "Category name"
// @Param page query int false "Page number (default: 1)"
// @Param per_page query int false "Items per page (default: 50)"
// @Success 200 {array} models.MerchantTag "Tags in category"
// @Failure 400 {object} web.HTTPError "Invalid category"
// @Failure 500 {object} web.HTTPError "Internal server error"
// @Router /merchant-tags/category/{category} [get]
func merchantTagsByCategory(c echo.Context) error {
	return c.JSON(http.StatusNotImplemented, map[string]string{
		"message": "Tags by category functionality not yet implemented",
	})
}

// merchantTagBatchCreate creates multiple tags at once
// @Summary Batch create tags
// @Description Create multiple merchant tags in a single request
// @tags merchant-tags
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param tags body services.BatchCreateTagRequest true "Tags to create"
// @Success 201 {array} models.MerchantTag "Created tags"
// @Failure 400 {object} web.HTTPError "Invalid tag data"
// @Failure 500 {object} web.HTTPError "Internal server error"
// @Router /merchant-tags/batch [post]
func merchantTagBatchCreate(c echo.Context) error {
	return c.JSON(http.StatusNotImplemented, map[string]string{
		"message": "Batch tag creation functionality not yet implemented",
	})
}

// merchantTagStatistics returns tag usage statistics
// @Summary Get tag statistics
// @Description Get statistics about merchant tag usage
// @tags merchant-tags
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Success 200 {object} object "Tag statistics"
// @Failure 500 {object} web.HTTPError "Internal server error"
// @Router /merchant-tags/statistics [get]
func merchantTagStatistics(c echo.Context) error {
	return c.JSON(http.StatusNotImplemented, map[string]string{
		"message": "Tag statistics functionality not yet implemented",
	})
}

// merchantTagUsage returns merchants using a specific tag
// @Summary Get merchants using tag
// @Description Get all merchants that use a specific tag
// @tags merchant-tags
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param merchanttag path int true "Tag ID"
// @Param page query int false "Page number (default: 1)"
// @Param per_page query int false "Items per page (default: 20)"
// @Success 200 {object} object "Merchants using the tag"
// @Failure 400 {object} web.HTTPError "Invalid tag ID"
// @Failure 404 {object} web.HTTPError "Tag not found"
// @Failure 500 {object} web.HTTPError "Internal server error"
// @Router /merchant-tags/{merchanttag}/merchants [get]
func merchantTagUsage(c echo.Context) error {
	return c.JSON(http.StatusNotImplemented, map[string]string{
		"message": "Tag usage functionality not yet implemented",
	})
}

// merchantTagSuggestions suggests tags for a merchant
// @Summary Suggest tags for merchant
// @Description Get tag suggestions based on merchant data
// @tags merchant-tags
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param merchant body models.Merchant true "Merchant data for suggestions"
// @Success 200 {array} models.MerchantTag "Suggested tags"
// @Failure 400 {object} web.HTTPError "Invalid merchant data"
// @Failure 500 {object} web.HTTPError "Internal server error"
// @Router /merchant-tags/suggest [post]
func merchantTagSuggestions(c echo.Context) error {
	return c.JSON(http.StatusNotImplemented, map[string]string{
		"message": "Tag suggestions functionality not yet implemented",
	})
}

// merchantTagInitializeSystem initializes system tags
// @Summary Initialize system tags
// @Description Create predefined system tags if they don't exist
// @tags merchant-tags
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Success 200 {object} object "Initialization result"
// @Failure 500 {object} web.HTTPError "Internal server error"
// @Router /merchant-tags/system/initialize [post]
func merchantTagInitializeSystem(c echo.Context) error {
	return c.JSON(http.StatusNotImplemented, map[string]string{
		"message": "System tag initialization functionality not yet implemented",
	})
}