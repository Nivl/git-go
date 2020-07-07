> ⚠️ This document is just a quick introduction/glossary that tries to explain things the easy way and might not be 100% accurate for the sake of staying easily readable and not too verbose.

If you ever wonder what a git command does internally, you can add `GIT_TRACE=1` in front of it.

Ex.:

```
❯ GIT_TRACE=1 git branch
19:42:49.324029 git.c:439               trace: built-in: git branch
19:42:49.328964 run-command.c:663       trace: run_command: unset GIT_PAGER_IN_USE; LV=-c 'diff-so-fancy | less --tabs=4 -RFX'
19:42:49.858665 git.c:439               trace: built-in: git config --list
* master
```

## Useful links

- https://git-scm.com/book/en/v2/

## Objects

### Resources

- https://git-scm.com/book/en/v2/Git-Internals-Git-Objects
- https://stackoverflow.com/questions/22968856/what-is-the-file-format-of-a-git-commit-object
- https://wyag.thb.lt/#objects

### Key Points

- In Git almost everything is an Object:
  - commits
  - tags (but only the tags created with the -a flag)
  - trees which are directories (ex. git add src/main.go will create a tree object for src)
  - blobs which are basically all the files you add with git add (so git add main.go will create a blob object containing main.go content).
- All objects are represented by a SHA (also referred to as OID for Object ID). This SHA corresponds to the SHA1 hash of the content of an object file before it has been zlib compressed (so `sha1(type + ' ' + data_size + 0x00 + data)`).
  -Objects are stored on disk and are compressed with zlib. They follow a specific format.
- They can be found at 2 different places:
  - In .git/objects
    - Some objects created less than 2 weeks ago as well as blobs not added to any commits are stored in there.
    - You can find an object using its SHA. The first directory is the first 2 chars of the SHA, then the file is the remaining 38 chars: .git/objects/sha[0:2]/sha[2:].
    - Ex: The commit with SHA `63a972a73a396a758178ca604e5d8acce693bcca` can be found at `.git/objects/63/a972a73a396a758178ca604e5d8acce693bcca`
  - In a packfile located at `.git/objects/packs`
    - See Packfile section below
- Objects that are not part of a packfile are known as "loose objects".
  - You can move loose objects to a packfile by running Git's garbage collector with `git gc`
- Some objects known as "dangling objects" are objects not attached to any references. They are unreachable, except if you already know their OID.

### Investigating/debugging Objects

#### Find the SHA/OID a file has (or would have) once added to git

```
❯ git hash-object main.go
03f6454b22ad871240b2505c0fb24d290d279d15
```

#### Get the type of an object from its SHA

```
❯ git cat-file -t  03f6454b22ad871240b2505c0fb24d290d279d15
blob
```

#### Get the content of an object from its SHA

```
❯ git cat-file -p  03f6454b22ad871240b2505c0fb24d290d279d15
[content of file main.go]
```

#### Other useful commands to look at an object

```
❯ git show sha # works with everything
❯ git ls-tree sha # for trees only
❯ git log sha # for commits only
❯ git fcsk # To verify and validate all objects
```

## Packfile

### Resources

- https://git-scm.com/book/en/v2/Git-Internals-Packfiles
- https://codewords.recurse.com/issues/three/unpacking-git-packfiles

### Key points

- A packfile is a single file containing the contents of all the objects. It’s basically an optimized local database.

- To prevent using too much space, a packfile will stores deltas instead of full objects when possible. Example:
  - You git add main.go and commit it,
  - You add a trailing new line at the end of main.go, git add the changes then commit.
  - You now have 2 blobs objects that are similar at 99%, and this one byte you added to main.go is costing you a lot of disk space.
  - Once put in the packfile, the second object will only contain deltas instead of the full content, reducing its size to only a few bytes.
- Packfiles come in pair with an index file (.idx). The index contains offsets into that packfile so you can quickly seek to a specific object (more about the format of the index).
- Git will automatically move objects inside a packfile from times to times (when you push or pull for example).
- You can manually move all packable dangling objects inside a packfile by running git gc. It’s totally safe to manually run, no side effects should happen.

### Investigating/Debugging Packfiles

#### List all the objects inside a Packfile

