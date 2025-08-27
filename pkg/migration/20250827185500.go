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
		ID:          "20250827185500",
		Description: "Add geographic fields to merchants table",
		Migrate: func(tx *xorm.Engine) error {
			// Add geographic columns to merchants table
			queries := []string{
				"ALTER TABLE merchants ADD COLUMN longitude DECIMAL(10,7) NULL",
				"ALTER TABLE merchants ADD COLUMN latitude DECIMAL(10,7) NULL", 
				"ALTER TABLE merchants ADD COLUMN geocode_accuracy DECIMAL(3,2) DEFAULT 0.0",
				"ALTER TABLE merchants ADD COLUMN geocode_address VARCHAR(500) NULL",
				"ALTER TABLE merchants ADD COLUMN geocode_service VARCHAR(50) NULL",
				"ALTER TABLE merchants ADD COLUMN is_manual_location BOOLEAN DEFAULT FALSE",
			}

			for _, query := range queries {
				_, err := tx.Exec(query)
				if err != nil {
					return err
				}
			}

			// Create indexes for geographic queries
			indexQueries := []string{
				"CREATE INDEX idx_merchants_longitude ON merchants(longitude)",
				"CREATE INDEX idx_merchants_latitude ON merchants(latitude)",
				"CREATE INDEX idx_merchants_location ON merchants(longitude, latitude)",
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
			// Drop indexes first
			dropIndexQueries := []string{
				"DROP INDEX IF EXISTS idx_merchants_longitude",
				"DROP INDEX IF EXISTS idx_merchants_latitude", 
				"DROP INDEX IF EXISTS idx_merchants_location",
			}

			for _, query := range dropIndexQueries {
				_, err := tx.Exec(query)
				if err != nil {
					return err
				}
			}

			// Drop columns
			dropColumnQueries := []string{
				"ALTER TABLE merchants DROP COLUMN IF EXISTS longitude",
				"ALTER TABLE merchants DROP COLUMN IF EXISTS latitude",
				"ALTER TABLE merchants DROP COLUMN IF EXISTS geocode_accuracy",
				"ALTER TABLE merchants DROP COLUMN IF EXISTS geocode_address",
				"ALTER TABLE merchants DROP COLUMN IF EXISTS geocode_service",
				"ALTER TABLE merchants DROP COLUMN IF EXISTS is_manual_location",
			}

			for _, query := range dropColumnQueries {
				_, err := tx.Exec(query)
				if err != nil {
					return err
				}
			}

			return nil
		},
	})
}