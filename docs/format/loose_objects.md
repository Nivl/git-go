# Loose objects

## Storage

The loose objects can be found in the `.git/objects` directory. They are
stored in a sub directory that is named after the 2 first characters of
the object's SHA, and in a file named after the other characters from the
SHA. For example, a loose object with the SHA `f0b577644139c6e04216d82f1dd4a5a63addeeca`
will be stored at `.git/objects/f0/b577644139c6e04216d82f1dd4a5a63addeeca`

## Format

- The content of each file is compressed with zlib
- The data start with the type of the object in ASCII (`commit`, `tree`, `blob`, or `tag`)
- Then a a space character (`' '`, `0x20`, `32`, `040`)
- The next few bytes correspond to the size of the object in ASCII
- Then comes a NULL character (`'\0'`, or just `0`)
- Everything else will be the object's actual data (see other documents
  on how to parse each object).
