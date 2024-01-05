package resources

import "context"

type deleteResult struct {
	exists  bool
	problem error
}

type deleteOp struct {
	replyTo chan deleteResult
	what    Meta
}

func (d *deleteOp) name() string {
	return "delete"
}

func (d *deleteOp) perform(ctx context.Context, c *Controller) {
	exists, err := c.deleteNode(ctx, d.what)
	//dispatch changes
	c.dispatchUpdates(ctx, DeletedEvent, d.what)

	select {
	case d.replyTo <- deleteResult{
		exists:  exists,
		problem: err,
	}:
	case <-ctx.Done():
	}
}
