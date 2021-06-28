# Paths

Git allows a lot of flexibility and control over where the internal files
are being written to. This document isn't exhaustive and only covers
what's implemented in this codebase.

The most common setting is to have a `.git` directory at
the root of the working tree that contains all git's files.

## Working Directory

This directory correspond to the current directory the program is called from.

- In the CLI, this path can be overwritten using the `-C` flag.
- In our Library, this path can be ovewriten using `LoadConfigOptions.WorkingDirectory`

## Work tree

The work tree corresponds to the directory that contains your source
code. By default, this path is found by using the Working Directory and
recursively moving up to the parent directory until the Git directory
is found.

- In the CLI this path can be overwritten using (list is in parsing order, every value overwrites the previous one):

  - The `core.Worktree` key of git's config file
  - The `GIT_WORK_TREE` env variable
  - The `work-tree` flag

- In our Library, this path can be ovewriten using (list is in parsing order, every value overwrites the previous one):
  - The `core.Worktree` key of git's config file
  - The `GIT_WORK_TREE` env variable
  - `LoadConfigOptions.WorkTreePath`

## Git directory

This git directory is the place that contains all, or most git related
data. By default, you can find it inside the worktree under the `.git` name.

- In the CLI this path can be overwritten using (list is in parsing order, every value overwrites the previous one):

  - The `GIT_DIR` env variable
  - The `git-dir` flag

- In our Library, this path can be ovewriten using (list is in parsing order, every value overwrites the previous one):
  - The `GIT_DIR` env variable
  - `LoadConfigOptions.GitDirPath`

## Git Common directory

This directory contains all the non-worktree-related files that are normally
in the Git directory. By default, this directory is the same as the Git Directory.

- In the CLI or our library, this path can be overwritten using (list is in parsing order, every value overwrites the previous one):

  - The `GIT_COMMON_DIR` env variable
  - Setting the path in the `$GIT_DIR/commondir` file

### What goes in this directory

- if .git/commondir exists, $GIT_COMMON_DIR will be set to the path specified
  in this file if it is not explicitly set. If the specified path is relative,
  it is relative to $GIT_DIR. The repository with commondir is incomplete without
  the repository pointed by "commondir"
- List of directories under GIT_COMMON_DIR
  - $GIT_COMMON_DIR/objects
  - $GIT_COMMON_DIR/refs
  - $GIT_COMMON_DIR/packed-refs
  - $GIT_COMMON_DIR/config
  - $GIT_COMMON_DIR/branches
  - $GIT_COMMON_DIR/hooks
  - $GIT_COMMON_DIR/info
  - $GIT_COMMON_DIR/remotes
  - $GIT_COMMON_DIR/logs
  - $GIT_COMMON_DIR/shallow
  - $GIT_COMMON_DIR/worktrees

## Git Object directory

This directory contains all the loose objects and the packfile of
the repository. By default it's found under `$GIT_COMMON_DIR/objects`.

- In the CLI or our library, this path can be overwritten using (list is in parsing order, every value overwrites the previous one):

  - The `GIT_OBJECT_DIRECTORY` env variable

## Configuration files

Git contains several layers of configuration files, that are all merged
together:

### System

The system config file is the highest level one, and is overwritten by
all the others. This is the configuration file used for all the user on
the machine. There are multiple paths where it can be found. By
default git reads the file located at `$(prefix)/etc/gitconfig`. `$(prefix)`
is a value set a compile time (during the compilation of git itself).

- In _our_ CLI and library this path can be set using (list is in parsing order, every value overwrites the previous one):

  - The `PREFIX` env variable

If no `$PREFIX`, we will for:

- `/etc/gitconfig` (unix)
- `/usr/local/etc/gitconfig` (unix)
- `/opt/homebrew/etc/gitconfig` (macos)
- `%ALLUSERSPROFILE%\Application Data\Git\config` (windows)
- `%ProgramFiles(x86)%\Git\etc\gitconfig` (windows)
- `%ProgramFiles%\Git\mingw64\etc\gitconfig` (windows)

### Global

The global config file is set at a user level, usually used to be able to have a
common configuration across repositories.

The common paths to store the global file are:

- `$HOME/.gitconfig` (all systems)
- `%USERPROFILE%\.gitconfig` (windows)
- `$XDG_CONFIG_HOME/git/.gitconfig` (unix)
- `$HOME/config/.git/.gitconfig` (unix)

### Local

The local config file is the most specific one and is bound to a specific repository. Its default path is `$GIT_COMMON_DIR/config`, but can be overwritten using the `$GIT_CONFIG` env variable.
