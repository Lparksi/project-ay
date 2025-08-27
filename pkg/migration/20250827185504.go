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

package migration

import (
	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20250827185504",
		Description: "Insert system merchant tags",
		Migrate: func(tx *xorm.Engine) error {
			// Create a system user ID (use 1 as default, or create a system user)
			systemUserID := int64(1)

			// Insert system tags
			systemTags := []map[string]interface{}{
				// Location-based tags
				{"title": "市中心", "description": "位于市中心区域", "category": "location", "hex_color": "3498db", "is_system": true, "sort_order": 1, "created_by_id": systemUserID},
				{"title": "商业区", "description": "位于商业区域", "category": "location", "hex_color": "2ecc71", "is_system": true, "sort_order": 2, "created_by_id": systemUserID},
				{"title": "住宅区", "description": "位于住宅区域", "category": "location", "hex_color": "f39c12", "is_system": true, "sort_order": 3, "created_by_id": systemUserID},
				{"title": "工业区", "description": "位于工业区域", "category": "location", "hex_color": "95a5a6", "is_system": true, "sort_order": 4, "created_by_id": systemUserID},
				
				// Business type tags
				{"title": "餐饮", "description": "餐饮类商户", "category": "type", "hex_color": "e74c3c", "is_system": true, "sort_order": 1, "created_by_id": systemUserID},
				{"title": "零售", "description": "零售类商户", "category": "type", "hex_color": "9b59b6", "is_system": true, "sort_order": 2, "created_by_id": systemUserID},
				{"title": "服务", "description": "服务类商户", "category": "type", "hex_color": "1abc9c", "is_system": true, "sort_order": 3, "created_by_id": systemUserID},
				{"title": "娱乐", "description": "娱乐类商户", "category": "type", "hex_color": "f1c40f", "is_system": true, "sort_order": 4, "created_by_id": systemUserID},
				
				// Size/scale tags
				{"title": "大型", "description": "大型商户", "category": "scale", "hex_color": "34495e", "is_system": true, "sort_order": 1, "created_by_id": systemUserID},
				{"title": "中型", "description": "中型商户", "category": "scale", "hex_color": "7f8c8d", "is_system": true, "sort_order": 2, "created_by_id": systemUserID},
				{"title": "小型", "description": "小型商户", "category": "scale", "hex_color": "bdc3c7", "is_system": true, "sort_order": 3, "created_by_id": systemUserID},
				
				// Status tags
				{"title": "活跃", "description": "活跃营业中", "category": "status", "hex_color": "27ae60", "is_system": true, "sort_order": 1, "created_by_id": systemUserID},
				{"title": "暂停", "description": "暂停营业", "category": "status", "hex_color": "f39c12", "is_system": true, "sort_order": 2, "created_by_id": systemUserID},
				{"title": "关闭", "description": "已关闭", "category": "status", "hex_color": "c0392b", "is_system": true, "sort_order": 3, "created_by_id": systemUserID},
			}

			for _, tag := range systemTags {
				// Check if tag already exists
				var count int64
				count, err := tx.Where("title = ? AND category = ? AND is_system = ?", 
					tag["title"], tag["category"], true).Count(&merchantTag20250827185501{})
				if err != nil {
					return err
				}
				
				if count == 0 {
					// Insert the tag
					insertQuery := `
						INSERT INTO merchant_tags (title, description, category, hex_color, is_system, sort_order, created_by_id, created, updated)
						VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
					`
					_, err = tx.Exec(insertQuery, 
						tag["title"], tag["description"], tag["category"], 
						tag["hex_color"], tag["is_system"], tag["sort_order"], tag["created_by_id"])
					if err != nil {
						return err
					}
				}
			}

			return nil
		},
		Rollback: func(tx *xorm.Engine) error {
			// Delete system tags
			_, err := tx.Where("is_system = ?", true).Delete(&merchantTag20250827185501{})
			return err
		},
	})
}