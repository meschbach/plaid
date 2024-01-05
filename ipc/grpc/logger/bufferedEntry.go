package logger

import (
	"git.meschbach.com/mee/platform.git/plaid/internal/plaid/resources"
	"time"
)

type bufferedEntry struct {
	when       time.Time
	text       string
	from       resources.Meta
	streamName string
}
