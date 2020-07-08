# git-go

Basic git implementation in pure Go

## Current features

### CLI

#### Porcelain

- [x] init

#### Plumbing

- [x] hash-object
- [x] cat-file

### API

- [x] Retrieve objects
- [x] Write loose objects
- [x] Read/Write References

## TODO (Short term)

- [x] Add tests
- [x] Run tests on Linux, Windows, and MacOS
- [ ] Add an interface for the Repository so it can be mocked
- [ ] Add support for Short SHA
- [ ] Add support for config file
- [ ] Add support for writing objects in packfile
- [ ] Add support for MIDX
- [x] Add support for trees with AsTree()
- [ ] Add object type to tree entries.
- [ ] Make objects immutable
- [ ] Add Clone/Fetch support with HTTP (Started on branch [`ml/feat/clone`](https://github.com/Nivl/git-go/tree/ml/feat/clone))
