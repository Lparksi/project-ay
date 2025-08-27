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

// MerchantTagService provides business logic for merchant tag operations
type MerchantTagService struct{}

// NewMerchantTagService creates a new merchant tag service
func NewMerchantTagService() *MerchantTagService {
	return &MerchantTagService{}
}

// TagSearchOptions represents search and filter options for tags
type TagSearchOptions struct {
	Search      string   `json:"search"`
	Categories  []string `json:"categories"`
	IncludeSystem bool   `json:"include_system"`
	SystemOnly  bool     `json:"system_only"`
	Page        int      `json:"page"`
	PerPage     int      `json:"per_page"`
	SortBy      string   `json:"sort_by"`
	SortOrder   string   `json:"sort_order"`
}

// TagSearchResult represents the result of a tag search
type TagSearchResult struct {
	Tags       []*models.MerchantTag `json:"tags"`
	TotalCount int64                 `json:"total_count"`
	Page       int                   `json:"page"`
	PerPage    int                   `json:"per_page"`
	HasMore    bool                  `json:"has_more"`
}

// TagCategoryInfo represents information about a tag category
type TagCategoryInfo struct {
	Category    string                `json:"category"`
	DisplayName string                `json:"display_name"`
	Tags        []*models.MerchantTag `json:"tags"`
	Count       int                   `json:"count"`
}

// BatchCreateTagRequest represents a request to create multiple tags
type BatchCreateTagRequest struct {
	Tags []struct {
		Title       string `json:"title" valid:"required"`
		Description string `json:"description"`
		HexColor    string `json:"hex_color"`
		Category    string `json:"category"`
		SortOrder   int    `json:"sort_order"`
	} `json:"tags"`
}

// SearchTags performs advanced search with filtering and sorting
func (mts *MerchantTagService) SearchTags(s *xorm.Session, auth web.Auth, options TagSearchOptions) (*TagSearchResult, error) {
	query := s.NewSession()

	// Apply system tag filters
	if options.SystemOnly {
		query = query.Where("is_system = ?", true)
	} else if !options.IncludeSystem {
		query = query.Where("is_system = ?", false)
	}

	// Apply text search
	if options.Search != "" {
		searchTerm := "%" + strings.ToLower(options.Search) + "%"
		query = query.And("(LOWER(title) LIKE ? OR LOWER(description) LIKE ? OR LOWER(category) LIKE ?)",
			searchTerm, searchTerm, searchTerm)
	}

	// Apply category filters
	if len(options.Categories) > 0 {
		query = query.And("category IN (?)", options.Categories)
	}

	// Get total count before pagination
	totalCount, err := query.Count(&models.MerchantTag{})
	if err != nil {
		return nil, err
	}

	// Apply sorting
	orderBy := "category ASC, sort_order ASC, title ASC" // Default sort
	if options.SortBy != "" {
		validSortFields := map[string]bool{
			"title":       true,
			"category":    true,
			"sort_order":  true,
			"created":     true,
			"updated":     true,
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
		options.PerPage = 50
	}
	
	limit := options.PerPage
	offset := (options.Page - 1) * options.PerPage

	// Execute query
	tags := []*models.MerchantTag{}
	err = query.
		OrderBy(orderBy).
		Limit(limit, offset).
		Find(&tags)
	if err != nil {
		return nil, err
	}

	// Load creators for each tag
	for _, tag := range tags {
		tag.ReadOne(s, auth) // This loads the creator
	}

	hasMore := int64(offset+len(tags)) < totalCount

	return &TagSearchResult{
		Tags:       tags,
		TotalCount: totalCount,
		Page:       options.Page,
		PerPage:    options.PerPage,
		HasMore:    hasMore,
	}, nil
}

// GetTagsByCategory returns all tags grouped by category
func (mts *MerchantTagService) GetTagsByCategory(s *xorm.Session, auth web.Auth, includeSystem bool) (map[string]*TagCategoryInfo, error) {
	query := s.NewSession()
	
	if !includeSystem {
		query = query.Where("is_system = ?", false)
	}

	tags := []*models.MerchantTag{}
	err := query.OrderBy("category ASC, sort_order ASC, title ASC").Find(&tags)
	if err != nil {
		return nil, err
	}

	// Group by category
	categories := make(map[string]*TagCategoryInfo)
	categoryDisplayNames := map[string]string{
		"location": "位置类型",
		"type":     "商户类型",
		"scale":    "规模大小",
		"status":   "营业状态",
		"service":  "服务类型",
		"":         "未分类",
	}

	for _, tag := range tags {
		category := tag.Category
		if category == "" {
			category = "uncategorized"
		}

		if _, exists := categories[category]; !exists {
			displayName := categoryDisplayNames[category]
			if displayName == "" {
				displayName = category
			}
			categories[category] = &TagCategoryInfo{
				Category:    category,
				DisplayName: displayName,
				Tags:        []*models.MerchantTag{},
				Count:       0,
			}
		}

		categories[category].Tags = append(categories[category].Tags, tag)
		categories[category].Count++
	}

	return categories, nil
}

// CreateTag creates a new merchant tag with validation
func (mts *MerchantTagService) CreateTag(s *xorm.Session, auth web.Auth, tag *models.MerchantTag) error {
	// Validate tag data
	err := mts.ValidateTagData(tag)
	if err != nil {
		return err
	}

	// Check for duplicate titles within the same category
	exists, err := s.Where("title = ? AND category = ?", tag.Title, tag.Category).Exist(&models.MerchantTag{})
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("tag with title '%s' already exists in category '%s'", tag.Title, tag.Category)
	}

	// Create the tag
	return tag.Create(s, auth)
}

