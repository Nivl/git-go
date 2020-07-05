package object_test

import (
	"testing"

	"github.com/Nivl/git-go/plumbing"
	"github.com/Nivl/git-go/plumbing/object"
	"github.com/stretchr/testify/assert"
)

func TestBlob(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		data := "this is a fake content"
		o := object.New(object.TypeBlob, []byte(data))
		blob := object.NewBlob(o)

		assert.Equal(t, 22, blob.Size())
		assert.Equal(t, []byte(data), blob.Bytes())
		assert.Equal(t, []byte(data), blob.BytesCopy())
		assert.False(t, blob.IsPersisted())
		assert.Equal(t, plumbing.NullOid, blob.ID())

		assert.Equal(t, o, blob.AsObject())
	})

	t.Run(".BytesCopy() should return immutable data", func(t *testing.T) {
		t.Parallel()

		data := "this is a fake content"
		o := object.New(object.TypeBlob, []byte(data))
		blob := object.NewBlob(o)

		assert.Equal(t, []byte(data), blob.BytesCopy())

		// We update the data, and make sure it hasn't actually
		// updates anything
		blob.BytesCopy()[0] = '0'
		assert.Equal(t, []byte(data), blob.BytesCopy())
	})

	t.Run(".Bytes() should return mutable data", func(t *testing.T) {
		t.Parallel()

		data := "this is a fake content"
		expected := "0his is a fake content"
		o := object.New(object.TypeBlob, []byte(data))
		blob := object.NewBlob(o)

		assert.Equal(t, []byte(data), blob.Bytes())

		// We update the data, and make sure it hasn't actually
		// updates anything
		blob.Bytes()[0] = '0'
		assert.NotEqual(t, []byte(data), blob.Bytes())
		assert.Equal(t, expected, string(blob.Bytes()))
	})
}
