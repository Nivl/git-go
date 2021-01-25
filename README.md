[![codecov](https://codecov.io/gh/Nivl/git-go/branch/main/graph/badge.svg?token=I0YC2EHRHB)](https://codecov.io/gh/Nivl/git-go)

# git-go

Basic git implementation in pure Go

## Current features

### CLI

#### Porcelain

- [x] init

#### Plumbing

- [x] hash-object
- [x] cat-file

### Library

- [x] Retrieve objects
- [x] Write loose objects
- [x] Read/Write References

## Roadmap

### CLI

See the [CLI project](https://github.com/Nivl/git-go/projects/2)

### Library

See the [Library project](https://github.com/Nivl/git-go/projects/1)

## Dev

We use [task](https://github.com/go-task/task) as task runner / build
tool. The main commands are:

- `task test` to run the tests
- `task install` to install the `git-go` to the GOPATH
- `task build` to create a `git` binary in the `./bin` directory
- `task dev -w` to have the binary at `./bin/git` automatically rebuilt with every change in the code
