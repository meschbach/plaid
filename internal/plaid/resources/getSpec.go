package resources

import "context"

type getSpecReply struct {
	spec    []byte
	exists  bool
	problem error
}

type getSpecOp struct {
	meta    Meta
	replyTo chan<- getSpecReply
}

func (g *getSpecOp) name() string {
	return "get-spec"
}

func (g *getSpecOp) perform(ctx context.Context, c *Controller) {
	doReply := func(r getSpecReply) {
		select {
		case g.replyTo <- r:
		case <-ctx.Done():
		}
	}

	node, err := c.getNode(ctx, g.meta)
	if err != nil {
		doReply(getSpecReply{problem: err})
		return
	}
	if node == nil {
		doReply(getSpecReply{exists: false})
		return
	}
	doReply(getSpecReply{
		spec:    node.spec,
		exists:  true,
		problem: nil,
	})
}
