package kit

import "context"

type Manager interface {
	UpdateStatus(ctx context.Context) error
}
