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

		// TODO(melvin): add actual test once where the NewBlob actually
		// returns an OID
		sha := "37a85621591d08c17487c6fcfa4b20510c241952"
		data := "this is a fake content"

		oid, _ := plumbing.NewOidFromStr(sha)
		blob := object.NewBlob(oid, []byte(data))

		assert.Equal(t, 22, blob.Size())
		assert.Equal(t, []byte(data), blob.Bytes())
	})

	t.Run(".Bytes() should return immutable data", func(t *testing.T) {
		t.Parallel()

		// TODO(melvin): add actual test once where the NewBlob actually
		// returns an OID
		sha := "37a85621591d08c17487c6fcfa4b20510c241952"
		data := "this is a fake content"

		oid, _ := plumbing.NewOidFromStr(sha)
		blob := object.NewBlob(oid, []byte(data))

		assert.Equal(t, []byte(data), blob.Bytes())

		// We update the data, and make sure it hasn't actually
		// updates anything
		blob.Bytes()[0] = '0'
		assert.Equal(t, []byte(data), blob.Bytes())
	})
}
