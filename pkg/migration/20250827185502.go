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

type merchantMerchantTag20250827185502 struct {
	ID         int64  `xorm:"bigint autoincr not null unique pk"`
	MerchantID int64  `xorm:"bigint not null"`
	TagID      int64  `xorm:"bigint not null"`
	Created    string `xorm:"created not null"`
}

func (merchantMerchantTag20250827185502) TableName() string {
	return "merchant_merchant_tags"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20250827185502",
		Description: "Create merchant_merchant_tags association table",
		Migrate: func(tx *xorm.Engine) error {
			err := tx.CreateTables(merchantMerchantTag20250827185502{})
			if err != nil {
				return err
			}

			// Create indexes and constraints
			queries := []string{
				"CREATE INDEX idx_merchant_merchant_tags_merchant ON merchant_merchant_tags(merchant_id)",
				"CREATE INDEX idx_merchant_merchant_tags_tag ON merchant_merchant_tags(tag_id)",
				"CREATE UNIQUE INDEX idx_merchant_merchant_tags_unique ON merchant_merchant_tags(merchant_id, tag_id)",
			}

			for _, query := range queries {
				_, err := tx.Exec(query)
				if err != nil {
					return err
				}
			}

			// Add foreign key constraints if supported
			fkQueries := []string{
				"ALTER TABLE merchant_merchant_tags ADD CONSTRAINT fk_merchant_merchant_tags_merchant FOREIGN KEY (merchant_id) REFERENCES merchants(id) ON DELETE CASCADE",
				"ALTER TABLE merchant_merchant_tags ADD CONSTRAINT fk_merchant_merchant_tags_tag FOREIGN KEY (tag_id) REFERENCES merchant_tags(id) ON DELETE CASCADE",
			}

			for _, query := range fkQueries {
				// Ignore errors for databases that don't support foreign keys
				tx.Exec(query)
			}

			return nil
		},
		Rollback: func(tx *xorm.Engine) error {
			return tx.DropTables(merchantMerchantTag20250827185502{})
		},
	})
}