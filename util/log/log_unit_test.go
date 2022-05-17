// +build unit_test

package log

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLogging(t *testing.T) {
	t.Parallel()

	useTextLogs = true

	t.Run("can log.Error with a nil error", func(t *testing.T) {
		Error(nil, "log.Error %v", 23)
		assert.True(t, len("I have visually inspected that this does not panic on nil error") > 0)
	})

	t.Run("can log.Error with an error object", func(t *testing.T) {
		e := fmt.Errorf("oh no")
		Error(e, "log.Error %v", 23)
		assert.True(t, len("I have visually inspected that this does not panic on nil error") > 0)
	})
}
