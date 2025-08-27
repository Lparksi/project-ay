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
	"time"

	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/utils"
	"code.vikunja.io/api/pkg/web"
	"xorm.io/xorm"
)

// MerchantTag represents a tag that can be associated with merchants
type MerchantTag struct {
	// The unique, numeric id of this merchant tag.
	ID int64 `xorm:"bigint autoincr not null unique pk" json:"id" param:"merchanttag"`
	// The title of the merchant tag.
	Title string `xorm:"varchar(250) not null" json:"title" valid:"required,runelength(1|250)" minLength:"1" maxLength:"250"`
	// The description of the merchant tag.
	Description string `xorm:"longtext null" json:"description"`
	// The color this tag has in hex format.
	HexColor string `xorm:"varchar(6) null" json:"hex_color" valid:"runelength(0|7)" maxLength:"7"`
	// The category/class this tag belongs to (e.g., "location", "type", "service").
	Category string `xorm:"varchar(100) null" json:"category" valid:"runelength(0|100)" maxLength:"100"`
	// Whether this is a system-defined tag (cannot be deleted by users).
	IsSystem bool `xorm:"bool default false" json:"is_system"`
	// Sort order for display purposes.
	SortOrder int `xorm:"int default 0" json:"sort_order"`

	CreatedByID int64 `xorm:"bigint not null" json:"-"`
	// The user who created this merchant tag
	CreatedBy *user.User `xorm:"-" json:"created_by"`

	// A timestamp when this merchant tag was created. You cannot change this value.
	Created time.Time `xorm:"created not null" json:"created"`
	// A timestamp when this merchant tag was last updated. You cannot change this value.
	Updated time.Time `xorm:"updated not null" json:"updated"`

	web.CRUDable    `xorm:"-" json:"-"`
	web.Permissions `xorm:"-" json:"-"`
}

// MerchantMerchantTag represents the many-to-many relationship between merchants and tags
type MerchantMerchantTag struct {
	ID         int64 `xorm:"bigint autoincr not null unique pk" json:"id"`
	MerchantID int64 `xorm:"bigint not null INDEX" json:"merchant_id"`
	TagID      int64 `xorm:"bigint not null INDEX" json:"tag_id"`
	
	// A timestamp when this association was created.
	Created time.Time `xorm:"created not null" json:"created"`
}

// TableName makes a pretty table name for MerchantTag
func (*MerchantTag) TableName() string {
	return "merchant_tags"
}

// TableName makes a pretty table name for MerchantMerchantTag
func (*MerchantMerchantTag) TableName() string {
	return "merchant_merchant_tags"
}

// GetID returns the ID of the merchant tag
func (mt *MerchantTag) GetID() int64 {
	return mt.ID
}

// Create creates a new merchant tag
func (mt *MerchantTag) Create(s *xorm.Session, a web.Auth) (err error) {
	u, err := user.GetFromAuth(a)
	if err != nil {
		return
	}

	mt.ID = 0
	mt.HexColor = utils.NormalizeHex(mt.HexColor)
	mt.CreatedBy = u
	mt.CreatedByID = u.ID

	_, err = s.Insert(mt)
	return
}

// Update updates a merchant tag
func (mt *MerchantTag) Update(s *xorm.Session, a web.Auth) (err error) {
	mt.HexColor = utils.NormalizeHex(mt.HexColor)

	_, err = s.
		ID(mt.ID).
		Cols(
			"title",
			"description",
			"hex_color",
			"category",
			"sort_order",
		).
		Update(mt)
	if err != nil {
		return
	}

	err = mt.ReadOne(s, a)
	return
}

// Delete deletes a merchant tag
func (mt *MerchantTag) Delete(s *xorm.Session, _ web.Auth) (err error) {
	// Don't allow deletion of system tags
	if mt.IsSystem {
		return ErrCannotDeleteSystemMerchantTag{TagID: mt.ID}
	}

	// Delete all associations first
	_, err = s.Where("tag_id = ?", mt.ID).Delete(&MerchantMerchantTag{})
	if err != nil {
		return err
	}

	// Delete the tag
	_, err = s.ID(mt.ID).Delete(&MerchantTag{})
	return err
}

