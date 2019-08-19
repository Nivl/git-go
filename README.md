# git

Basic git implementation in pure Go

## Operations supported

- [x] Init (`agit init`, `InitRepository()`)
- [x] Dangling objects (`agit hash-object`, `agit cat-file`)

## TODO (Short term)

- [x] Write packfile reader
  - [x] Use a bufio for the entire packfile (not index)
  - [x] Deltified object
- [x] Reduce calls to ReadAt when possible (for example we can batch the header of the packfile in one read of 12 bytes instead of 3 read of 4 bytes).
- [ ] Add linter
- [ ] Handle Short SHA
- [ ] Add a command to test AsCommit()
- [ ] Add support for trees with AsTree()
- [ ] Add test for everything Object related
- [ ] Add support for writing in packfile/dangling objects
