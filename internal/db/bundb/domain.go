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

package bundb

import (
	"context"
	"errors"
	"net/url"
	"slices"

	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/paging"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/superseriousbusiness/gotosocial/internal/util"
	"github.com/uptrace/bun"
)

type domainDB struct {
	db    *bun.DB
	state *state.State
}

func (d *domainDB) CreateDomainAllow(ctx context.Context, allow *gtsmodel.DomainAllow) error {
	// Normalize the domain as punycode
	var err error
	allow.Domain, err = util.Punify(allow.Domain)
	if err != nil {
		return err
	}

	// Attempt to store domain allow in DB
	if _, err := d.db.NewInsert().
		Model(allow).
		Exec(ctx); err != nil {
		return err
	}

	// Clear the domain allow cache (for later reload)
	d.state.Caches.DB.DomainAllow.Clear()

	return nil
}

func (d *domainDB) GetDomainAllow(ctx context.Context, domain string) (*gtsmodel.DomainAllow, error) {
	// Normalize the domain as punycode
	domain, err := util.Punify(domain)
	if err != nil {
		return nil, err
	}

	// Check for easy case, domain referencing *us*
	if domain == "" || domain == config.GetAccountDomain() ||
		domain == config.GetHost() {
		return nil, db.ErrNoEntries
	}

	var allow gtsmodel.DomainAllow

	// Look for allow matching domain in DB
	q := d.db.
		NewSelect().
		Model(&allow).
		Where("? = ?", bun.Ident("domain_allow.domain"), domain)
	if err := q.Scan(ctx); err != nil {
		return nil, err
	}

	return &allow, nil
}

func (d *domainDB) GetDomainAllows(ctx context.Context) ([]*gtsmodel.DomainAllow, error) {
	allows := []*gtsmodel.DomainAllow{}

	if err := d.db.
		NewSelect().
		Model(&allows).
		Scan(ctx); err != nil {
		return nil, err
	}

	return allows, nil
}

func (d *domainDB) GetDomainAllowByID(ctx context.Context, id string) (*gtsmodel.DomainAllow, error) {
	var allow gtsmodel.DomainAllow

	q := d.db.
		NewSelect().
		Model(&allow).
		Where("? = ?", bun.Ident("domain_allow.id"), id)
	if err := q.Scan(ctx); err != nil {
		return nil, err
	}

	return &allow, nil
}

func (d *domainDB) DeleteDomainAllow(ctx context.Context, domain string) error {
	// Normalize the domain as punycode
	domain, err := util.Punify(domain)
	if err != nil {
		return err
	}

	// Attempt to delete domain allow
	if _, err := d.db.NewDelete().
		Model((*gtsmodel.DomainAllow)(nil)).
		Where("? = ?", bun.Ident("domain_allow.domain"), domain).
		Exec(ctx); err != nil {
		return err
	}

	// Clear the domain allow cache (for later reload)
	d.state.Caches.DB.DomainAllow.Clear()

	return nil
}

func (d *domainDB) CreateDomainBlock(ctx context.Context, block *gtsmodel.DomainBlock) error {
	// Normalize the domain as punycode
	var err error
	block.Domain, err = util.Punify(block.Domain)
	if err != nil {
		return err
	}

	// Attempt to store domain block in DB
	if _, err := d.db.NewInsert().
		Model(block).
		Exec(ctx); err != nil {
		return err
	}

	// Clear the domain block cache (for later reload)
	d.state.Caches.DB.DomainBlock.Clear()

	return nil
}

func (d *domainDB) GetDomainBlock(ctx context.Context, domain string) (*gtsmodel.DomainBlock, error) {
	// Normalize the domain as punycode
	domain, err := util.Punify(domain)
	if err != nil {
		return nil, err
	}

	// Check for easy case, domain referencing *us*
	if domain == "" || domain == config.GetAccountDomain() ||
		domain == config.GetHost() {
		return nil, db.ErrNoEntries
	}

	var block gtsmodel.DomainBlock

	// Look for block matching domain in DB
	q := d.db.
		NewSelect().
		Model(&block).
		Where("? = ?", bun.Ident("domain_block.domain"), domain)
	if err := q.Scan(ctx); err != nil {
		return nil, err
	}

	return &block, nil
}

