package object_test

import (
	"bytes"
	"testing"

	"github.com/Nivl/git-go/plumbing"
	"github.com/Nivl/git-go/plumbing/object"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAsCommit(t *testing.T) {
	t.Parallel()
	t.Run("regular commit with all the fields", func(t *testing.T) {
		t.Parallel()

		treeID, _ := plumbing.NewOidFromStr("f0b577644139c6e04216d82f1dd4a5a63addeeca")
		oid, _ := plumbing.NewOidFromStr("0343d67ca3d80a531d0d163f0078a81c95c9085a")
		parentID, _ := plumbing.NewOidFromStr("9785af758bcc96cd7237ba65eb2c9dd1ecaa3321")

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

		o := object.NewWithID(oid, object.TypeCommit, rawData)
		expectedSigName := "Melvin Laplanche"
		expectedSigEmail := "melvin.wont.reply@gmail.com"
		expectedSigTimestamp := int64(1566115917)
		expectedSigOffset := 3600 * -7

		ci, err := o.AsCommit()
		require.NoError(t, err)

		assert.Equal(t, o.ID, ci.ID)
		assert.Equal(t, "f0b577644139c6e04216d82f1dd4a5a63addeeca", ci.TreeID.String(), "invalid tree id")

		require.NotNil(t, ci.Author, "author missing")
		assert.Equal(t, expectedSigName, ci.Author.Name, "invalid author name")
		assert.Equal(t, expectedSigEmail, ci.Author.Email, "invalid author email")
		assert.Equal(t, expectedSigTimestamp, ci.Author.Time.Unix(), "invalid author timestamp")
		_, tzOffset := ci.Committer.Time.Zone()
		assert.Equal(t, expectedSigOffset, tzOffset, "invalid author timezone offset")

		require.NotNil(t, ci.Committer, "committer missing")
		assert.Equal(t, expectedSigName, ci.Committer.Name, "invalid committer name")
		assert.Equal(t, expectedSigEmail, ci.Committer.Email, "invalid committer email")
		assert.Equal(t, expectedSigTimestamp, ci.Committer.Time.Unix(), "invalid committer timestamp")
		_, tzOffset = ci.Committer.Time.Zone()
		assert.Equal(t, expectedSigOffset, tzOffset, "invalid committer timezone offset")

		require.Len(t, ci.ParentIDs, 1, "invalid amount of parent")
		assert.Equal(t, "9785af758bcc96cd7237ba65eb2c9dd1ecaa3321", ci.ParentIDs[0].String(), "invalid parent id")
	})
}
