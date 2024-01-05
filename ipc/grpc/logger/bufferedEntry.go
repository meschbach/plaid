package logger

import (
	"github.com/meschbach/plaid/internal/plaid/resources"
	"time"
)

type bufferedEntry struct {
	when       time.Time
	text       string
	from       resources.Meta
	streamName string
}
