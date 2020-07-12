# Packfile (V2) Index

## Useful links

- https://shafiul.github.io/gitbook/7_the_packfile.html

## Storage

The packfile index files are stored in `.git/objects/pack` and have
the extension `.idx`.

## Format

The index contains a header, 5 layers, and a footer

### Header

- The header is 8 bytes long and contains two pieces of information:
  - A 4 bytes long magic `255`, `'t'`, `'O'`, `'c'` (or `255`, `116`,
    `79`, `99` in decimal, `0o377`, `0o164`, `0o117`, `0o143` in octal,
    or `0xff`, `0x74`, `0x4f`, `0x63` in hexadecimal)
  - A 4 bytes long version (`int32`). Ex. `0`, `0`, `0`, `2`.

### Layer1 (fanout table)

Layer1 is used to count how many objects have a SHA starting by a
specific value (ex. How many objects have a SHA starting by `e3`)

- It’s 1024 bytes long (256 entries of 4 bytes, so from `00`, to `ff`).
- Entries are ordered. So the entry `e3` is at the position `0xe3 * 4`
- Each entry contains the **cumulative** number of objects having an SHA
  starting by any values between 0 and `SHA[0]`. `SHA[0]` is an hex number
  that corresponds to the first 2 chars of a SHA.
  - For example, the `SHA[0]`
    of `9b91da06e69613397b38e0808e0ba5ee6983251b` is `0x9b` which corresponds
    to `155` in decimal, and the cumulative count will be equal of the amount
    of objects starting by `00` , up to `9b`.
  - Another example: let's say we have 10 objects starting by `00`, 5 by
    `01`, 0 by `02`, and 4 by `03`.
    - `00` will have 10
    - `01` will have 15 (the count at `00` (10) + `01` (5))
    - `02` will have 15 (the count at `00` (10) + `01` (5) + `02` (0))
    - `03` will have 19 (the count at `00` (10) + `01` (5) + `02` (0) + `03` (4))
- The entries are big-endian encoded.

Example:

If we want to know how many objects have a SHA that starts by `9b`, we first
need to find the cumulative count of objects for the entry  `0x9b`, which
is located at position `155 * 4` (or `0x9b * 4`) in layer1. Then we need to
get the **cumulative** count of objects for the previous entry (`0x9a`, at
position `154 * 4`), and do some basic math:
`nb_objects_starting_by_9b = cumul_9b - cumul_9a`

### Layer2 (SHA list)

Layer2 contains the SHAs of all the object contained in
the packfile. We use Layer 2 for two things:

1. Knowing if a object is part of this packfile
1. Finding the index of the object

#### Format

- Each SHA is 20 bytes
- The size of the layer is `X * 20` bytes, where `X` is the number
  of objects in the packfile (the cumul at `0xff` in layer1).
- SHAs are hex encoded (ASCII would be 40 bytes, so it shaves off 50% of space).
  - For the SHA `9b91da06e69613397b38e0808e0ba5ee6983251b`, the bytes will be
    `0x9b`, `0x91`, `0xda`, `0x06`, `0xe6`, ...

#### Tips

- You can find an object quickly by using the data found in layer1
  - If your SHA stars by `9b`, you can know how many object there are before
    by looking for `9a` in Layer1. This allows you to skip a bunch of SHA and
    to only loop over the SHAs staring by `9b`.
- Knowing the index of the object is important. If your SHA is the
  15th SHA in layer2, then it will also be the 15th entry in layer3 and
  the 15th entry in layer4. This allows you to quickly find the data you
  need without having to loop over everything.

### Layer3 (CRC Checksums)

Layer 3 contains a [CRC](https://en.wikipedia.org/wiki/Cyclic_redundancy_check)
value (int32) for each objects. It’s used to check that data are not corrupt
in the packfile.

- The size of layer3 is `X * 4` bytes, where X is the number of objects
  in the packfile.

### Layer4 (Packfile Offsets)

Layer4 contains the offset of each objects inside the packfile (`.pack`).

- Each offset is 4 bytes
- Layer4 total size is `X * 4` bytes, where X is the number of objects
  in the packfile.
- The first **bit** (not byte, 1 byte = 8 bits) of each offset (called
  MSB for Most Significant Bit) is used to store a special value, and
  is not part of the offset.
- If the packfile is smaller than 2GB:
  - The MSB will always be 0
  - The remaining bits (31, because it's 4 bytes of 8 bits minus the MSB,
    so `4*8-1`) correspond to the offset of the object in the packfile.
- If the packfile is bigger than 2GB:
  - The MSB may be 0, or 1
  - If 0, then the next 31 bits will contain the offset of the object
    in the packfile.
  - If 1, then the packfile offset doesn't fit in 4 bytes and has been
    stored in layer5. In that case the next 31 bits will correspond to the
    new location of the offset in layer5.
  - The offsets are big-endian encoded

#### Tips

- You can check the MSB using bitwise operator: `(offset >> 31) == 1`

### Layer5 (Large packfile Offsets)

Layer5 only exists for packfile bigger than 2GB. It's basically the
same as Layer4 but the offset is 8 bytes instead of 4 bytes (4 bytes
was not enough to store the offsets)

- Each offset is 8 bytes
- Layer5 size is `Y * 8`, where `Y` is the number of offsets in Layer4
  that have their MSB set to 1.
- Offsets are big-endian encoded.

### Footer

- The footer is 40 bytes long and contains 2 SHAs
  - The first one is the SHA1 sum of the packfile
  - The second one is the SHA1 sum of the index file (this file)
    without this SHA itself

## Debugging

If you need to debug an index, the best is to use hexdump
with its `-C` and `-s [offset]` flags. This would allow you to print the
data of the index at a specific offset in a format that is easier to read
and to compare it with the data you have.