// BatchCreateTags creates multiple tags at once
func (mts *MerchantTagService) BatchCreateTags(s *xorm.Session, auth web.Auth, request BatchCreateTagRequest) ([]*models.MerchantTag, error) {
	if len(request.Tags) == 0 {
		return nil, fmt.Errorf("no tags provided")
	}

	if len(request.Tags) > 100 {
		return nil, fmt.Errorf("too many tags, maximum 100 allowed")
	}

	createdTags := make([]*models.MerchantTag, 0, len(request.Tags))

	for i, tagData := range request.Tags {
		tag := &models.MerchantTag{
			Title:       tagData.Title,
			Description: tagData.Description,
			HexColor:    tagData.HexColor,
			Category:    tagData.Category,
			SortOrder:   tagData.SortOrder,
			IsSystem:    false, // Batch created tags are never system tags
		}

		err := mts.CreateTag(s, auth, tag)
		if err != nil {
			return nil, fmt.Errorf("error creating tag %d (%s): %w", i+1, tagData.Title, err)
		}

		createdTags = append(createdTags, tag)
	}

	return createdTags, nil
}

// UpdateTag updates an existing tag with validation
func (mts *MerchantTagService) UpdateTag(s *xorm.Session, auth web.Auth, tag *models.MerchantTag) error {
	// Validate tag data
	err := mts.ValidateTagData(tag)
	if err != nil {
		return err
	}

	// Check if it's a system tag
	existing := &models.MerchantTag{ID: tag.ID}
	err = existing.ReadOne(s, auth)
	if err != nil {
		return err
	}

	if existing.IsSystem {
		// Only allow updating description and sort order for system tags
		existing.Description = tag.Description
		existing.SortOrder = tag.SortOrder
		return existing.Update(s, auth)
	}

	// Check for duplicate titles within the same category (excluding current tag)
	exists, err := s.Where("title = ? AND category = ? AND id != ?", tag.Title, tag.Category, tag.ID).Exist(&models.MerchantTag{})
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("tag with title '%s' already exists in category '%s'", tag.Title, tag.Category)
	}

	// Update the tag
	return tag.Update(s, auth)
}