func (d *domainDB) GetDomainBlocks(ctx context.Context) ([]*gtsmodel.DomainBlock, error) {
	blocks := []*gtsmodel.DomainBlock{}

	if err := d.db.
		NewSelect().
		Model(&blocks).
		Scan(ctx); err != nil {
		return nil, err
	}

	return blocks, nil
}

func (d *domainDB) GetDomainBlockByID(ctx context.Context, id string) (*gtsmodel.DomainBlock, error) {
	var block gtsmodel.DomainBlock

	q := d.db.
		NewSelect().
		Model(&block).
		Where("? = ?", bun.Ident("domain_block.id"), id)
	if err := q.Scan(ctx); err != nil {
		return nil, err
	}

	return &block, nil
}

func (d *domainDB) DeleteDomainBlock(ctx context.Context, domain string) error {
	// Normalize the domain as punycode
	domain, err := util.Punify(domain)
	if err != nil {
		return err
	}

	// Attempt to delete domain block
	if _, err := d.db.NewDelete().
		Model((*gtsmodel.DomainBlock)(nil)).
		Where("? = ?", bun.Ident("domain_block.domain"), domain).
		Exec(ctx); err != nil {
		return err
	}

	// Clear the domain block cache (for later reload)
	d.state.Caches.DB.DomainBlock.Clear()

	return nil
}

func (d *domainDB) IsDomainBlocked(ctx context.Context, domain string) (bool, error) {
	// Normalize the domain as punycode
	domain, err := util.Punify(domain)
	if err != nil {
		return false, err
	}

	// Domain referencing *us* cannot be blocked.
	if domain == "" || domain == config.GetAccountDomain() ||
		domain == config.GetHost() {
		return false, nil
	}

	// Check the cache for an explicit domain allow (hydrating the cache with callback if necessary).
	explicitAllow, err := d.state.Caches.DB.DomainAllow.Matches(domain, func() ([]string, error) {
		var domains []string

		// Scan list of all explicitly allowed domains from DB
		q := d.db.NewSelect().
			Table("domain_allows").
			Column("domain")
		if err := q.Scan(ctx, &domains); err != nil {
			return nil, err
		}

		return domains, nil
	})
	if err != nil {
		return false, err
	}

	// Check the cache for a domain block (hydrating the cache with callback if necessary)
	explicitBlock, err := d.state.Caches.DB.DomainBlock.Matches(domain, func() ([]string, error) {
		var domains []string

		// Scan list of all blocked domains from DB
		q := d.db.NewSelect().
			Table("domain_blocks").
			Column("domain")
		if err := q.Scan(ctx, &domains); err != nil {
			return nil, err
		}

		return domains, nil
	})
	if err != nil {
		return false, err
	}

	// Calculate if blocked
	// based on federation mode.
	switch mode := config.GetInstanceFederationMode(); mode {

	case config.InstanceFederationModeBlocklist:
		// Blocklist/default mode: explicit allow
		// takes precedence over explicit block.
		//
		// Domains that have neither block
		// or allow entries are allowed.
		return !(explicitAllow || !explicitBlock), nil

	case config.InstanceFederationModeAllowlist:
		// Allowlist mode: explicit block takes
		// precedence over explicit allow.
		//
		// Domains that have neither block
		// or allow entries are blocked.
		return (explicitBlock || !explicitAllow), nil

	default:
		// This should never happen but account
		// for it anyway to make the code tidier.
		return false, gtserror.Newf("unrecognized federation mode: %s", mode)
	}
}

func (d *domainDB) AreDomainsBlocked(ctx context.Context, domains []string) (bool, error) {
	for _, domain := range domains {
		if blocked, err := d.IsDomainBlocked(ctx, domain); err != nil {
			return false, err
		} else if blocked {
			return blocked, nil
		}
	}
	return false, nil
}

func (d *domainDB) IsURIBlocked(ctx context.Context, uri *url.URL) (bool, error) {
	return d.IsDomainBlocked(ctx, uri.Hostname())
}

