package object_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/Nivl/git-go/ginternals"
	"github.com/Nivl/git-go/ginternals/object"
	"github.com/Nivl/git-go/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAsCommit(t *testing.T) {
	t.Parallel()

	t.Run("regular commit with all the fields", func(t *testing.T) {
		t.Parallel()

		treeID, _ := ginternals.NewOidFromStr("f0b577644139c6e04216d82f1dd4a5a63addeeca")
		parentID, _ := ginternals.NewOidFromStr("9785af758bcc96cd7237ba65eb2c9dd1ecaa3321")

		var b bytes.Buffer
		b.WriteString("tree ")
		b.WriteString(treeID.String())
		b.WriteString("\n")

		b.WriteString("parent ")
		b.WriteString(parentID.String())
		b.WriteString("\n")

		b.WriteString(`author Melvin Laplanche <melvin.wont.reply@gmail.com> 1566115917 -0700
committer Melvin Laplanche <melvin.wont.reply@gmail.com> 1566115917 -0700
gpgsig -----BEGIN PGP SIGNATURE-----

 iQIzBAABCAAdFiEE9vjmBp5ZMl+LWBekLDB+DQQTNEsFAl1ZCE0ACgkQLDB+DQQT
 NEuyIQ/+P14N/BK8dnqnLcMhjoGS86fy14MCqo3hPJxPWl0Qw0JQ5APDRNqnPiT6
 7z25y7e+RqeRR6OnNQhK5Tgv34BGrXcLuqQqE+9QWSZZV6XzbBNwkPBp/ZgzncQh
 ZL6ywGD0LAYom3g+KuJpeeBdVZ7XCmh7a2sLYEQG2gmasU2CslRPdooMGZ4RvdLd
 KjiykE5wMKXH2/6TgI7sxGgFXni+63x3yF2gBcAQAPn6j3YpPPW8yBrYjYTfWS/G
 mNbluh0jwCWXeTCJof5eCO3WYvUpoAuG4JYMoVV3hxM/RbtbZxtdX5MKYIlEb2Un
 M4VY8RUkzXvvlMigQFO2BPP5JKD5ep3nVYqKpEiTc+Qx1pInq8iELGDni4H2dtPV
 DlFkiEs2Rdlxn17pEs6OWIlJtpCRcKUAg2ehyiiybqCaNYtTAWUO+/Ku0SnovLTp
 sTtvd466SP0GyC8WqqG223ljPwVgPOe/y5ZvRuUY+1CcT4I3iIE/wXcbw9ldZd51
 Tmvx/aZSXpRE8DvYsN4yQpeeJFNVaoTO0IRNf8AG8YQzchRUxdd1l0uy5o2evGXE
 /mZenHRSs/LNfYEwfNhJy6tPGAI9to/O15UHVRS1nneuacMSIyjxYg/kfhmSZKoz
 o9fizcxapx+JwVYHviO6wVdSbgS2aO1u9/whof3Fkm+/Luvo0J4=
 =/Zem
 -----END PGP SIGNATURE-----

commit head

commit body

commit footer`)
		rawData := b.Bytes()

		o := object.New(object.TypeCommit, rawData)
		expectedSigName := "Melvin Laplanche"
		expectedSigEmail := "melvin.wont.reply@gmail.com"
		expectedSigTimestamp := int64(1566115917)
		expectedSigOffset := 3600 * -7

		ci, err := o.AsCommit()
		require.NoError(t, err)

		assert.Equal(t, o.ID(), ci.ID())
		assert.Equal(t, "f0b577644139c6e04216d82f1dd4a5a63addeeca", ci.TreeID().String(), "invalid tree id")

		require.NotZero(t, ci.Author(), "author missing")
		assert.Equal(t, expectedSigName, ci.Author().Name, "invalid author name")
		assert.Equal(t, expectedSigEmail, ci.Author().Email, "invalid author email")
		assert.Equal(t, expectedSigTimestamp, ci.Author().Time.Unix(), "invalid author timestamp")
		_, tzOffset := ci.Committer().Time.Zone()
		assert.Equal(t, expectedSigOffset, tzOffset, "invalid author timezone offset")

		require.NotZero(t, ci.Committer(), "committer missing")
		assert.Equal(t, expectedSigName, ci.Committer().Name, "invalid committer name")
		assert.Equal(t, expectedSigEmail, ci.Committer().Email, "invalid committer email")
		assert.Equal(t, expectedSigTimestamp, ci.Committer().Time.Unix(), "invalid committer timestamp")
		_, tzOffset = ci.Committer().Time.Zone()
		assert.Equal(t, expectedSigOffset, tzOffset, "invalid committer timezone offset")

		require.Len(t, ci.ParentIDs(), 1, "invalid amount of parent")
		assert.Equal(t, "9785af758bcc96cd7237ba65eb2c9dd1ecaa3321", ci.ParentIDs()[0].String(), "invalid parent id")

		expectedGPG := `-----BEGIN PGP SIGNATURE-----

 iQIzBAABCAAdFiEE9vjmBp5ZMl+LWBekLDB+DQQTNEsFAl1ZCE0ACgkQLDB+DQQT
 NEuyIQ/+P14N/BK8dnqnLcMhjoGS86fy14MCqo3hPJxPWl0Qw0JQ5APDRNqnPiT6
 7z25y7e+RqeRR6OnNQhK5Tgv34BGrXcLuqQqE+9QWSZZV6XzbBNwkPBp/ZgzncQh
 ZL6ywGD0LAYom3g+KuJpeeBdVZ7XCmh7a2sLYEQG2gmasU2CslRPdooMGZ4RvdLd
 KjiykE5wMKXH2/6TgI7sxGgFXni+63x3yF2gBcAQAPn6j3YpPPW8yBrYjYTfWS/G
 mNbluh0jwCWXeTCJof5eCO3WYvUpoAuG4JYMoVV3hxM/RbtbZxtdX5MKYIlEb2Un
 M4VY8RUkzXvvlMigQFO2BPP5JKD5ep3nVYqKpEiTc+Qx1pInq8iELGDni4H2dtPV
 DlFkiEs2Rdlxn17pEs6OWIlJtpCRcKUAg2ehyiiybqCaNYtTAWUO+/Ku0SnovLTp
 sTtvd466SP0GyC8WqqG223ljPwVgPOe/y5ZvRuUY+1CcT4I3iIE/wXcbw9ldZd51
 Tmvx/aZSXpRE8DvYsN4yQpeeJFNVaoTO0IRNf8AG8YQzchRUxdd1l0uy5o2evGXE
 /mZenHRSs/LNfYEwfNhJy6tPGAI9to/O15UHVRS1nneuacMSIyjxYg/kfhmSZKoz
 o9fizcxapx+JwVYHviO6wVdSbgS2aO1u9/whof3Fkm+/Luvo0J4=
 =/Zem
 -----END PGP SIGNATURE-----`
		assert.Equal(t, expectedGPG, ci.GPGSig(), "invalid gpgsig")

		expectedMessage := `commit head

commit body

commit footer`
		assert.Equal(t, expectedMessage, ci.Message(), "invalid Message")
	})
}

