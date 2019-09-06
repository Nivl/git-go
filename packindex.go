package git

import (
	"bytes"
	"encoding/binary"
	"os"
	"sort"

	"github.com/pkg/errors"
)

var (
	// indexHeader represents the header of an index file.
	// the first 4 bytes contain the magic, the 4 next bytes
	// contains the version of the file
	indexHeader     = []byte{255, 't', 'O', 'c', 0, 0, 0, 2}
	layer1Size      = 1024
	layer2EntrySize = OidSize
	layer3EntrySize = 4
	layer4EntrySize = 4
	layer5EntrySize = 8
)

// ErrObjectNotFound is an error corresponding to a git object not being
// found
var ErrObjectNotFound = errors.New("object not found")

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
//         To get the total of object stating with 9b, you will need
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
type PackIndex struct {
	r *os.File
}

// NewPackIndexFromFile returns an index object from the given file
// The index will need to be closed using Close()
func NewPackIndexFromFile(filePath string) (*PackIndex, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, errors.Wrapf(err, "could not open %s", filePath)
	}

	// Let's validate the header
	header := make([]byte, len(indexHeader))
	_, err = f.Read(header)
	if err != nil {
		return nil, errors.Wrap(err, "could read header of index file")
	}
	if !bytes.Equal(header, indexHeader) {
		return nil, errors.Wrap(err, "invalid header")
	}

	return &PackIndex{
		r: f,
	}, nil
}

// Close frees the resources
func (idx *PackIndex) Close() error {
	return idx.r.Close()
}

// GetObjectOffset returns the offset of Oid in the packfile
// If the object is not found ErrObjectNotFound is returned
//
// The way this works is:
// - First we need to check layer1 to get data about how many
//   objects are stored in the packfile.
// - Then we need to look up in layer 2 for the index of the oid we're
//   working on.
// - Then we can look in layer 4 for the object's offset
// - If the packfile is too big (> 2GB), we might need to look in layer5
//   For the actual offset, because layer4 only store int32 offsets
//   while layer5 stores int64 offsets
func (idx *PackIndex) GetObjectOffset(oid Oid) (uint64, error) {
	// First we need to check how many objects in the packfile start by
	// oid[0], how many objects are between 0x00 and oid[0] (cumul), and
	// how many objects are in the packfile
	objCount, cumul, totalObjects, err := idx.ObjectCountAt(oid[0])
	if err != nil {
		return 0, errors.Wrap(err, "could not get object count")
	}
	if objCount == 0 {
		return 0, ErrObjectNotFound
	}

	// Now that we have the cumul, the count and the total, we can
	// find the index of our oid in layer2.
	// Because the data are always ordered in the same way in every layers
	// the index will be use to retrieve the data in the other layer.
	// For example, if our oid is the 3rd in layer2, it will also be the
	// 3rd in layer3 and layer4.
	oidIdx, err := idx.index(oid, objCount, cumul)
	if err != nil {
		return 0, errors.Wrap(err, "could not oid index")
	}

	// Now we can lookup in layer4 for the object's offset in the packfile
	layer2offset := len(indexHeader) + layer1Size
	layer2Size := int64(totalObjects) * int64(layer2EntrySize)
	layer3Size := int64(totalObjects) * int64(layer3EntrySize)
	layer4Offset := int64(layer2offset) + layer2Size + layer3Size
	entryOffset := layer4Offset + int64(oidIdx*layer4EntrySize)

	// Now we can just get the layer4 entry that contains the offset
	entryValue := make([]byte, layer4EntrySize)
	_, err = idx.r.ReadAt(entryValue, entryOffset)
	if err != nil {
		return 0, errors.Wrapf(err, "could not read object offset in layer4 at offset %d", entryOffset)
	}
	entry := binary.BigEndian.Uint32(entryValue)

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
	offset := entry & 0b01111111111111111111111111111111
	// if the msb is not set, then the offset is valid, and we're done
	if !msb {
		return uint64(offset), nil
	}

	// If the msb is set, then the offset we got is to get an entry in layer5,
	// which will contain the offset in the packfile
	layer4Size := int64(totalObjects) + int64(layer4EntrySize)
	layer5Offset := layer4Offset + layer4Size
	entryOffset = layer5Offset + int64(offset)

	entryValue = make([]byte, layer5EntrySize)
	_, err = idx.r.ReadAt(entryValue, entryOffset)
	if err != nil {
		return 0, errors.Wrapf(err, "could not read object offset in layer5 at offset %d", entryOffset)
	}
	finalOffset := binary.BigEndian.Uint64(entryValue)

	return finalOffset, nil
}

