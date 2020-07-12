# Tree objects

## Useful links

- https://git-scm.com/book/en/v2/Git-Internals-Git-Objects
- https://github.com/git/git/blob/master/Documentation/technical/index-format.txt
- Parsing trees: https://wyag.thb.lt/#orga708f5f

## Format

- A tree contains a list of entries back-to-back
- Each entry is composed of:
  - An octal mode, in ASCII, representing the permission of the entry. The
    mode is similar to the UNIX permissions but is not the same:
    - For blobs, the modes supported by git are:
      - `100644`: For regular files
      - `100755`: for executable files
      - `100664`: [**non-standard**](https://github.com/git/git/blob/bd42bbe1a46c0fe486fc33e82969275e27e4dc19/fsck.c#L725)
        this used to be used for regular file with group writing permission
    - For trees:
      - `040000`: For Directories
    - Others:
      - `120000`: For symlinks
      - `160000` For gitlink (submodules)
  - A space character (`' '`, `0x20`, `32`, `040`)
  - The path of the entry, in ASCII
  - A NULL character (`'\0'`, or just `0`)
  - An hex-encoded SHA corresponding to the SHA of the targeted object