func TestAsTree(t *testing.T) {
	t.Parallel()

	t.Run("regular tree", func(t *testing.T) {
		t.Parallel()

		treeSHA := "e5b9e846e1b468bc9597ff95d71dfacda8bd54e3"

		testFile := fmt.Sprintf("tree_%s", treeSHA)
		content, err := os.ReadFile(filepath.Join(testutil.TestdataPath(t), testFile))
		require.NoError(t, err)

		o := object.New(object.TypeTree, content)
		tree, err := o.AsTree()
		require.NoError(t, err)

		assert.Equal(t, o.ID(), tree.ID())
		assert.Len(t, tree.Entries(), 13)
	})
}

func TestAsBlob(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(filepath.Join(testutil.TestdataPath(t), "blob_642480605b8b0fd464ab5762e044269cf29a60a3"))
	require.NoError(t, err)

	o := object.New(object.TypeBlob, content)
	blob := o.AsBlob()

	assert.Equal(t, o.ID(), blob.ID())
	assert.Equal(t, o.Size(), blob.Size())
	assert.Equal(t, o.Bytes(), blob.Bytes())
}

func TestType(t *testing.T) {
	t.Parallel()

	t.Run("type.String()", func(t *testing.T) {
		t.Parallel()

		testCases := []struct {
			desc           string
			typ            object.Type
			expected       string
			expectsFailure bool
		}{
			{
				desc:     "a commit should be displayed at commit",
				typ:      object.TypeCommit,
				expected: "commit",
			},
			{
				desc:     "a tree should be displayed at tree",
				typ:      object.TypeTree,
				expected: "tree",
			},
			{
				desc:     "a blob should be displayed at blob",
				typ:      object.TypeBlob,
				expected: "blob",
			},
			{
				desc:     "a tag should be displayed at tag",
				typ:      object.TypeTag,
				expected: "tag",
			},
			{
				desc:     "a osf-delta should be displayed at osf-delta",
				typ:      object.ObjectDeltaOFS,
				expected: "osf-delta",
			},
			{
				desc:     "a ref-delta should be displayed at ref-delta",
				typ:      object.ObjectDeltaRef,
				expected: "ref-delta",
			},
			{
				desc:           "Invalid type should panic",
				typ:            object.Type(5),
				expectsFailure: true,
			},
		}
		for i, tc := range testCases {
			tc := tc
			t.Run(fmt.Sprintf("%d/%s", i, tc.desc), func(t *testing.T) {
				t.Parallel()

				if tc.expectsFailure {
					assert.Panics(t, func() {
						tc.typ.String() //nolint:govet // we just want a panic
					})
					return
				}
				assert.Equal(t, tc.expected, tc.typ.String())
			})
		}
	})

	t.Run("type.IsValid()", func(t *testing.T) {
		t.Parallel()

		// sugars
		valid := true
		invalid := false
		testCases := []struct {
			desc     string
			typ      object.Type
			expected bool
		}{
			{
				desc:     "TypeCommit should be valid",
				typ:      object.TypeCommit,
				expected: valid,
			},
			{
				desc:     "TypeTree should be valid",
				typ:      object.TypeTree,
				expected: valid,
			},
			{
				desc:     "TypeBlob should be valid",
				typ:      object.TypeBlob,
				expected: valid,
			},
			{
				desc:     "TypeTag should be valid",
				typ:      object.TypeTag,
				expected: valid,
			},
			{
				desc:     "ObjectDeltaOFS should be valid",
				typ:      object.ObjectDeltaOFS,
				expected: valid,
			},
			{
				desc:     "ObjectDeltaRef should be valid",
				typ:      object.ObjectDeltaRef,
				expected: valid,
			},
			{
				desc:     "Invalid type should be invalid",
				typ:      object.Type(5),
				expected: invalid,
			},
		}
		for i, tc := range testCases {
			tc := tc
			t.Run(fmt.Sprintf("%d/%s", i, tc.desc), func(t *testing.T) {
				t.Parallel()

				assert.Equal(t, tc.expected, tc.typ.IsValid())
			})
		}
	})

	t.Run("NewTypeFromString", func(t *testing.T) {
		t.Parallel()

		testCases := []struct {
			desc           string
			typ            string
			expected       object.Type
			expectsFailure bool
		}{
			{
				desc:     "TypeCommit should be valid",
				typ:      "commit",
				expected: object.TypeCommit,
			},
			{
				desc:     "TypeTree should be valid",
				typ:      "tree",
				expected: object.TypeTree,
			},
			{
				desc:     "TypeBlob should be valid",
				typ:      "blob",
				expected: object.TypeBlob,
			},
			{
				desc:     "TypeTag should be valid",
				typ:      "tag",
				expected: object.TypeTag,
			},
			{
				desc:           "Invalid type should be invalid",
				typ:            "doesnt-exists",
				expectsFailure: true,
			},
		}
		for i, tc := range testCases {
			tc := tc
			t.Run(fmt.Sprintf("%d/%s", i, tc.desc), func(t *testing.T) {
				t.Parallel()

				out, err := object.NewTypeFromString(tc.typ)
				if tc.expectsFailure {
					require.Equal(t, object.ErrObjectUnknown, err)
					return
				}

				assert.Equal(t, tc.expected, out)
			})
		}
	})
}

