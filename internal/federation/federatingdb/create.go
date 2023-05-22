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

package federatingdb

import (
	"context"
	"errors"
	"fmt"

	"codeberg.org/gruf/go-kv"
	"codeberg.org/gruf/go-logger/v2/level"
	"github.com/superseriousbusiness/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
)

// Create adds a new entry to the database which must be able to be
// keyed by its id.
//
// Note that Activity values received from federated peers may also be
// created in the database this way if the Federating Protocol is
// enabled. The client may freely decide to store only the id instead of
// the entire value.
//
// The library makes this call only after acquiring a lock first.
//
// Under certain conditions and network activities, Create may be called
// multiple times for the same ActivityStreams object.
func (f *federatingDB) Create(ctx context.Context, asType vocab.Type) error {
	if log.Level() >= level.DEBUG {
		i, err := marshalItem(asType)
		if err != nil {
			return err
		}
		l := log.WithContext(ctx).
			WithField("create", i)
		l.Trace("entering Create")
	}

	receivingAccount, requestingAccount := extractFromCtx(ctx)
	if receivingAccount == nil {
		// If the receiving account wasn't set on the context, that means this request didn't pass
		// through the API, but came from inside GtS as the result of another activity on this instance. That being so,
		// we can safely just ignore this activity, since we know we've already processed it elsewhere.
		return nil
	}

	switch asType.GetTypeName() {
	case ap.ActivityCreate:
		// CREATE SOMETHING
		return f.activityCreate(ctx, asType, receivingAccount, requestingAccount)
	case ap.ActivityFollow:
		// FOLLOW SOMETHING
		return f.activityFollow(ctx, asType, receivingAccount, requestingAccount)
	case ap.ActivityLike:
		// LIKE SOMETHING
		return f.activityLike(ctx, asType, receivingAccount, requestingAccount)
	case ap.ActivityFlag:
		// FLAG / REPORT SOMETHING
		return f.activityFlag(ctx, asType, receivingAccount, requestingAccount)
	}
	return nil
}

/*
	BLOCK HANDLERS
*/

func (f *federatingDB) activityBlock(ctx context.Context, asType vocab.Type, receiving *gtsmodel.Account, requestingAccount *gtsmodel.Account) error {
	blockable, ok := asType.(vocab.ActivityStreamsBlock)
	if !ok {
		return errors.New("activityBlock: could not convert type to block")
	}

	block, err := f.typeConverter.ASBlockToBlock(ctx, blockable)
	if err != nil {
		return fmt.Errorf("activityBlock: could not convert Block to gts model block")
	}

	block.ID = id.NewULID()

	if err := f.state.DB.PutBlock(ctx, block); err != nil {
		return fmt.Errorf("activityBlock: database error inserting block: %s", err)
	}

	f.state.Workers.EnqueueFederator(ctx, messages.FromFederator{
		APObjectType:     ap.ActivityBlock,
		APActivityType:   ap.ActivityCreate,
		GTSModel:         block,
		ReceivingAccount: receiving,
	})
	return nil
}

/*
	CREATE HANDLERS
*/

func (f *federatingDB) activityCreate(ctx context.Context, asType vocab.Type, receivingAccount *gtsmodel.Account, requestingAccount *gtsmodel.Account) error {
	create, ok := asType.(vocab.ActivityStreamsCreate)
	if !ok {
		return errors.New("activityCreate: could resolve %T to Create")
	}

	// Create should have an object.
	object := create.GetActivityStreamsObject()
	if object == nil {
		return errors.New("Create had no Object")
	}

	// Iterate through the Object(s) to see what we're meant to be creating.
	errs := make(gtserror.MultiError, 0, object.Len())
	for iter := object.Begin(); iter != object.End(); iter = iter.Next() {
		objectType := iter.GetType()
		if objectType == nil {
			// Currently we can't do anything with just a Create
			// of something that's not an Object with a type.
			errs.Append(errors.New("object of Create was not a Type"))
			continue
		}

		// Process object according to its type.
		// TODO: possibly add more types here.
		switch typeName := objectType.GetTypeName(); typeName {
		case ap.ObjectNote:
			// CREATE A NOTE
			if err := f.createNote(
				ctx,
				iter.GetActivityStreamsNote(),
				receivingAccount,
				requestingAccount,
			); err != nil {
				errs.Append(err)
			}
		default:
			log.WithContext(ctx).
				WithFields(kv.Fields{
					{"receivingAccount", receivingAccount.URI},
					{"requestingAccount", requestingAccount.URI},
					{"typeName", typeName},
				}...).
				Debug("Object of Create was a type we couldn't handle")
		}
	}

	if len(errs) != 0 {
		return fmt.Errorf("activityCreate: one or more errors while processing activity: %w", errs.Combine())
	}

	return nil
}

