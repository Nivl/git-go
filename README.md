# git-go

Basic git implementation in pure Go

## Current features

### CLI

#### Porcelain

- [x] ini

#### Plumbing

- [x] hash-object
- [x] cat-file

### API

- [x] Read packfile

## TODO (Short term)

- [x] Add tests
- [ ] Run tests on Linux, Windows, and MacOS
- [ ] Add proper support for MIDX
- [ ] Handle Short SHA
- [x] Add support for trees with AsTree()
- [ ] Add object type to tree entries.
- [ ] Add support for writing in packfile/loose objects
- [ ] Add Clone/Fetch support with HTTP (Started on branch [`ml/feat/clone`](https://github.com/Nivl/git-go/tree/ml/feat/clone))
- [ ] Make objects immutable
- [ ] Add an interface for the Repository so it can be mocked
