package packfile

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"sort"
	"sync"

	"github.com/Nivl/git-go/ginternals"
	"github.com/Nivl/git-go/ginternals/githash"
	"github.com/Nivl/git-go/internal/readutil"
)

const (
	layer1Size      = 1024
	layer3EntrySize = 4
	layer4EntrySize = 4
)

// indexHeader represents the header of an index file.
// the first 4 bytes contain the magic, the 4 next bytes
// contains the version of the file.
// We only support Version 2
func indexHeader() []byte {
	return []byte{255, 't', 'O', 'c', 0, 0, 0, 2}
}

// PackIndex represents a packfile's PackIndex file (.idx)
// The index contains data to help parsing the packfile
// The index contains a header, 5 layers, and a footer.
// header: 8 bytes - See indexHeader to know the header format
// Layer1: 1024 bytes. Contains 256 entries of 4 bytes.
//         Each entry contains the CUMULATIVE number of objects having
//         a oid starting by oid[0].
//         (oid[0] is an hex number, 0 <= x <= 255).
//         It's used to count how many objects have a SHA starting by
//         a specific value.
//         Example:
//         oid[0] represents the value of the 2 first chars of a SHA
//         So for 9b91da06e69613397b38e0808e0ba5ee6983251b, oid[0]
//         is equal to '9b' which corresponds to 155.
//         You'll then find the CUMULATIVE object count at the
//         position 155 * 4 in layer1.
//         To get the total of object starting with 9b, you will need
//         to look at the previous entry (9a at 154 * 4), and do
//         total_at_9b = cumul_9b - cummul_9a
// Layer2: x*20 bytes - Contains the IDs (20 Bytes each) of all the objects
//		   contained in the packfile
// Layer3: x*4 bytes - Contains a CRC (Cyclic redundancy check) value
//         for each object. It's used to check that data did not get corrupt
//         by network operations.
//         https://en.wikipedia.org/wiki/Cyclic_redundancy_check
// Layer4: x*4 - Contains the offset of each objects inside the packfile.
//         The first bit (and not byte, 1 byte = 8 bits) of the offset
//         (called MSB for Most Significant Bit) is used to store a special
//         value, and is not part of the offset:
//
//         If the packfile is < 2GB
//           - The MSB will always be 0
//           - The remaining bit (31, because it's 4 bytes of 8 bits
//             minus the MSB, so 4*8-1) correspond to the offset of
//             the object in the packfile.
//
//         If the packfile is > 2GB
//           - The MSB may be 0, or 1
//           - If 0, then the next 31 bits will contain the offset of
//             the object in the packfile.
//           - If 1, then the packfile offset doesn't fit in 4 bytes and
//             has been stored in layer5. In that case the next 31 bits will
//             corresponds to the new location of the offset in
//             layer5.
// Layer5: y*8 bytes - Only exists for packfile bigger than 2GB.
//         Basically the same as Layer4 but the offsets are on 8 bytes
//         instead of 4, because 4 bytes was too small to store those
//         offsets.
// Footer: 40 bytes - Contains 2 sha of 20 bytes each
//         The first is the sha1 sum of the packfile
//         The second is the sha1 sum of the index file minus this sha
//
// Resources:
// https://codewords.recurse.com/issues/three/unpacking-git-packfiles#idx-files
// https://git-scm.com/docs/pack-format
//
//nolint:govet // aligning the memory makes the struct harder to read since we want to keep "parseError" and "parsed" together
type PackIndex struct {
	mu sync.Mutex

	hash githash.Hash

	r          readutil.BufferedReader
	hashOffset map[githash.Oid]uint64

	parseError error
	parsed     bool
}

// NewIndex returns an index object from the given reader
func NewIndex(r readutil.BufferedReader, hash githash.Hash) (idx *PackIndex, err error) {
	// Let's validate the header
	header := make([]byte, len(indexHeader()))
	_, err = r.Read(header)
	if err != nil {
		return nil, fmt.Errorf("could read header of index file: %w", err)
	}
	if !bytes.Equal(header, indexHeader()) {
		return nil, fmt.Errorf("invalid header: %w", ErrInvalidMagic)
	}

	return &PackIndex{
		r:    r,
		hash: hash,
	}, nil
}

// GetObjectOffset returns the offset of Oid in the packfile
// If the object is not found ginternals.ErrObjectNotFound is returned
func (idx *PackIndex) GetObjectOffset(oid githash.Oid) (uint64, error) {
	if err := idx.parse(); err != nil {
		return 0, fmt.Errorf("could not parse the index file: %w", err)
	}
	offset, exists := idx.hashOffset[oid]
	if !exists {
		return 0, ginternals.ErrObjectNotFound
	}
	return offset, nil
}

