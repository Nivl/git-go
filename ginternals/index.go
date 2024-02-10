package ginternals

// Index represents a git index file
// An index file contains 4 sections. A header, a list of entries,
// a list of extensions, and a footer.
// Header: 12 bytes
//         The first 4 bytes contain the magic ('D', 'I', 'R', 'C')
//         The next 4 bytes contains the version (0, 0, 0, 2)
//             Valid versions are 2, 3, and 4
//         The last 4 bytes contains the number of entries in the file
// Entries: Variable size
//          Index entries are sorted in ascending order by name. Duplicate
//              names are sorted by their stage number.
//          An entry represents a file, except when sparse-checkout
//              is enabled (both in the config and extension), in which
//              the entry may also represents a directory. Directories have
//              the mode 040000, include the `SKIP_WORKTREE` bit, and the
//              path ends with a directory separator.
//          Data (see stat(2) for more info on some fields):
//              - 4 bytes: the ctime seconds.
//                  ctime: Last time the file's metadata changed
//              - 4 bytes: the ctime nanosecond fractions
//              - 4 bytes: the mtime seconds
//                  mtime: Last time the file's data changed
//              - 4 bytes: mtime nanosecond fractions
//              - 4 bytes: dev (device ID)
//              - 4 bytes: ino (inode's number or file's serial number)
//              - 4 bytes: mode of the entry (high to low, left to right)
//                  - Object type (4 bits)
//                    1000 (regular file), 1010 (symbolic link)
//                    1110 (gitlink)
//                  - unused bits (3 bits)
//                  - UNIX perms (9 bits). Only 0755 and 0644 are valid
//                    for regular files. Symbolic links and gitlinks
//                    have value 0 in this field.
//                  - TODO(melvin): are the last 16bits are unused?
//              - 4 bytes: uid (user ID)
//              - 4 bytes: gid (group ID)
//              - 2 bytes: flags (high to low, left to right)
//                  - assume-valid flag (1 bit)
//                  - extended flag (1 bit). Must be 0 in V2
//                  - stage (2 bits). Used during merge
//                  - name length (12 bits).
//                      - If 0xFFF, the length didn't fit in 12 bits
//              - For version > 3 only
//                  - 2 bytes: extra-data (high to low, left to right). Only
//                      used "extended flag" is 1.
//                      - 1 bit reserved for future
//                      - skip-worktree flag (1 bit). used by sparse checkout
//                      - intent-to-add flag (1 bit). used by "git add -N"
//                      - 13 bits unused. Must be 0.
//              - Entry path name (variable size)
//                  - For version > 4:
//                      - The data starts with a number of variable size
//                          similar to OFS_DELTA.
//                      - The data then contains a variable number of
//                        bytes, representing a string.
//                      - Ends with a NULL byte.
//                      The way this works is that since the entries are
//                        ordered by name, we can reuse part of the previous
//                        entry's name and append to it. The N number
//                        corresponds to the number of character to remove
//                        from the previous entry name. And the string
//                        is what needs to be padded.
//                        Ex. If the previous entry is MyFile1, and the
//                        second entry is MyFile2, then the "N" is 1 (remove
//                        1 char) and the string is "2".
//                  - For version < 4:
//                      1 to 8 NULL bytes as padding
// Extensions: Variable size
//         The first 4 bytes contain the signature. if the firs byte
//             is a chat between A and Z, the extension is optional
//         The next 4 bytes contain the size of the extension
//         The next X bytes contain the extension
// Footer: 20 bytes
//         Contains the SHA1 sum of the packfile (without this SHA)
// https://git-scm.com/docs/index-format
//
// TODO(melvin): Implement Sparse checkout support
// TODO(melvin): Implement split index mode
//    https://git-scm.com/docs/index-format#_split_index
type Index struct {
	version int
}
