# Packfile (V2)

## Useful links

- https://git-scm.com/docs/pack-format
- https://shafiul.github.io/gitbook/7_the_packfile.html
- https://codewords.recurse.com/issues/three/unpacking-git-packfiles
- https://medium.com/@concertdaw/sneaky-git-number-encoding-ddcc5db5329f

## Storage

The packfile files are stored in `.git/objects/pack` and have
the extension `.pack`.

## Format

The packfile contains a header, a "body", and a footer.

### Header

- The header is 12 bytes long and contains 3 pieces of information:
  - A 4 bytes long magic: `'P'`, `'A'`, `'C'`, `'K'`.
  - A 4 bytes long version (`int32`). Ex. `0`, `0`, `0`, `2`.
  - A 4 bytes long number (`int32`) containing the number of objects
    in the packfile.

### Body

The body contains all the objects back to back. You need to know the offset
of an object to be able to find it in a perf friendly manner (except if
you'd rather loop over every objects one-by-one).

- Each object starts with a few bytes of metadata:
  - 1 byte that contains:
    - 1 bit (MSB) that is used to know if we need to read the next byte or not
    - 3 bit that contains the type of the object
    - 4 bits that contains a chunk of the object size
  - X more bytes that each contains:
    - 1 bit (MSB) that is used to know if we need to read the next byte or not
    - 7 bits that contain a chunk of the object size (to concat with the
      other chunks)
- The final size of an object is composed of all the chunks read,
  and is little-endian encoded
- ⚠️ The size of the object stored in the the metadata is the **actual size**
  of the object, and not its zlib-compressed size. This means the size cannot
  be used to allocate a buffer to store the object content since the
  object content is zlib compressed (and yes, the compression can sometimes
  use more space than the actual object).
- The data of an object is stored right after its metadata and is zlib
  compressed
  - To parse the content of a common object, See the Commit, Tree, Blob,
    and Tag format documentation.
  - There are 2 special objects that need extra parsing: The Delta. This is
    a way for git to reduce the size of the packfile. Instead of storing 2
    similar blobs, git will store only one as blob, and the other one as
    delta. The delta’s data will be a list of instructions used to rebuild
    the blob based on the other object. There are 2 kinds of delta:
    - **delta refs**: Those deltas’ content starts with a 20 bytes long
      oid corresponding to the base object that needs to be used to rebuild
      our object.
    - **delta offset**: Those deltas contain a relative offset to the base
      object, stored in an unknown amount of bytes (similar to the object size,
      but with a few changes):
      - Each byte contains 1 bit (MSB) to know if we need to read the next
        byte or not
      - Each byte contains 7 bits that represents a chunk of the offset
      - The offset is big-endian encoded (and not little-endian encoded
        like the size)
      - In order to save more space (I think?), each chunk is missing 1,
        except for the last chunk. This means that once you unset the MSB
        of the byte, you need to do `chunk = chunk + 1`.
    - After reading those data, both deltas will contain the same type of
      content.
      - They start with the size of the source object (stored the same way as
        the size of the object)
      - Then comes the size of the target object.
      - After this, you get a series of instructions used to create a new
        object. The first byte you will read will contain a MSB, if this MSB
        is set to 1 the instruction is a COPY, otherwise it’s an INSERT.
        - **COPY** (this one is tricky):
          - The last 4 bits of the byte you read contain the information
            about how many bytes you now need to read to get the offset that
            corresponds to where the copy starts in the base object. You only
            care about the amount of 1s, so if you have `1010`, that’s 2 bytes
            to read, `1110`, that’s 3, `1001` is also 2, etc.
          - Once you read those bytes, you might need to insert a few bytes
            of 0 in between them. If the 4 bits of information you had was
            `1010`: it means you need to rebuild the offset by inserting the
            first byte you read, then inserting an `int8(0)`, inserting the 2nd
            byte you read, and inserting an `int8(0)`. For `1010`, the final
            number will be `[first_byte]00000000[second_byte]00000000`. This
            is a way for git to save a bunch of space by not having to store
            empty bytes.
          - The offset is little-endian encoded.
          - Once you rebuilt this offset, you need to do the exact same
            thing to get the amount of byte that needs to be copied. The
            information for this size is stored on 3bits right after the
            MSB (basically in between the MSB and the information of the
            offset, in the instruction byte).
          - Once you have the offset and the copy size, you can just copy
            the data from the base object to the new object.
          - The next byte will contain a new instruction.
        - **INSERT**:
          The byte you just read contains the amount of bytes that need to be copied from the delta’s content to the new object. Ex. if the byte contains 9, you need to copy the next 9 bytes to the new object, and the next byte after those will contain the next instruction.

### Footer

The footer is 20 bytes long and contains the SHA1 sum the file
(without the SHA itself, for obvious reasons)

## Debugging

If you need to debug a packfile, the best is to use hexdump
with its `-C` and `-s [offset]` flags. This would allow you to print the
data of the packfile at a specific offset in a format that is easier to read
and to compare it with the data you have.