// ObjectCountAt searches in layer1 the number of objects starting
// with the given prefix.
// Returns the number of objects at prefix, the cumulative number
// of objects we have at this prefix, and the total amount of object
// the packfile has.
//
// The prefix corresponds of the 2 first chars of a SHA (similar to what
// you can find in .git/objects for danglings objects).
// Because the count is cumulative, to know how many items there is for
// 0x88 (for example), we need to compare it with the count of 0x87
// Ex. For the object fde92e904ca4678cdf23e72582c27a50c310d96d
// the prefix is "0xfd", and the count will be "${count at fd} - ${count at fc}"
func (idx *PackIndex) ObjectCountAt(prefix byte) (count, cumul, total uint32, err error) {
	layer1Offset := len(indexHeader)
	entrySize := 4
	entry := make([]byte, 4)

	// First we get the total of objects in the packfile
	lastEntryOffset := 255 * entrySize
	_, err = idx.r.ReadAt(entry, int64(layer1Offset+lastEntryOffset))
	if err != nil {
		return 0, 0, 0, errors.Wrap(err, "couldn't get the total number of objects")
	}
	total = binary.BigEndian.Uint32(entry)

	prevPrefix := prefix - 1
	// If we're trying to get the count at position 0, then there's
	// nothing before, so we make sure not to have a negative "previous"
	if prefix == 0 {
		prevPrefix = 0
	}

	offset := int64(layer1Offset + int(prevPrefix)*entrySize)
	_, err = idx.r.ReadAt(entry, offset)
	if err != nil {
		return 0, 0, 0, errors.Wrapf(err, "couldn't read previous entry at pos %d", offset)
	}
	prevCumul := binary.BigEndian.Uint32(entry)

	// If we want the count a position 0, then the cumul and the count
	// are the same
	if prefix == 0 {
		count = prevCumul
		cumul = prevCumul
		return count, cumul, total, nil
	}

	// If we want the count for the last position, we already have the total
	// which is also our cumul
	if prefix == 255 {
		count = total - prevCumul
		cumul = total
		return count, cumul, total, nil
	}

	// otherwise we just move to the next entry, and read it
	offset += int64(entrySize)
	_, err = idx.r.ReadAt(entry, offset)
	if err != nil {
		return 0, 0, 0, errors.Wrapf(err, "couldn't read current entry at pos %d", offset)
	}
	cumul = binary.BigEndian.Uint32(entry)
	count = cumul - prevCumul
	return count, cumul, total, nil
}

// index searches for the index of $oid in layer2.
// $oidCount represents the number of oids at oid[0]
// $oidCumul represents the number of oids from 0x00 to oid[0]
func (idx *PackIndex) index(oid Oid, oidCount, oidCumul uint32) (int, error) {
	// layer2offset corresponds to the beginning of layer2 in the file
	layer2offset := len(indexHeader) + layer1Size
	// First index corresponds to the index of the first oid starting by
	// oid[0]
	firstOidIndex := int(oidCumul - oidCount)
	// listOffset correspond to the beginning of the list of the oids
	// starting by oid[0] in Layer2
	listOffset := firstOidIndex * layer2EntrySize

	// Now we grab all the oids that start by oid[0].
	oidRange := make([]byte, oidCount*OidSize)
	rangeOffset := int64(layer2offset + listOffset)
	_, err := idx.r.ReadAt(oidRange, rangeOffset)
	if err != nil {
		return 0, errors.Wrapf(err, "could not read %d oids in layer2 at offset %d", oidCount, rangeOffset)
	}

	// The next step is to do a binary search to find our oid in the list
	// (that's much faster than looping around everything on big packfile)
	oidIdxRel := sort.Search(int(oidCount), func(i int) bool {
		start := i * OidSize
		currentOid := oidRange[start : start+OidSize]
		return bytes.Compare(oid.Bytes(), currentOid) <= 0
	})

	// We need to make sure we found our oid
	start := oidIdxRel * OidSize
	if oidIdxRel >= int(oidCount) || !bytes.Equal(oid.Bytes(), oidRange[start:start+OidSize]) {
		return 0, ErrObjectNotFound
	}

	return firstOidIndex + oidIdxRel, nil
}
