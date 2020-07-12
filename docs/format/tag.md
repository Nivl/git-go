# Tag objects

## Format

A tag is very similar to a commit:

```
object d0197f865ee274c57efbd4027e5632a95c8f5897
type commit
tag cli@1.12.1
tagger Melvin Laplanche <melvin.wont.reply@gmail.com> 1562092118 -0700

cli@1.12.1
```

- First we have a list of key/value in ASCII:
  - `object` contains the SHA of the targeted object
  - `type` contains the type of the target object
  - `tag` contains the annotation of the tag
  - `tagger` contains a Signature (see Commit documentation) of the
    person that created the tag
- After all of this comes an empty line
- The tag's data ends with the tag's message
