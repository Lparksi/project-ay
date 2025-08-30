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
		ID:          "20241230120000",
		Description: "Add merchant mappings table for persistent field mappings",
		Migrate: func(tx *xorm.Engine) error {
			// Create merchant_mappings table with simple structure
			type MerchantMapping struct {
				ID          int64  `xorm:"bigint autoincr not null unique pk"`
				FieldName   string `xorm:"varchar(100) not null"`
				Placeholder string `xorm:"varchar(100) not null"`
				DisplayText string `xorm:"varchar(500) not null"`
				LabelID     int64  `xorm:"bigint not null"`
				IsActive    bool   `xorm:"bool not null default true"`
				OwnerID     int64  `xorm:"bigint not null"`
				Created     string `xorm:"datetime not null created"`
				Updated     string `xorm:"datetime not null updated"`
			}

			// Use Sync2 to create table - this is more reliable
			err := tx.Sync2(MerchantMapping{})
			if err != nil {
				return err
			}

			return nil
		},
		Rollback: func(tx *xorm.Engine) error {
			// Simple rollback - just drop the table
			return tx.DropTables("merchant_mappings")
		},
	})
}
