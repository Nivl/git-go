package syncutil_test

import (
	"sync"
	"testing"
	"time"

	"github.com/Nivl/git-go/internal/syncutil"
	"github.com/stretchr/testify/assert"
)

func TestNamedMutex(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		a := []byte{'A'}
		b := []byte{'B'}

		mu := syncutil.NewNamedMutex(2)
		mu.Lock(a)
		mu.Lock(b)
		mu.Unlock(b)
		mu.Unlock(a)
	})

	t.Run("should still work with an invalid max", func(t *testing.T) {
		t.Parallel()

		a := []byte{'A'}
		b := []byte{'B'}

		mu := syncutil.NewNamedMutex(0)
		mu.Lock(a)
		mu.Lock(b)
		mu.Unlock(b)
		mu.Unlock(a)
	})

	t.Run("Lock should lock, Unlock should unlock", func(t *testing.T) {
		t.Parallel()

		mu := syncutil.NewNamedMutex(2)
		out := []string{}
		a := []byte{'A'}
		wg := sync.WaitGroup{}
		wg.Add(1)

		mu.Lock(a)

		go func() {
			mu.Lock(a)
			defer mu.Unlock(a)
			defer wg.Done()

			out = append(out, "goroutine")
		}()

		// we wait a long time to make sure the go-routine has locked
		time.Sleep(300 * time.Millisecond)
		out = append(out, "main")
		mu.Unlock(a)

		wg.Wait()
		assert.Equal(t, "main", out[0])
		assert.Equal(t, "goroutine", out[1])
	})
}