// DeleteTag deletes a tag and removes all associations
func (mts *MerchantTagService) DeleteTag(s *xorm.Session, auth web.Auth, tagID int64) error {
	tag := &models.MerchantTag{ID: tagID}
	err := tag.ReadOne(s, auth)
	if err != nil {
		return err
	}

	if tag.IsSystem {
		return fmt.Errorf("cannot delete system tag")
	}

	return tag.Delete(s, auth)
}

// GetTagUsageStatistics returns statistics about tag usage
func (mts *MerchantTagService) GetTagUsageStatistics(s *xorm.Session, auth web.Auth) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total tags
	totalCount, err := s.Count(&models.MerchantTag{})
	if err != nil {
		return nil, err
	}
	stats["total_tags"] = totalCount

	// System vs user tags
	systemCount, err := s.Where("is_system = ?", true).Count(&models.MerchantTag{})
	if err != nil {
		return nil, err
	}
	stats["system_tags"] = systemCount
	stats["user_tags"] = totalCount - systemCount

	// Tags by category
	type CategoryCount struct {
		Category string `json:"category"`
		Count    int    `json:"count"`
	}
	
	categoryCounts := []CategoryCount{}
	err = s.
		Select("COALESCE(category, 'uncategorized') as category, COUNT(*) as count").
		GroupBy("category").
		OrderBy("count DESC").
		Find(&categoryCounts)
	if err != nil {
		return nil, err
	}
	stats["tags_by_category"] = categoryCounts

	// Most used tags
	type TagUsage struct {
		TagID    int64  `json:"tag_id"`
		Title    string `json:"title"`
		Category string `json:"category"`
		Count    int    `json:"count"`
	}
	
	tagUsage := []TagUsage{}
	err = s.SQL(`
		SELECT 
			mt.id as tag_id,
			mt.title,
			mt.category,
			COUNT(mmt.merchant_id) as count
		FROM merchant_tags mt
		LEFT JOIN merchant_merchant_tags mmt ON mt.id = mmt.tag_id
		GROUP BY mt.id, mt.title, mt.category
		ORDER BY count DESC
		LIMIT 10
	`).Find(&tagUsage)
	if err != nil {
		return nil, err
	}
	stats["most_used_tags"] = tagUsage

	// Unused tags
	unusedCount, err := s.SQL(`
		SELECT COUNT(*)
		FROM merchant_tags mt
		LEFT JOIN merchant_merchant_tags mmt ON mt.id = mmt.tag_id
		WHERE mmt.tag_id IS NULL
	`).Count()
	if err != nil {
		return nil, err
	}
	stats["unused_tags"] = unusedCount

	return stats, nil
}

// GetMerchantsUsingTag returns merchants that use a specific tag
func (mts *MerchantTagService) GetMerchantsUsingTag(s *xorm.Session, auth web.Auth, tagID int64, page, perPage int) ([]*models.Merchant, int64, error) {
	if page <= 0 {
		page = 1
	}
	if perPage <= 0 {
		perPage = 20
	}

	// Get total count
	totalCount, err := s.
		Join("INNER", "merchant_merchant_tags", "merchants.id = merchant_merchant_tags.merchant_id").
		Where("merchant_merchant_tags.tag_id = ? AND merchants.owner_id = ?", tagID, auth.GetID()).
		Count(&models.Merchant{})
	if err != nil {
		return nil, 0, err
	}

	// Get merchants
	merchants := []*models.Merchant{}
	limit := perPage
	offset := (page - 1) * perPage

	err = s.
		Join("INNER", "merchant_merchant_tags", "merchants.id = merchant_merchant_tags.merchant_id").
		Where("merchant_merchant_tags.tag_id = ? AND merchants.owner_id = ?", tagID, auth.GetID()).
		Limit(limit, offset).
		OrderBy("merchants.created DESC").
		Find(&merchants)
	if err != nil {
		return nil, 0, err
	}

	// Load associated data
	for _, merchant := range merchants {
		merchant.LoadTags(s)
		merchant.LoadPrimaryGeoPoint(s)
	}

	return merchants, totalCount, nil
}