func TestCompress(t *testing.T) {
	t.Parallel()

	t.Run("tree", func(t *testing.T) {
		t.Parallel()

		treeSHA := "e5b9e846e1b468bc9597ff95d71dfacda8bd54e3"

		testFile := fmt.Sprintf("tree_%s", treeSHA)
		content, err := os.ReadFile(filepath.Join(testutil.TestdataPath(t), testFile))
		require.NoError(t, err)

		o := object.New(object.TypeTree, content)
		_, err = o.Compress()
		require.NoError(t, err)
		assert.Equal(t, treeSHA, o.ID().String())

		// TODO(melvin): Test the compressed object
	})
}

func TestAsTag(t *testing.T) {
	t.Parallel()

	t.Run("regular tag with all the fields", func(t *testing.T) {
		t.Parallel()

		objID, _ := ginternals.NewOidFromStr("9785af758bcc96cd7237ba65eb2c9dd1ecaa3321")

		var b bytes.Buffer
		b.WriteString("object ")
		b.WriteString(objID.String())
		b.WriteString("\n")

		b.WriteString(`type commit
tag tag.name
tagger Melvin Laplanche <melvin.wont.reply@gmail.com> 1566115917 -0700
gpgsig -----BEGIN PGP SIGNATURE-----

 iQIzBAABCAAdFiEE9vjmBp5ZMl+LWBekLDB+DQQTNEsFAl1ZCE0ACgkQLDB+DQQT
 NEuyIQ/+P14N/BK8dnqnLcMhjoGS86fy14MCqo3hPJxPWl0Qw0JQ5APDRNqnPiT6
 7z25y7e+RqeRR6OnNQhK5Tgv34BGrXcLuqQqE+9QWSZZV6XzbBNwkPBp/ZgzncQh
 ZL6ywGD0LAYom3g+KuJpeeBdVZ7XCmh7a2sLYEQG2gmasU2CslRPdooMGZ4RvdLd
 KjiykE5wMKXH2/6TgI7sxGgFXni+63x3yF2gBcAQAPn6j3YpPPW8yBrYjYTfWS/G
 mNbluh0jwCWXeTCJof5eCO3WYvUpoAuG4JYMoVV3hxM/RbtbZxtdX5MKYIlEb2Un
 M4VY8RUkzXvvlMigQFO2BPP5JKD5ep3nVYqKpEiTc+Qx1pInq8iELGDni4H2dtPV
 DlFkiEs2Rdlxn17pEs6OWIlJtpCRcKUAg2ehyiiybqCaNYtTAWUO+/Ku0SnovLTp
 sTtvd466SP0GyC8WqqG223ljPwVgPOe/y5ZvRuUY+1CcT4I3iIE/wXcbw9ldZd51
 Tmvx/aZSXpRE8DvYsN4yQpeeJFNVaoTO0IRNf8AG8YQzchRUxdd1l0uy5o2evGXE
 /mZenHRSs/LNfYEwfNhJy6tPGAI9to/O15UHVRS1nneuacMSIyjxYg/kfhmSZKoz
 o9fizcxapx+JwVYHviO6wVdSbgS2aO1u9/whof3Fkm+/Luvo0J4=
 =/Zem
 -----END PGP SIGNATURE-----

tag message`)
		rawData := b.Bytes()

		o := object.New(object.TypeTag, rawData)
		expectedSigName := "Melvin Laplanche"
		expectedSigEmail := "melvin.wont.reply@gmail.com"
		expectedSigTimestamp := int64(1566115917)
		expectedSigOffset := 3600 * -7

		tag, err := o.AsTag()
		require.NoError(t, err)

		assert.Equal(t, o.ID(), tag.ID())
		assert.Equal(t, objID, tag.Target())

		require.NotZero(t, tag.Tagger(), "tagger missing")
		assert.Equal(t, expectedSigName, tag.Tagger().Name, "invalid tagger name")
		assert.Equal(t, expectedSigEmail, tag.Tagger().Email, "invalid tagger email")
		assert.Equal(t, expectedSigTimestamp, tag.Tagger().Time.Unix(), "invalid tagger timestamp")
		_, tzOffset := tag.Tagger().Time.Zone()
		assert.Equal(t, expectedSigOffset, tzOffset, "invalid tagger timezone offset")

		assert.Equal(t, object.TypeCommit, tag.Type(), "invalid commit type")

		expectedGPG := `-----BEGIN PGP SIGNATURE-----

 iQIzBAABCAAdFiEE9vjmBp5ZMl+LWBekLDB+DQQTNEsFAl1ZCE0ACgkQLDB+DQQT
 NEuyIQ/+P14N/BK8dnqnLcMhjoGS86fy14MCqo3hPJxPWl0Qw0JQ5APDRNqnPiT6
 7z25y7e+RqeRR6OnNQhK5Tgv34BGrXcLuqQqE+9QWSZZV6XzbBNwkPBp/ZgzncQh
 ZL6ywGD0LAYom3g+KuJpeeBdVZ7XCmh7a2sLYEQG2gmasU2CslRPdooMGZ4RvdLd
 KjiykE5wMKXH2/6TgI7sxGgFXni+63x3yF2gBcAQAPn6j3YpPPW8yBrYjYTfWS/G
 mNbluh0jwCWXeTCJof5eCO3WYvUpoAuG4JYMoVV3hxM/RbtbZxtdX5MKYIlEb2Un
 M4VY8RUkzXvvlMigQFO2BPP5JKD5ep3nVYqKpEiTc+Qx1pInq8iELGDni4H2dtPV
 DlFkiEs2Rdlxn17pEs6OWIlJtpCRcKUAg2ehyiiybqCaNYtTAWUO+/Ku0SnovLTp
 sTtvd466SP0GyC8WqqG223ljPwVgPOe/y5ZvRuUY+1CcT4I3iIE/wXcbw9ldZd51
 Tmvx/aZSXpRE8DvYsN4yQpeeJFNVaoTO0IRNf8AG8YQzchRUxdd1l0uy5o2evGXE
 /mZenHRSs/LNfYEwfNhJy6tPGAI9to/O15UHVRS1nneuacMSIyjxYg/kfhmSZKoz
 o9fizcxapx+JwVYHviO6wVdSbgS2aO1u9/whof3Fkm+/Luvo0J4=
 =/Zem
 -----END PGP SIGNATURE-----`
		assert.Equal(t, expectedGPG, tag.GPGSig(), "invalid gpgsig")

		expectedMessage := `tag message`
		assert.Equal(t, expectedMessage, tag.Message(), "invalid Message")
	})
}
