# Commit objects

## Useful links

- The Commit Objects: https://git-scm.com/book/en/v2/Git-Internals-Git-Objects#_git_commit_objects
- `git commit` documentation: https://git-scm.com/docs/git-commit
- Anatomy of a commit: https://wyag.thb.lt/#orgea3fb85

## Format

### Commit

A commit contains a list of key/value in ASCII that ends with a blank line,
followed by the commit messages. Here’s an example of a commit:

```
tree e5b9e846e1b468bc9597ff95d71dfacda8bd54e3
parent 6097a04b7a327c4be68f222ca66e61b8e1abe5c1
author Melvin Laplanche <melvin.wont.reply@gmail.com> 1592615777 -0700
committer Melvin <Nivl@users.noreply.github.com> 1592616250 -0700
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
----END PGP SIGNATURE-----

doc: Update TODOs in readme
```

- First we have a list of key/value in ASCII:
  - `tree` contains the SHA of the commit's tree
  - `parent` is optional and can appear multiple times. They contains
    the SHA of the parent commit. See the [`git merge` documentation](https://git-scm.com/docs/git-merge#_fast_forward_merge) for more details
    - If the commit has no parent it means the commit is the very first one
      in it's history (for example the first commit of the repo or the
      first commit of an orphan branch)
    - Commits with one parent are usually regular commits or commits
      from coming from a [fast-forward merge](https://git-scm.com/docs/git-merge#_fast_forward_merge)
    - Commits with 2 or more parents are the result of a [true merge](https://git-scm.com/docs/git-merge#_true_merge)
      (no fast-forward)
  - `author` contains a signature (see below for format) representing
    the person that made the changes being committed
  - `committer` contains a signature (see below for format) representing
    the person who created the commit. The difference with Author is around
    permissions. Some systems (for example), don't allow allow contributors
    to send a commit or a Pull Request directly to the codebase. Contributors
    need to send a patch using a mailing list, and a maintainer would review
    the diff and create a commit off it. The maintainer would be the
    `committer`, and the person that sent the patch would be the `author`. This
    lets other maintainers know who did the work and who added it to the system.
  - `pgpsig` is optional and is only there if the commit has been signed
    with a PGP key. The content is on multiple line and contains the PGP
    signature of the key used. This is used as a security measure to make
    sure the person that created the commit really is who they say they are.
- After all of this comes an empty line
- The commit's data ends with the commit message, which can be over multiple
  lines.

### Signature

A signature has a basic format:

- The name of the person
- Followed by a space char character (`' '`, `0x20`, `32`, `040`)
- Followed by a `<` (`'<'`, `0x3c`, `60`, `074`)
- Followed by the email address of the person
- Followed by a `>` (`'>'`, `0x3e`, `62`, `076`)
- Another space character (`' '`, `0x20`, `32`, `040`)
- a UNIX timestamp (local to the user timezone) of when the commit has been created
- Another space character (`' '`, `0x20`, `32`, `040`)
- Finally, the timezone with sign of where the commit got created (ex `-0700`, `+0000`, `+0100`, etc.)