// createNote handles a Create activity with a Note type.
func (f *federatingDB) createNote(ctx context.Context, note vocab.ActivityStreamsNote, receivingAccount *gtsmodel.Account, requestingAccount *gtsmodel.Account) error {
	// Note must have an attributedTo for us to process it.
	noteAttributedTo := note.GetActivityStreamsAttributedTo()
	if noteAttributedTo == nil {
		return errors.New("createNote: note had no attributedTo")
	}

	// Check if we have a forward. In other words, was the
	// note posted to our inbox by at least one actor who
	// created the note, or are they just forwarding it?
	forward := true

	// Compare the attributedTo(s) with the URI of the Actor
	// who posted this to our inbox. If the actor who posted
	// the Note to our inbox is the same as at least one of
	// the creators of the Note, then it's not a forward.
	for iter := noteAttributedTo.Begin(); iter != noteAttributedTo.End(); iter = iter.Next() {
		if !iter.IsIRI() {
			continue
		}

		if iri := iter.GetIRI(); iri != nil && iri.String() == requestingAccount.URI {
			forward = false
			break
		}
	}

	// If we do have a forward, we should ignore the content for
	// now and just dereference based on the URL/ID of the note
	// instead, to get the content straight from the poster's mouth.
	if forward {
		id := note.GetJSONLDId()
		if !id.IsIRI() {
			// If the ID isn't an IRI, then firstly this
			// is weird, and secondly we can't process it.
			return nil
		}

		// Process the Note asynchronously, we're done here.
		f.state.Workers.EnqueueFederator(ctx, messages.FromFederator{
			APObjectType:     ap.ObjectNote,
			APActivityType:   ap.ActivityCreate,
			APIri:            id.GetIRI(),
			APObjectModel:    nil,
			GTSModel:         nil,
			ReceivingAccount: receivingAccount,
		})

		return nil
	}

	// If we reach this point, we know the status wasn't forwarded
	// to us, but was delivered by at least one of the Actors who
	// created it, so proceed with processing it as normal.
	status, err := f.typeConverter.ASStatusToStatus(ctx, note)
	if err != nil {
		return fmt.Errorf("createNote: error converting note to status: %w", err)
	}

	// id the status based on the time it was created;
	// this allows for backdating of statuses.
	statusID, err := id.NewULIDFromTime(status.CreatedAt)
	if err != nil {
		return fmt.Errorf("createNote: error creating id for note: %w", err)
	}
	status.ID = statusID

	if err := f.state.DB.PutStatus(ctx, status); err != nil {
		if errors.Is(err, db.ErrAlreadyExists) {
			// The status already exists in the database, which
			// means we've already handled everything else, so
			// we can just return nil here and be done with it.
			return nil
		}

		// An actual error has happened.
		return fmt.Errorf("createNote: database error inserting status: %w", err)
	}

	// Do further processing asynchronously.
	f.state.Workers.EnqueueFederator(ctx, messages.FromFederator{
		APObjectType:     ap.ObjectNote,
		APActivityType:   ap.ActivityCreate,
		APObjectModel:    note,
		GTSModel:         status,
		ReceivingAccount: receivingAccount,
	})

	return nil
}

/*
	FOLLOW HANDLERS
*/

func (f *federatingDB) activityFollow(ctx context.Context, asType vocab.Type, receivingAccount *gtsmodel.Account, requestingAccount *gtsmodel.Account) error {
	follow, ok := asType.(vocab.ActivityStreamsFollow)
	if !ok {
		return errors.New("activityFollow: could not convert type to follow")
	}

	followRequest, err := f.typeConverter.ASFollowToFollowRequest(ctx, follow)
	if err != nil {
		return fmt.Errorf("activityFollow: could not convert Follow to follow request: %s", err)
	}

	followRequest.ID = id.NewULID()

	if err := f.state.DB.PutFollowRequest(ctx, followRequest); err != nil {
		return fmt.Errorf("activityFollow: database error inserting follow request: %s", err)
	}

	f.state.Workers.EnqueueFederator(ctx, messages.FromFederator{
		APObjectType:     ap.ActivityFollow,
		APActivityType:   ap.ActivityCreate,
		GTSModel:         followRequest,
		ReceivingAccount: receivingAccount,
	})

	return nil
}

/*
	LIKE HANDLERS
*/

func (f *federatingDB) activityLike(ctx context.Context, asType vocab.Type, receivingAccount *gtsmodel.Account, requestingAccount *gtsmodel.Account) error {
	like, ok := asType.(vocab.ActivityStreamsLike)
	if !ok {
		return errors.New("activityLike: could not convert type to like")
	}

	fave, err := f.typeConverter.ASLikeToFave(ctx, like)
	if err != nil {
		return fmt.Errorf("activityLike: could not convert Like to fave: %w", err)
	}

	fave.ID = id.NewULID()

	if err := f.state.DB.PutStatusFave(ctx, fave); err != nil {
		if errors.Is(err, db.ErrAlreadyExists) {
			// The Like already exists in the database, which
			// means we've already handled side effects. We can
			// just return nil here and be done with it.
			return nil
		}
		return fmt.Errorf("activityLike: database error inserting fave: %w", err)
	}

	f.state.Workers.EnqueueFederator(ctx, messages.FromFederator{
		APObjectType:     ap.ActivityLike,
		APActivityType:   ap.ActivityCreate,
		GTSModel:         fave,
		ReceivingAccount: receivingAccount,
	})

	return nil
}

/*
	FLAG HANDLERS
*/

func (f *federatingDB) activityFlag(ctx context.Context, asType vocab.Type, receivingAccount *gtsmodel.Account, requestingAccount *gtsmodel.Account) error {
	flag, ok := asType.(vocab.ActivityStreamsFlag)
	if !ok {
		return errors.New("activityFlag: could not convert type to flag")
	}

	report, err := f.typeConverter.ASFlagToReport(ctx, flag)
	if err != nil {
		return fmt.Errorf("activityFlag: could not convert Flag to report: %w", err)
	}

	report.ID = id.NewULID()

	if err := f.state.DB.PutReport(ctx, report); err != nil {
		return fmt.Errorf("activityFlag: database error inserting report: %w", err)
	}

	f.state.Workers.EnqueueFederator(ctx, messages.FromFederator{
		APObjectType:     ap.ActivityFlag,
		APActivityType:   ap.ActivityCreate,
		GTSModel:         report,
		ReceivingAccount: receivingAccount,
	})

	return nil
}
