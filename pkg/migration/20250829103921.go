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

type merchant20250829103921 struct {
	ID                  int64  `xorm:"bigint autoincr not null unique pk"`
	Title               string `xorm:"varchar(250) not null"`
	LegalRepresentative string `xorm:"varchar(250) null"`
	BusinessAddress     string `xorm:"varchar(500) null"`
	BusinessDistrict    string `xorm:"varchar(250) null"`
	ValidTime           string `xorm:"varchar(250) null"`
	TrafficConditions   string `xorm:"longtext null"`
	FixedEvents         string `xorm:"longtext null"`
	TerminalType        string `xorm:"varchar(250) null"`
	SpecialTimePeriods  string `xorm:"longtext null"`
	CustomFilters       string `xorm:"longtext null"`
	OwnerID             int64  `xorm:"bigint not null INDEX"`
}

func (merchant20250829103921) TableName() string {
	return "merchants"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20250829103921",
		Description: "Create merchants table",
		Migrate: func(tx *xorm.Engine) error {
			return tx.Sync2(merchant20250829103921{})
		},
		Rollback: func(tx *xorm.Engine) error {
			return tx.DropTables(merchant20250829103921{})
		},
	})
}