// ReadAll gets all merchant tags a user can use
func (mt *MerchantTag) ReadAll(s *xorm.Session, a web.Auth, search string, page int, perPage int) (ls interface{}, resultCount int, numberOfEntries int64, err error) {
	query := s.NewSession()

	if search != "" {
		query = query.Where("title LIKE ? OR description LIKE ? OR category LIKE ?", 
			"%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	limit, start := getLimitFromPageIndex(page, perPage)

	// Get total count
	totalItems, err := query.Count(&MerchantTag{})
	if err != nil {
		return nil, 0, 0, err
	}

	// Get merchant tags
	merchantTags := []*MerchantTag{}
	err = query.
		Limit(limit, start).
		OrderBy("category ASC, sort_order ASC, title ASC").
		Find(&merchantTags)
	if err != nil {
		return nil, 0, 0, err
	}

	// Load creators
	for _, tag := range merchantTags {
		tag.CreatedBy, _ = user.GetUserByID(s, tag.CreatedByID)
	}

	numberOfTotalItems := len(merchantTags)

	return merchantTags, numberOfTotalItems, totalItems, nil
}

// ReadOne gets one merchant tag
func (mt *MerchantTag) ReadOne(s *xorm.Session, _ web.Auth) (err error) {
	exists, err := s.Where("id = ?", mt.ID).Get(mt)
	if err != nil {
		return err
	}
	if !exists {
		return ErrMerchantTagDoesNotExist{TagID: mt.ID}
	}

	u, err := user.GetUserByID(s, mt.CreatedByID)
	if err != nil {
		return
	}

	mt.CreatedBy = u
	return
}

// CanWrite checks if the user can write to this merchant tag
func (mt *MerchantTag) CanWrite(s *xorm.Session, auth web.Auth) (bool, error) {
	return mt.checkPermission(s, auth.GetID(), PermissionWrite)
}

// CanRead checks if the user can read this merchant tag
func (mt *MerchantTag) CanRead(s *xorm.Session, auth web.Auth) (bool, int, error) {
	canRead, err := mt.checkPermission(s, auth.GetID(), PermissionRead)
	return canRead, int(PermissionRead), err
}

// CanDelete checks if the user can delete this merchant tag
func (mt *MerchantTag) CanDelete(s *xorm.Session, auth web.Auth) (bool, error) {
	// System tags cannot be deleted
	if mt.IsSystem {
		return false, nil
	}
	return mt.checkPermission(s, auth.GetID(), PermissionAdmin)
}

// CanCreate checks if the user can create a merchant tag
func (mt *MerchantTag) CanCreate(s *xorm.Session, auth web.Auth) (bool, error) {
	return true, nil
}

// CanUpdate checks if the user can update this merchant tag
func (mt *MerchantTag) CanUpdate(s *xorm.Session, auth web.Auth) (bool, error) {
	return mt.checkPermission(s, auth.GetID(), PermissionWrite)
}

func (mt *MerchantTag) checkPermission(s *xorm.Session, userID int64, permission Permission) (bool, error) {
	// Load the merchant tag if we don't have one
	if mt.ID != 0 {
		_, err := s.Where("id = ?", mt.ID).Get(mt)
		if err != nil {
			return false, err
		}
	}

	// Creator can do everything (except delete system tags, handled in CanDelete)
	if mt.CreatedByID == userID {
		return true, nil
	}

	// For now, all users can read all tags, but only creators can modify
	if permission == PermissionRead {
		return true, nil
	}

	return false, nil
}

// GetMerchantTagsByCategory returns all merchant tags grouped by category
func GetMerchantTagsByCategory(s *xorm.Session) (map[string][]*MerchantTag, error) {
	tags := []*MerchantTag{}
	err := s.OrderBy("category ASC, sort_order ASC, title ASC").Find(&tags)
	if err != nil {
		return nil, err
	}

	// Group by category
	result := make(map[string][]*MerchantTag)
	for _, tag := range tags {
		category := tag.Category
		if category == "" {
			category = "uncategorized"
		}
		result[category] = append(result[category], tag)
	}

	return result, nil
}

// GetMerchantTagsByMerchantID returns all tags associated with a specific merchant
func GetMerchantTagsByMerchantID(s *xorm.Session, merchantID int64) ([]*MerchantTag, error) {
	tags := []*MerchantTag{}
	err := s.
		Join("INNER", "merchant_merchant_tags", "merchant_tags.id = merchant_merchant_tags.tag_id").
		Where("merchant_merchant_tags.merchant_id = ?", merchantID).
		OrderBy("merchant_tags.category ASC, merchant_tags.sort_order ASC, merchant_tags.title ASC").
		Find(&tags)
	
	return tags, err
}

// AssociateMerchantWithTags associates a merchant with multiple tags
func AssociateMerchantWithTags(s *xorm.Session, merchantID int64, tagIDs []int64) error {
	// First, remove existing associations
	_, err := s.Where("merchant_id = ?", merchantID).Delete(&MerchantMerchantTag{})
	if err != nil {
		return err
	}

	// Add new associations
	for _, tagID := range tagIDs {
		association := &MerchantMerchantTag{
			MerchantID: merchantID,
			TagID:      tagID,
		}
		_, err = s.Insert(association)
		if err != nil {
			return err
		}
	}

	return nil
}

// CreateSystemMerchantTags creates predefined system tags
func CreateSystemMerchantTags(s *xorm.Session, userID int64) error {
	systemTags := []MerchantTag{
		// Location-based tags
		{Title: "市中心", Description: "位于市中心区域", Category: "location", HexColor: "3498db", IsSystem: true, SortOrder: 1, CreatedByID: userID},
		{Title: "商业区", Description: "位于商业区域", Category: "location", HexColor: "2ecc71", IsSystem: true, SortOrder: 2, CreatedByID: userID},
		{Title: "住宅区", Description: "位于住宅区域", Category: "location", HexColor: "f39c12", IsSystem: true, SortOrder: 3, CreatedByID: userID},
		{Title: "工业区", Description: "位于工业区域", Category: "location", HexColor: "95a5a6", IsSystem: true, SortOrder: 4, CreatedByID: userID},
		
		// Business type tags
		{Title: "餐饮", Description: "餐饮类商户", Category: "type", HexColor: "e74c3c", IsSystem: true, SortOrder: 1, CreatedByID: userID},
		{Title: "零售", Description: "零售类商户", Category: "type", HexColor: "9b59b6", IsSystem: true, SortOrder: 2, CreatedByID: userID},
		{Title: "服务", Description: "服务类商户", Category: "type", HexColor: "1abc9c", IsSystem: true, SortOrder: 3, CreatedByID: userID},
		{Title: "娱乐", Description: "娱乐类商户", Category: "type", HexColor: "f1c40f", IsSystem: true, SortOrder: 4, CreatedByID: userID},
		
		// Size/scale tags
		{Title: "大型", Description: "大型商户", Category: "scale", HexColor: "34495e", IsSystem: true, SortOrder: 1, CreatedByID: userID},
		{Title: "中型", Description: "中型商户", Category: "scale", HexColor: "7f8c8d", IsSystem: true, SortOrder: 2, CreatedByID: userID},
		{Title: "小型", Description: "小型商户", Category: "scale", HexColor: "bdc3c7", IsSystem: true, SortOrder: 3, CreatedByID: userID},
		
		// Status tags
		{Title: "活跃", Description: "活跃营业中", Category: "status", HexColor: "27ae60", IsSystem: true, SortOrder: 1, CreatedByID: userID},
		{Title: "暂停", Description: "暂停营业", Category: "status", HexColor: "f39c12", IsSystem: true, SortOrder: 2, CreatedByID: userID},
		{Title: "关闭", Description: "已关闭", Category: "status", HexColor: "c0392b", IsSystem: true, SortOrder: 3, CreatedByID: userID},
	}

	for _, tag := range systemTags {
		// Check if tag already exists
		exists, err := s.Where("title = ? AND category = ? AND is_system = ?", tag.Title, tag.Category, true).Exist(&MerchantTag{})
		if err != nil {
			return err
		}
		if !exists {
			_, err = s.Insert(&tag)
			if err != nil {
				return err
			}
		}
	}

	return nil
}