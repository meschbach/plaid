package fsn

import (
	"github.com/meschbach/plaid/controllers/filewatch"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFileWatchCompliance(t *testing.T) {
	t.Run("Ensure interface compliance", func(t *testing.T) {
		fsn := NewCore()
		assert.Implements(t, (*filewatch.FileSystem)(nil), fsn)
	})
}
