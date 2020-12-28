package errutil_test

import (
	"errors"
	"testing"

	"github.com/Nivl/git-go/internal/errutil"
	"github.com/stretchr/testify/assert"
)

type closer struct {
	actualClose func() error
}

func (c closer) Close() error {
	return c.actualClose()
}

func TestClose(t *testing.T) {
	t.Parallel()

	t.Run("Should call close and set the error", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("expected error")

		closed := false
		var err error
		c := closer{
			actualClose: func() error {
				closed = true
				return expectedErr
			},
		}

		errutil.Close(c, &err)
		assert.True(t, closed, "Close() should have been called")
		assert.Equal(t, expectedErr, err)
	})

	t.Run("Should call close and NOT set the error", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("expected error")
		closed := false
		err := expectedErr
		c := closer{
			actualClose: func() error {
				closed = true
				return errors.New("unexpected error")
			},
		}

		errutil.Close(c, &err)
		assert.True(t, closed, "Close() should have been called")
		assert.Equal(t, expectedErr, err)
	})
}