```
❯ git verify-pack -v .git/objects/pack/pack-7a16e4488ae40c7d2bc56ea2bd43e25212a66c45.idx
0155eb4229851634a0f03eb265b69f5a2d56f341 tree   71 76 5400
05408d195263d853f09dca71d55116663690c27c blob   12908 3478 874
09f01cea547666f58d6a8d809583841a7c6f0130 tree   106 107 5086
1a410efbd13591db07496601ebc7a059dd55cfe9 commit 225 151 322
...
```

The 1st column contains the SHA of the object, the 2nd the type, the 3rd the size of the object, the 4th contains the size of the object once zlib compressed, and the last column contains the object location in the packfile (offset in byte).

#### Unpack all objects of a Packfile

```
# From within the repository
❯ mv .git/objects/pack/pack-HASH.pack . # Move the packfile away
❯ git unpack-objects < pack-HASH.pack   # Unpack the packfile in the current repo
```

It’s required to move the packfile out of the .git directory, because git won’t let you create objects that already exist in a packfile.

#### Other useful commands

```
❯ git gc # pack dangling objects and optimize the packfiles
❯ git repack # optimize a packfile by repacking it
```

## References

### resource

- https://git-scm.com/book/en/v2/Git-Internals-Git-References

### Key points

References are basically labels. They are a way to link a user-friendly name to a SHA, preventing you from having to know and remembering SHAs.

There are 2 types of references:

- Symbolic references: they are references pointing to another reference
- OID references: they are reference pointing to an object

References can be found at 2 differences places:

- In the `.git/refs` directory, where each reference will be in a file
- In the `.git/packed-refs` file, where each references will be on a line.
  A reference can appear both in `.git/refs` AND `.git/packed-refs` with a different target. In this case, `.git/refs` will contain the most up-to-date data.

### Branches

Branches are references located in `.git/refs/heads`. They point to a single SHA and are automatically updated when creating new branches, commits, etc.

### HEAD

HEAD is basically a reference to the last commit of the current branch (but not always). You can go higher in the history by adding `~` followed by a number (ex: `HEAD~1` correspond to the commit right after the last one), or by adding a bunch of `^`, (ex: use `HEAD^^` to get the third commit (= `HEAD~2`)).

The HEAD is located in `.git/HEAD`, and contains either a `SHA` (detached head, if you check out a commit for example), or a symbolic reference to a branch, if you’re in a branch.

#### Examples

Given the following history

```
* 5457f77c15 fix(files): Do not multi-select deleted layers (#1316)
* 36793a8812 fix: comment form sizing (#1377)
* fa5b5732c2 fix: Specs after merge
* 7d2d198eaa fix: lastPulledAt should be set to the pushedAt timestamp (#1149)
* d42fda8c0b fix: copy links not routing to correct scroll within comment feed (#1363)
```

We can use rev-parse to get the referenced commit SHA:

```
❯ git rev-parse HEAD
5457f77c153d0f17042ee425f4985566fb21c02c
❯ git rev-parse HEAD-1
36793a88124435fd7bc328ddb7799572dc560646
❯ git rev-parse HEAD~2
fa5b5732c25f7365795d1bf06fefdf529c83f7c6
❯ git rev-parse HEAD^^^ # In zsh you have to escape ^: git rev-parse HEAD\^\^\^
7d2d198eaae90213e4aff3673b59289d1f681787
```

### Working tree

The working tree is basically your file system. When you open a file in your code editor and start changing things, you’re editing that file on the working tree.

#### Investigating/Debugging the Working Tree

- `git status` to see the files/directory that changed (sections `Changes not staged for commit` and `Untracked files`, changes are appearing in red if you have colors enabled)
- `git diff` to see the specific changes in each files
- `git checkout -- <file>` can be used to revert the changes of a file

### Index

#### Key Points

- This is not the same as a packfile’s index.
- The index is also known as the staging area
- This is where all your changes go when using git add
- When creating a commit, git is committing what’s in the index
- The index is a binary file located at .git/index

#### Investigating/debugging the Index

- `git status` to see what changed in the index (section `Changes to be committed`, the changes will appear in green if you have colors enabled).
- `git reset HEAD <file>` can be used to remove files from the index without impacting the working tree.