// parse extracts all the data from the index and puts them in memory.
func (idx *PackIndex) parse() (err error) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	// No reason to call this method more than once
	if idx.parsed {
		return nil
	}

	// If the method failed, then there's no reason to try again,
	// especially that the underlying reader doesn't get its cursor
	// reset
	if idx.parseError != nil {
		return idx.parseError
	}
	defer func() {
		if err != nil {
			idx.parseError = err
		}
	}()

	bufInt32 := make([]byte, 4)
	bufInt64 := make([]byte, 8)
	bufOid := make([]byte, idx.hash.OidSize())

	// First we parse layer1 to get the count of objects in the packfile.
	// Since layer1 stores a cumul, all we have to do is to get the number
	// at the last position, which is at 0xff (or 255). See doc for
	// more details
	lastEntryRelOffset := 255 * 4 // an entry is an int32, so 4 bytes
	// We move the pointer to the data we need
	_, err = idx.r.Discard(lastEntryRelOffset)
	if err != nil {
		return fmt.Errorf("could not move pointer to the last entry of layer1: %w", err)
	}
	// we now can read the count
	_, err = io.ReadFull(idx.r, bufInt32)
	if err != nil {
		return fmt.Errorf("couldn't get the total number of objects: %w", err)
	}
	objectCount := int(binary.BigEndian.Uint32(bufInt32))

	// Now we can allocate the right amount of memory to store all the
	// oids temporarily in an ordered list, and fill it by parsing
	// layer2 which contains all oids back-to-back
	oids := make([]githash.Oid, 0, objectCount)
	// we basically need to get everything in between layer2 and
	// layer3
	layer2offset := len(indexHeader()) + layer1Size
	layer2Size := objectCount * idx.hash.OidSize()
	layer3offset := layer2offset + layer2Size

	for i := 0; i < objectCount; i++ {
		currentOffset := layer2offset + i*idx.hash.OidSize()
		// this should only happen if the indexfile is invalid and
		// layer2 is smaller than it should
		if currentOffset >= layer3offset {
			return fmt.Errorf("oid %d is out of bound in layer2: %w", i, os.ErrNotExist)
		}

		_, err = io.ReadFull(idx.r, bufOid)
		if err != nil {
			return fmt.Errorf("couldn't get the oid at offset %d: %w", currentOffset, err)
		}
		oid, err := idx.hash.ConvertFromBytes(bufOid)
		if err != nil {
			return fmt.Errorf("invalid oid at offset %d: %w", currentOffset, err)
		}
		oids = append(oids, oid)
	}

	// We don't care about layer3 just yet so we skip it
	// TODO(melvin): parse and use layer3
	// https://golang.org/pkg/hash/crc32/
	layer3Size := objectCount * layer3EntrySize
	_, err = idx.r.Discard(layer3Size)
	if err != nil {
		return fmt.Errorf("could not skip layer3: %w", err)
	}

	// We can now allocate our final map (oid => offset) and fill it with the
	// correct offsets by reading into layer4 and layer5
	// We'll first loop over layer4, then into layer if needed
	idx.hashOffset = make(map[githash.Oid]uint64, objectCount)
	layer4Offset := layer2offset + layer2Size + layer3Size
	layer4Size := objectCount * layer4EntrySize
	layer5Offset := int64(layer4Offset + layer4Size)

	// Before fetching the data in layer 4, we need to make a list to
	// store the object that we'll need to find in layer5. Because we use
	// a buffered reader, we cannot go back and forth between layer4 and 5,
	// so if layer4 contains a layer5 object, we'll have to read it later
	type layer5Data struct {
		oid            githash.Oid
		relativeOffset uint64
	}
	layer5offsets := []*layer5Data{}

	// now we can start parsing layer4
	for i, oid := range oids {
		currentOffset := int64(layer4Offset + i*layer4EntrySize)
		// this should only happen if the indexfile is invalid and
		// layer4 is smaller than it should
		if currentOffset >= layer5Offset {
			return fmt.Errorf("oid %s is out of bound in layer4: %w", oid.String(), os.ErrNotExist)
		}
		_, err = io.ReadFull(idx.r, bufInt32)
		if err != nil {
			return fmt.Errorf("couldn't read offset of oid %s at position %d (layer4): %w", oid.String(), currentOffset, err)
		}
		entry := binary.BigEndian.Uint32(bufInt32)

		// The entry contains 2 information, a MSB and the offset.
		// The MSB correspond to the first bit on the very left, and the
		// offset is stored in the 31 next bits (because its a 32bits number)

		// One way to get the MSB value is to push it 31 bits to the right.
		// If the MSB is one, then our 32bits number will now be
		// 00000000000000000000000000000001, which is the binary
		// representation of 1
		// If the MSB is 0, then all the bits will be set to 0, which is
		// the binary representation of a 0.
		msb := (entry >> 31) == 1

		// Now to get the offset we need to force the MSB to be 0.
		// To do so we can use a binary mask with a AND. We use 0 for the
		// bits we want to change to 0, and 1 for the bits we want to stay at
		// their current value.
		offset := uint64(entry & 0b01111111111111111111111111111111)
		// If the msb is not set, then the offset is valid, and we're done.
		// If the msb is set then the offset we got is to get an entry in
		// layer5, which will contain the offset in the packfile
		if msb {
			layer5offsets = append(layer5offsets, &layer5Data{
				oid:            oid,
				relativeOffset: offset,
			})
			continue
		}
		idx.hashOffset[oid] = offset
	}

	// Now we go get the offset from layer5
	// We need to make sure we access the offset in the right order
	// since we won't be able to go back to a lower offset
	sort.Slice(layer5offsets, func(i, j int) bool { return layer5offsets[i].relativeOffset < layer5offsets[j].relativeOffset })
	currentRelativeOffset := uint64(0)
	for _, data := range layer5offsets {
		// This should never happen since the offsert should be back-
		// to-back, but it cost nothing to double check
		if data.relativeOffset != currentRelativeOffset {
			return fmt.Errorf("expected oid %s to be at (relative) offset %d, but is at %d instead (in layer5 %d): %w", data.oid.String(), currentRelativeOffset, data.relativeOffset, layer5Offset, os.ErrNotExist)
		}

		entryOffset := layer5Offset + int64(data.relativeOffset)
		_, err = io.ReadFull(idx.r, bufInt64)
		if err != nil {
			return fmt.Errorf("couldn't read offset of oid %s at position %d (layer5): %w", data.oid.String(), entryOffset, err)
		}
		offset := binary.BigEndian.Uint64(bufInt64)
		idx.hashOffset[data.oid] = offset
	}
	idx.parsed = true
	return nil
}