func (d *domainDB) AreURIsBlocked(ctx context.Context, uris []*url.URL) (bool, error) {
	for _, uri := range uris {
		if blocked, err := d.IsDomainBlocked(ctx, uri.Hostname()); err != nil {
			return false, err
		} else if blocked {
			return blocked, nil
		}
	}
	return false, nil
}

func (d *domainDB) getDomainPermissionDraft(
	ctx context.Context,
	lookup string,
	dbQuery func(*gtsmodel.DomainPermissionDraft) error,
	keyParts ...any,
) (*gtsmodel.DomainPermissionDraft, error) {
	// Fetch perm draft from database cache with loader callback.
	permDraft, err := d.state.Caches.DB.DomainPermissionDraft.LoadOne(
		lookup,
		// Only called if not cached.
		func() (*gtsmodel.DomainPermissionDraft, error) {
			var permDraft gtsmodel.DomainPermissionDraft
			if err := dbQuery(&permDraft); err != nil {
				return nil, err
			}
			return &permDraft, nil
		},
		keyParts...,
	)
	if err != nil {
		return nil, err
	}

	if gtscontext.Barebones(ctx) {
		// No need to fully populate.
		return permDraft, nil
	}

	if permDraft.CreatedByAccount == nil {
		// Not set, fetch from database.
		permDraft.CreatedByAccount, err = d.state.DB.GetAccountByID(
			gtscontext.SetBarebones(ctx),
			permDraft.CreatedByAccountID,
		)
		if err != nil {
			return nil, gtserror.Newf("error populating created by account: %w", err)
		}
	}

	return permDraft, nil
}

func (d *domainDB) GetDomainPermissionDraftByID(
	ctx context.Context,
	id string,
) (*gtsmodel.DomainPermissionDraft, error) {
	return d.getDomainPermissionDraft(
		ctx,
		"ID",
		func(permDraft *gtsmodel.DomainPermissionDraft) error {
			return d.db.
				NewSelect().
				Model(permDraft).
				Where("? = ?", bun.Ident("domain_permission_draft.id"), id).
				Scan(ctx)
		},
		id,
	)
}

func (d *domainDB) GetDomainPermissionDrafts(
	ctx context.Context,
	permType *gtsmodel.DomainPermissionType,
	permSubID string,
	domain string,
	page *paging.Page,
) (
	[]*gtsmodel.DomainPermissionDraft,
	error,
) {
	var (
		// Get paging params.
		minID = page.GetMin()
		maxID = page.GetMax()
		limit = page.GetLimit()
		order = page.GetOrder()

		// Make educated guess for slice size
		permDraftIDs = make([]string, 0, limit)
	)

	q := d.db.
		NewSelect().
		TableExpr(
			"? AS ?",
			bun.Ident("domain_permission_drafts"),
			bun.Ident("domain_permission_draft"),
		).
		// Select only IDs from table
		Column("domain_permission_draft.id")

	// Return only items with id
	// lower than provided maxID.
	if maxID != "" {
		q = q.Where(
			"? < ?",
			bun.Ident("domain_permission_draft.id"),
			maxID,
		)
	}

	// Return only items with id
	// greater than provided minID.
	if minID != "" {
		q = q.Where(
			"? > ?",
			bun.Ident("domain_permission_draft.id"),
			minID,
		)
	}

	// Return only items with
	// given subscription ID.
	if permType != nil {
		q = q.Where(
			"? = ?",
			bun.Ident("domain_permission_draft.permission_type"),
			*permType,
		)
	}

	// Return only items with
	// given subscription ID.
	if permSubID != "" {
		q = q.Where(
			"? = ?",
			bun.Ident("domain_permission_draft.subscription_id"),
			permSubID,
		)
	}

	// Return only items
	// with given domain.
	if domain != "" {
		var err error

		// Normalize domain as punycode.
		domain, err = util.Punify(domain)
		if err != nil {
			return nil, gtserror.Newf("error punifying domain %s: %w", domain, err)
		}

		q = q.Where(
			"? = ?",
			bun.Ident("domain_permission_draft.domain"),
			domain,
		)
	}

	if limit > 0 {
		// Limit amount of
		// items returned.
		q = q.Limit(limit)
	}

	if order == paging.OrderAscending {
		// Page up.
		q = q.OrderExpr(
			"? ASC",
			bun.Ident("domain_permission_draft.id"),
		)
	} else {
		// Page down.
		q = q.OrderExpr(
			"? DESC",
			bun.Ident("domain_permission_draft.id"),
		)
	}

	if err := q.Scan(ctx, &permDraftIDs); err != nil {
		return nil, err
	}

	// Catch case of no items early
	if len(permDraftIDs) == 0 {
		return nil, db.ErrNoEntries
	}

	// If we're paging up, we still want items
	// to be sorted by ID desc, so reverse slice.
	if order == paging.OrderAscending {
		slices.Reverse(permDraftIDs)
	}

	// Allocate return slice (will be at most len permDraftIDs)
	permDrafts := make([]*gtsmodel.DomainPermissionDraft, 0, len(permDraftIDs))
	for _, id := range permDraftIDs {
		permDraft, err := d.GetDomainPermissionDraftByID(ctx, id)
		if err != nil {
			log.Errorf(ctx, "error getting domain permission draft %q: %v", id, err)
			continue
		}

		// Append to return slice
		permDrafts = append(permDrafts, permDraft)
	}

	return permDrafts, nil
}

