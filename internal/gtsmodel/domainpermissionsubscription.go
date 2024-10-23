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

type DomainPermissionSubscription struct {
	ID                 string               `bun:"type:CHAR(26),pk,nullzero,notnull,unique"`                    // ID of this item in the database.
	CreatedAt          time.Time            `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"` // Time when this item was created.
	Title              string               `bun:",nullzero"`                                                   // Moderator-set title for this list.
	PermissionType     DomainPermissionType `bun:",notnull"`                                                    // Permission type of the subscription.
	AsDraft            *bool                `bun:",nullzero,notnull,default:true"`                              // Create domain permission entries resulting from this subscription as drafts.
	CreatedByAccountID string               `bun:"type:CHAR(26),nullzero,notnull"`                              // Account ID of the creator of this subscription.
	CreatedByAccount   *Account             `bun:"-"`                                                           // Account corresponding to createdByAccountID.
	ContentType        string               `bun:",nullzero,notnull"`                                           // Content type to expect from the URI.
	URI                string               `bun:",unique,nullzero,notnull"`                                    // URI of the domain permission list.
	FetchUsername      string               `bun:",nullzero"`                                                   // Username to send when doing a GET of URI using basic auth.
	FetchPassword      string               `bun:",nullzero"`                                                   // Password to send when doing a GET of URI using basic auth.
	FetchedAt          time.Time            `bun:"type:timestamptz,nullzero"`                                   // Time when fetch of URI was last attempted.
	IsError            *bool                `bun:",nullzero,notnull,default:false"`                             // True if last fetch attempt of URI resulted in an error.
	Error              string               `bun:",nullzero"`                                                   // If IsError=true, this field contains the error resulting from the attempted fetch.
	Count              uint64               `bun:""`                                                            // Count of domain permission entries discovered at URI.
}
