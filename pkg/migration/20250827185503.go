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
		ID:          "20250827185503",
		Description: "Create geo_points table with PostGIS support",
		Migrate: func(tx *xorm.Engine) error {
			// Enable PostGIS extension if not already enabled
			_, err := tx.Exec("CREATE EXTENSION IF NOT EXISTS postgis")
			if err != nil {
				return err
			}

			// Create geo_points table
			createTableQuery := `
				CREATE TABLE geo_points (
					id BIGSERIAL PRIMARY KEY,
					merchant_id BIGINT NOT NULL,
					location GEOMETRY(POINT, 4326),
					original_address VARCHAR(500),
					formatted_address VARCHAR(500),
					accuracy_score DECIMAL(3,2) DEFAULT 0.0,
					geocoding_service VARCHAR(50),
					is_manual BOOLEAN DEFAULT FALSE,
					is_primary BOOLEAN DEFAULT TRUE,
					metadata TEXT,
					created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
					updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP
				)
			`
			_, err = tx.Exec(createTableQuery)
			if err != nil {
				return err
			}

			// Create indexes
			indexQueries := []string{
				"CREATE INDEX idx_geo_points_merchant ON geo_points(merchant_id)",
				"CREATE INDEX idx_geo_points_primary ON geo_points(merchant_id, is_primary)",
				"CREATE INDEX idx_geo_points_location ON geo_points USING GIST(location)",
				"CREATE INDEX idx_geo_points_service ON geo_points(geocoding_service)",
			}

			for _, query := range indexQueries {
				_, err := tx.Exec(query)
				if err != nil {
					return err
				}
			}

			// Add foreign key constraint
			fkQuery := "ALTER TABLE geo_points ADD CONSTRAINT fk_geo_points_merchant FOREIGN KEY (merchant_id) REFERENCES merchants(id) ON DELETE CASCADE"
			// Ignore errors for databases that don't support foreign keys
			tx.Exec(fkQuery)

			// Create trigger to update timestamp
			triggerQuery := `
				CREATE OR REPLACE FUNCTION update_geo_points_updated_column()
				RETURNS TRIGGER AS $$
				BEGIN
					NEW.updated = CURRENT_TIMESTAMP;
					RETURN NEW;
				END;
				$$ language 'plpgsql';

				CREATE TRIGGER update_geo_points_updated 
					BEFORE UPDATE ON geo_points 
					FOR EACH ROW EXECUTE FUNCTION update_geo_points_updated_column();
			`
			_, err = tx.Exec(triggerQuery)
			if err != nil {
				return err
			}

			return nil
		},
		Rollback: func(tx *xorm.Engine) error {
			// Drop trigger and function
			_, err := tx.Exec("DROP TRIGGER IF EXISTS update_geo_points_updated ON geo_points")
			if err != nil {
				return err
			}
			_, err = tx.Exec("DROP FUNCTION IF EXISTS update_geo_points_updated_column()")
			if err != nil {
				return err
			}

			// Drop table
			_, err = tx.Exec("DROP TABLE IF EXISTS geo_points")
			return err
		},
	})
}