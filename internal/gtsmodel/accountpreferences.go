// GoToSocial
// Copyright (C) GoToSocial Authors admin@gotosocial.org
// SPDX-License-Identifier: AGPL-3.0-or-later
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
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package gtsmodel

import "time"

type AccountPreferences struct {
	ID                string     `bun:"type:CHAR(26),pk,nullzero,notnull,unique"`                    // ID of this item in the database.
	CreatedAt         time.Time  `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"` // Creation time of this item.
	UpdatedAt         time.Time  `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"` // Last updated time of this time.
	AccountID         string     `bun:"type:CHAR(26),pk,nullzero,notnull,unique"`                    // ID of the account to which these preferences correspond.
	StatusLanguage    string     `bun:",nullzero,notnull,default:'en'"`                              // Default post language for this account.
	StatusPrivacy     Visibility `bun:",nullzero"`                                                   // Default post privacy for this account.
	StatusSensitive   *bool      `bun:",nullzero,notnull,default:false"`                             // Set posts from this account to sensitive by default.
	StatusContentType string     `bun:",nullzero"`                                                   // Default format for statuses posted by this account.
	HideCollections   *bool      `bun:",nullzero,notnull,default:false"`                             // Hide this account's collections.
	EnableRSS         *bool      `bun:",nullzero,notnull,default:false"`                             // Enable RSS feed subscription for this account's public posts at [URL]/feed
	CustomCSS         string     `bun:",nullzero"`                                                   // Custom CSS that should be displayed for this Account's profile and statuses.
}