func (d *domainDB) PutDomainPermissionDraft(
	ctx context.Context,
	permDraft *gtsmodel.DomainPermissionDraft,
) error {
	var err error

	// Normalize the domain as punycode
	permDraft.Domain, err = util.Punify(permDraft.Domain)
	if err != nil {
		return gtserror.Newf("error punifying domain %s: %w", permDraft.Domain, err)
	}

	return d.state.Caches.DB.DomainPermissionDraft.Store(
		permDraft,
		func() error {
			_, err := d.db.
				NewInsert().
				Model(permDraft).
				Exec(ctx)
			return err
		},
	)
}

func (d *domainDB) DeleteDomainPermissionDraft(
	ctx context.Context,
	id string,
) error {
	// Delete the permDraft from DB.
	q := d.db.NewDelete().
		TableExpr(
			"? AS ?",
			bun.Ident("domain_permission_drafts"),
			bun.Ident("domain_permission_draft"),
		).
		Where(
			"? = ?",
			bun.Ident("domain_permission_draft.id"),
			id,
		)

	_, err := q.Exec(ctx)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return err
	}

	// Invalidate any cached model by ID.
	d.state.Caches.DB.DomainPermissionDraft.Invalidate("ID", id)

	return nil
}

func (d *domainDB) getDomainPermissionSubscription(
	ctx context.Context,
	lookup string,
	dbQuery func(*gtsmodel.DomainPermissionSubscription) error,
	keyParts ...any,
) (*gtsmodel.DomainPermissionSubscription, error) {
	// Fetch perm subscription from database cache with loader callback.
	permDraft, err := d.state.Caches.DB.DomainPermissionSubscription.LoadOne(
		lookup,
		// Only called if not cached.
		func() (*gtsmodel.DomainPermissionSubscription, error) {
			var permDraft gtsmodel.DomainPermissionSubscription
			if err := dbQuery(&permDraft); err != nil {
				return nil, err
			}
			return &permDraft, nil
		},
		keyParts...,
	)
	if err != nil {
		return nil, err
	}

	if gtscontext.Barebones(ctx) {
		// No need to fully populate.
		return permDraft, nil
	}

	if permDraft.CreatedByAccount == nil {
		// Not set, fetch from database.
		permDraft.CreatedByAccount, err = d.state.DB.GetAccountByID(
			gtscontext.SetBarebones(ctx),
			permDraft.CreatedByAccountID,
		)
		if err != nil {
			return nil, gtserror.Newf("error populating created by account: %w", err)
		}
	}

	return permDraft, nil
}