// SuggestTags suggests tags based on merchant data
func (mts *MerchantTagService) SuggestTags(s *xorm.Session, merchant *models.Merchant) ([]*models.MerchantTag, error) {
	suggestions := []*models.MerchantTag{}

	// Suggest based on business district
	if merchant.BusinessDistrict != "" {
		districtTags := []*models.MerchantTag{}
		err := s.Where("category = 'location' AND (title LIKE ? OR description LIKE ?)", 
			"%"+merchant.BusinessDistrict+"%", "%"+merchant.BusinessDistrict+"%").
			Limit(3).Find(&districtTags)
		if err == nil {
			suggestions = append(suggestions, districtTags...)
		}
	}

	// Suggest based on terminal type
	if merchant.TerminalType != "" {
		typeTags := []*models.MerchantTag{}
		err := s.Where("category = 'type' AND (title LIKE ? OR description LIKE ?)", 
			"%"+merchant.TerminalType+"%", "%"+merchant.TerminalType+"%").
			Limit(3).Find(&typeTags)
		if err == nil {
			suggestions = append(suggestions, typeTags...)
		}
	}

	// Suggest popular tags
	popularTags := []*models.MerchantTag{}
	err := s.SQL(`
		SELECT mt.* 
		FROM merchant_tags mt
		JOIN merchant_merchant_tags mmt ON mt.id = mmt.tag_id
		GROUP BY mt.id
		ORDER BY COUNT(mmt.merchant_id) DESC
		LIMIT 5
	`).Find(&popularTags)
	if err == nil {
		suggestions = append(suggestions, popularTags...)
	}

	// Remove duplicates
	seen := make(map[int64]bool)
	uniqueSuggestions := []*models.MerchantTag{}
	for _, tag := range suggestions {
		if !seen[tag.ID] {
			seen[tag.ID] = true
			uniqueSuggestions = append(uniqueSuggestions, tag)
		}
	}

	return uniqueSuggestions, nil
}

// ValidateTagData validates tag data before creation/update
func (mts *MerchantTagService) ValidateTagData(tag *models.MerchantTag) error {
	if strings.TrimSpace(tag.Title) == "" {
		return fmt.Errorf("tag title is required")
	}

	if len(tag.Title) > 250 {
		return fmt.Errorf("tag title too long, maximum 250 characters")
	}

	if len(tag.Description) > 1000 {
		return fmt.Errorf("tag description too long, maximum 1000 characters")
	}

	if len(tag.Category) > 100 {
		return fmt.Errorf("tag category too long, maximum 100 characters")
	}

	if tag.HexColor != "" && !isValidHexColor(tag.HexColor) {
		return fmt.Errorf("invalid hex color format")
	}

	return nil
}

// isValidHexColor checks if a string is a valid hex color
func isValidHexColor(color string) bool {
	if len(color) != 6 {
		return false
	}
	
	for _, char := range color {
		if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f') || (char >= 'A' && char <= 'F')) {
			return false
		}
	}
	
	return true
}

// InitializeSystemTags creates system tags if they don't exist
func (mts *MerchantTagService) InitializeSystemTags(s *xorm.Session, userID int64) error {
	return models.CreateSystemMerchantTags(s, userID)
}

// ExportTags exports tags to a structured format
func (mts *MerchantTagService) ExportTags(s *xorm.Session, auth web.Auth, options TagSearchOptions) ([]*models.MerchantTag, error) {
	// Remove pagination for export
	options.Page = 0
	options.PerPage = 0

	result, err := mts.SearchTags(s, auth, options)
	if err != nil {
		return nil, err
	}

	return result.Tags, nil
}