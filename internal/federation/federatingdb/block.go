package federatingdb

import (
	"context"
	"fmt"

	"github.com/superseriousbusiness/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
)

func (f *federatingDB) Block(ctx context.Context, block vocab.ActivityStreamsBlock) error {
	receivingAccount, _ := extractFromCtx(ctx)
	if receivingAccount == nil {
		// If the receiving account wasn't set on the context, that means this request didn't pass
		// through the API, but came from inside GtS as the result of another activity on this instance. That being so,
		// we can safely just ignore this activity, since we know we've already processed it elsewhere.
		return nil
	}

	gtsBlock, err := f.typeConverter.ASBlockToBlock(ctx, block)
	if err != nil {
		return fmt.Errorf("Block: could not convert Block to gts model block")
	}

	gtsBlock.ID = id.NewULID()

	if err := f.state.DB.PutBlock(ctx, gtsBlock); err != nil {
		return fmt.Errorf("Block: database error inserting block: %s", err)
	}

	f.state.Workers.EnqueueFederator(ctx, messages.FromFederator{
		APObjectType:     ap.ActivityBlock,
		APActivityType:   ap.ActivityCreate,
		GTSModel:         block,
		ReceivingAccount: receivingAccount,
	})

	return nil
}