func (d *domainDB) GetDomainPermissionSubscriptionByID(
	ctx context.Context,
	id string,
) (*gtsmodel.DomainPermissionSubscription, error) {
	return d.getDomainPermissionSubscription(
		ctx,
		"ID",
		func(permDraft *gtsmodel.DomainPermissionSubscription) error {
			return d.db.
				NewSelect().
				Model(permDraft).
				Where("? = ?", bun.Ident("domain_permission_subscription.id"), id).
				Scan(ctx)
		},
		id,
	)
}

func (d *domainDB) GetDomainPermissionSubscriptions(
	ctx context.Context,
	permType *gtsmodel.DomainPermissionType,
	page *paging.Page,
) (
	[]*gtsmodel.DomainPermissionSubscription,
	error,
) {
	var (
		// Get paging params.
		minID = page.GetMin()
		maxID = page.GetMax()
		limit = page.GetLimit()
		order = page.GetOrder()

		// Make educated guess for slice size
		permSubIDs = make([]string, 0, limit)
	)

	q := d.db.
		NewSelect().
		TableExpr(
			"? AS ?",
			bun.Ident("domain_permission_subscriptions"),
			bun.Ident("domain_permission_subscription"),
		).
		// Select only IDs from table
		Column("domain_permission_subscription.id")

	// Return only items with id
	// lower than provided maxID.
	if maxID != "" {
		q = q.Where(
			"? < ?",
			bun.Ident("domain_permission_subscription.id"),
			maxID,
		)
	}

	// Return only items with id
	// greater than provided minID.
	if minID != "" {
		q = q.Where(
			"? > ?",
			bun.Ident("domain_permission_subscription.id"),
			minID,
		)
	}

	// Return only items with
	// given subscription ID.
	if permType != nil {
		q = q.Where(
			"? = ?",
			bun.Ident("domain_permission_subscription.permission_type"),
			*permType,
		)
	}

	if limit > 0 {
		// Limit amount of
		// items returned.
		q = q.Limit(limit)
	}

	if order == paging.OrderAscending {
		// Page up.
		q = q.OrderExpr(
			"? ASC",
			bun.Ident("domain_permission_subscription.id"),
		)
	} else {
		// Page down.
		q = q.OrderExpr(
			"? DESC",
			bun.Ident("domain_permission_subscription.id"),
		)
	}

	if err := q.Scan(ctx, &permSubIDs); err != nil {
		return nil, err
	}

	// Catch case of no items early
	if len(permSubIDs) == 0 {
		return nil, db.ErrNoEntries
	}

	// If we're paging up, we still want items
	// to be sorted by ID desc, so reverse slice.
	if order == paging.OrderAscending {
		slices.Reverse(permSubIDs)
	}

	// Allocate return slice (will be at most len permSubIDs).
	permSubs := make([]*gtsmodel.DomainPermissionSubscription, 0, len(permSubIDs))
	for _, id := range permSubIDs {
		permDraft, err := d.GetDomainPermissionSubscriptionByID(ctx, id)
		if err != nil {
			log.Errorf(ctx, "error getting domain permission subscription %q: %v", id, err)
			continue
		}

		// Append to return slice
		permSubs = append(permSubs, permDraft)
	}

	return permSubs, nil
}

func (d *domainDB) PutDomainPermissionSubscription(
	ctx context.Context,
	permSubscription *gtsmodel.DomainPermissionSubscription,
) error {
	return d.state.Caches.DB.DomainPermissionSubscription.Store(
		permSubscription,
		func() error {
			_, err := d.db.
				NewInsert().
				Model(permSubscription).
				Exec(ctx)
			return err
		},
	)
}

func (d *domainDB) DeleteDomainPermissionSubscription(
	ctx context.Context,
	id string,
) error {
	// Delete the permSub from DB.
	q := d.db.NewDelete().
		TableExpr(
			"? AS ?",
			bun.Ident("domain_permission_subscriptions"),
			bun.Ident("domain_permission_subscription"),
		).
		Where(
			"? = ?",
			bun.Ident("domain_permission_subscription.id"),
			id,
		)

	_, err := q.Exec(ctx)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return err
	}

	// Invalidate any cached model by ID.
	d.state.Caches.DB.DomainPermissionSubscription.Invalidate("ID", id)

	return nil
}
