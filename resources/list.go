package resources

import (
	"context"
)

type listOp struct {
	kind    Type
	replyTo chan Meta
}

func (l *listOp) name() string {
	return "list"
}

func (l *listOp) perform(ctx context.Context, c *Controller) {
	for _, name := range c.resources.ListNames(l.kind) {
		l.replyTo <- Meta{
			Type: l.kind,
			Name: name,
		}
	}
	close(l.replyTo)
}
