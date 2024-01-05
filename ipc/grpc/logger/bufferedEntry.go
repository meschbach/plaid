package logger

import (
	"github.com/meschbach/plaid/resources"
	"time"
)

type bufferedEntry struct {
	when       time.Time
	text       string
	from       resources.Meta
	streamName string
}
