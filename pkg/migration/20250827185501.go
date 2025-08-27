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

type merchantTag20250827185501 struct {
	ID          int64  `xorm:"bigint autoincr not null unique pk"`
	Title       string `xorm:"varchar(250) not null"`
	Description string `xorm:"longtext null"`
	HexColor    string `xorm:"varchar(6) null"`
	Category    string `xorm:"varchar(100) null"`
	IsSystem    bool   `xorm:"bool default false"`
	SortOrder   int    `xorm:"int default 0"`
	CreatedByID int64  `xorm:"bigint not null"`
	Created     string `xorm:"created not null"`
	Updated     string `xorm:"updated not null"`
}

func (merchantTag20250827185501) TableName() string {
	return "merchant_tags"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20250827185501",
		Description: "Create merchant_tags table",
		Migrate: func(tx *xorm.Engine) error {
			err := tx.CreateTables(merchantTag20250827185501{})
			if err != nil {
				return err
			}

			// Create indexes
			indexQueries := []string{
				"CREATE INDEX idx_merchant_tags_category ON merchant_tags(category)",
				"CREATE INDEX idx_merchant_tags_created_by ON merchant_tags(created_by_id)",
				"CREATE INDEX idx_merchant_tags_system ON merchant_tags(is_system)",
				"CREATE INDEX idx_merchant_tags_sort ON merchant_tags(category, sort_order, title)",
			}

			for _, query := range indexQueries {
				_, err := tx.Exec(query)
				if err != nil {
					return err
				}
			}

			return nil
		},
		Rollback: func(tx *xorm.Engine) error {
			return tx.DropTables(merchantTag20250827185501{})
		},
	})
}