package logger

import (
	"github.com/meschbach/go-junk-bucket/pkg/streams"
	"sync"
)

type drain struct {
	changes     sync.Mutex
	id          int64
	offset      int64
	observatory *observedSink[bufferedEntry]
	stream      *streams.Buffer[bufferedEntry]
}
