// Package packfile contains methods and structs to read and write packfiles
package packfile

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"errors"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/Nivl/git-go/ginternals"
	"github.com/Nivl/git-go/ginternals/object"
	"github.com/spf13/afero"
	"golang.org/x/xerrors"
)

const (
	// packfileHeaderSize contains the size of the header of a packfile.
	// the first 4 bytes contain the magic, the 4 next bytes contains the
	// version, and the last 4 bytes contains the number of objects in
	// the packfile, for a total of 12 bytes
	packfileHeaderSize = 12
)

func packfileMagic() []byte {
	return []byte{'P', 'A', 'C', 'K'}
}

func packfileVersion() []byte {
	return []byte{0, 0, 0, 2}
}

var (
	// ErrIntOverflow is an error thrown when the packfile couldn't
	// be parsed because some data couldn't fit in an int64
	ErrIntOverflow = errors.New("int64 overflow")
	// ErrInvalidMagic is an error thrown when a file doesn't have
	// the expected magic.
	ErrInvalidMagic = errors.New("invalid magic")
	// ErrInvalidVersion is an error thrown when a file has an
	// unsupported version
	ErrInvalidVersion = errors.New("invalid version")
)

// Pack represents a Packfile
// The packfile contains a header, a content, and a footer
// Header: 12 bytes
//         The first 4 bytes contain the magic ('P', 'A', 'C', 'K')
//         The next 4 bytes contains the version (0, 0, 0, 2)
//         The last 4 bytes contains the number of objects in the packfile
// Content: Variable size
//          The content contains all the objects of the packfile, each zlib
//          compressed.
//          Before every zlib compressed objects comes a few bytes of
//          metadata about the object (the type and size of the object).
//          The size of the metadata is variable, so every byte contains
//          a MSB (Most Significant bit, the most left bit of a byte) that
//          indicates if the next byte is also part of the size or not.
//          The very first byte of the metadata contains:
//          - The MSB (1 bit)
//          - The type of the object (3 bits)
//          - the beginning of the size (4 bits)
//          The subsequent bytes contains:
//          - The MSB (1 bit)
//			- The next part of the size (7 bits)
//         The chucks of the size are little-endian encoded (right to left):
//         Final_size = [part_2][part_1][part_0]
//         /!\ The size of the object cannot be used to extract the
//         object. The size corresponds to the real size of the object
//         and not the size of the zlib compressed object (which is)
//         what we have here). It's possible that the compressed object
//         has a bigger size than the de-compressed object.
// Footer: 20 bytes
//         Contains the SHA1 sum of the packfile (without this SHA)
// https://github.com/git/git/blob/master/Documentation/technical/pack-format.txt
type Pack struct {
	r       afero.File
	idxFile afero.File
	idx     *PackIndex
	header  [packfileHeaderSize]byte
	id      ginternals.Oid

	// Mutex used to protect the exported methods from being called
	// concurrently
	mu sync.Mutex
}

// NewFromFile returns a pack object from the given file
// The pack will need to be closed using Close()
func NewFromFile(fs afero.Fs, filePath string) (pack *Pack, err error) {
	f, err := fs.Open(filePath)
	if err != nil {
		return nil, xerrors.Errorf("could not open %s: %w", filePath, err)
	}
	defer func() {
		if err != nil {
			f.Close() //nolint:errcheck // it already failed
		}
	}()

	p := &Pack{
		r:  f,
		id: ginternals.NullOid,
	}

	// Let's validate the header
	_, err = f.ReadAt(p.header[:], 0)
	if err != nil {
		return nil, xerrors.Errorf("could read header of packfile: %w", err)
	}
	if !bytes.Equal(p.header[0:4], packfileMagic()) {
		return nil, xerrors.Errorf("invalid header: %w", ErrInvalidMagic)
	}
	if !bytes.Equal(p.header[4:8], packfileVersion()) {
		return nil, xerrors.Errorf("invalid header: %w", ErrInvalidVersion)
	}

	// Now we load the index file
	indexFilePath := strings.TrimSuffix(filePath, ExtPackfile) + ExtIndex
	p.idxFile, err = os.Open(indexFilePath)
	if err != nil {
		return nil, xerrors.Errorf("could not open %s: %w", indexFilePath, err)
	}
	defer func() {
		if err != nil {
			p.idxFile.Close() //nolint:errcheck // it already failed
		}
	}()
	p.idx, err = NewIndex(bufio.NewReader(p.idxFile))
	if err != nil {
		return nil, xerrors.Errorf("could create index for %s: %w", indexFilePath, err)
	}

	return p, nil
}

// getRawObjectAt return the raw object located at the given offset,
// including its base info if the object is a delta
func (pck *Pack) getRawObjectAt(oid ginternals.Oid, objectOffset uint64) (o *object.Object, deltaBaseSHA ginternals.Oid, deltaBaseOffset uint64, err error) {
	_, err = pck.r.Seek(int64(objectOffset), io.SeekStart)
	if err != nil {
		return nil, ginternals.NullOid, 0, xerrors.Errorf("could not seek from 0 to object offset %d: %w", objectOffset, err)
	}
	buf := bufio.NewReader(pck.r)

	// parse the metadata of the object
	// the metadata is X bytes long and contains:
	// 1 first byte that contains
	//   - a MSB (1 bit)
	//   - the Object type (3 bits)
	//   - the beginning of the object size (4 bits)
	// X more bytes that contains:
	//   - a MSB (a bit)
	//   - the next part of the size (7 bits)
	// Once the MSB of a byte is 0 it means the byte is the last
	// one we need to read.
	// The object size can't really be bigger than 64bits (8 bytes) otherwise
	// we have nothing to store it, so we can just read an enough amount of
	// bytes right away and we should not need to read more.
	// Assuming the worst case scenario (64 bits) we need to read:
	// - 8 bytes (because an int64 is on 8 byte)
	// - 1 extra byte because each of the 8 bytes are missing one bit (because
	//   of the MSB). We lose a total of 8bits, so 1 byte.
	// - 1 extra bit for safety because the first byte we get uses 3 bits
	//   for the type
	// Total: 10 bytes
	metadata, err := buf.Peek(10)
	if err != nil {
		return nil, ginternals.NullOid, 0, xerrors.Errorf("could not get object meta: %w", err)
	}

	// We now need to extract the type of the object. The type is a number
	// between 1 and 7.
	// To extract it (bits 2, 3, and 4) we apply a mask to unset
	// all the bits we don't want, then we move our 3 bits to the
	// right with ">> 4"
	// value       : MTTT_SSSS // M = MSB ; T = type ; S = size
	// & 0111_0000 : 0TTT_0000
	// >> 4        : 0000_0TTT
	objectType := object.Type((metadata[0] & 0b_0111_0000) >> 4)
	if !objectType.IsValid() {
		return nil, ginternals.NullOid, 0, xerrors.Errorf("unknown object type %d", objectType)
	}

	// The first part of the size is on the last 4 bits of the byte.
	// We can use a mask to only keep the bits we want
	// value       : MTTT_SSSS // M = MSB ; T = type; S = size
	// & 0000_1111  : 0000_SSSS
	objectSize := uint64(metadata[0] & 0b_0000_1111)
	metadataSize := 1

	// To know if we need to read more bytes, we need to check the MSB
	// 1 = we read more, 0 = we're done
	if pck.isMSBSet(metadata[0]) {
		size, byteRead, err := pck.readSize(metadata[1:])
		if err != nil {
			return nil, ginternals.NullOid, 0, xerrors.Errorf("couldn't read object size: %w", err)
		}
		metadataSize += byteRead
		// we add 4bits to the right of $size, then we merge everything with |
		// Example:
		// with size = 1001 and objectsize = 1011
		// size << 4  : 1001_0000
		// | size     : 1001_1011
		objectSize |= (size << 4)
	}

	// Since we used Peek() to get the metadata (because we didn't know its
	// size), we now need to discard the right amount of bytes to move
	// our internal cursor to the object data
	if _, err = buf.Discard(metadataSize); err != nil {
		return nil, ginternals.NullOid, 0, xerrors.Errorf("could not skip the metadata: %w", err)
	}

	// Some objects are deltified and need extra parsing before getting to
	// the object content.
	// This is a way for git to only store the changes between 2 similar objects
	// instead of storing 2 full objects. This reduces disk usage.
	// There's 2 types of delta:
	// Refs: This delta contains the SHA of the base object
	// ofs: This Delta contains a negative offset to the base object
	var baseObjectOffset uint64
	var baseObjectOid ginternals.Oid
	switch objectType { //nolint:exhaustive // only 2 types have a special treatment
	case object.ObjectDeltaRef:
		baseObjectSHA := make([]byte, ginternals.OidSize)
		_, err = buf.Read(baseObjectSHA)
		if err != nil {
			return nil, ginternals.NullOid, 0, xerrors.Errorf("could not get base object SHA: %w", err)
		}
		baseObjectOid, err = ginternals.NewOidFromHex(baseObjectSHA)
		if err != nil {
			return nil, ginternals.NullOid, 0, xerrors.Errorf("could not parse base object SHA %#v: %w", baseObjectSHA, err)
		}
	case object.ObjectDeltaOFS:
		// we're assuming the offset is no bigger than 9 bytes to fit an int64.
		// We use 9 instead of 8 because the numbers are on 7bits instead of 8
		// so we need to read an extra byte
		offsetParts, err := buf.Peek(9)
		if err != nil {
			return nil, ginternals.NullOid, 0, xerrors.Errorf("could not get base object offset: %w", err)
		}
		offset, bytesRead, err := pck.readDeltaOffset(offsetParts)
		if err != nil {
			return nil, ginternals.NullOid, 0, xerrors.Errorf("couldn't read base object offset: %w", err)
		}
		baseObjectOffset = objectOffset - offset

		// Since we used Peek() because we didn't know the offset size, we
		// now need to discard the right amount of bytes to move our internal
		// cursor to the object data
		if _, err = buf.Discard(bytesRead); err != nil {
			return nil, ginternals.NullOid, 0, xerrors.Errorf("could not skip the offset: %w", err)
		}
	}

	// We can now fetch the actual data of the object, which is zlib encoded
	zlibR, err := zlib.NewReader(buf)
	if err != nil {
		return nil, ginternals.NullOid, 0, xerrors.Errorf("could not get zlib reader: %w", err)
	}
	defer func() {
		closeErr := zlibR.Close()
		if err == nil {
			err = closeErr
		}
	}()

	objectData := bytes.Buffer{}
	_, err = io.Copy(&objectData, zlibR)
	if err != nil {
		return nil, ginternals.NullOid, 0, xerrors.Errorf("could not decompress: %w", err)
	}

	if objectData.Len() != int(objectSize) {
		return nil, ginternals.NullOid, 0, xerrors.Errorf("object size not valid. expecting %d, got %d", objectSize, objectData.Len())
	}
	return object.NewWithID(oid, objectType, objectData.Bytes()), baseObjectOid, baseObjectOffset, nil
}

// getObjectAt return the object located at the given offset
func (pck *Pack) getObjectAt(oid ginternals.Oid, objectOffset uint64) (*object.Object, error) {
	o, baseOid, baseOffset, err := pck.getRawObjectAt(oid, objectOffset)
	if err != nil {
		return nil, err
	}

	// If the object is not deltified, we don't have anything to do
	if o.Type() != object.ObjectDeltaRef && o.Type() != object.ObjectDeltaOFS {
		return o, nil
	}

	// we retrieve the base object
	var base *object.Object
	if baseOid != ginternals.NullOid {
		base, err = pck.GetObject(baseOid)
		if err != nil {
			return nil, xerrors.Errorf("could not get base object %s: %w", baseOid.String(), err)
		}
	} else {
		// we pass NullOid because we don't know the SHA of the base
		base, err = pck.getObjectAt(ginternals.NullOid, baseOffset)
		if err != nil {
			return nil, xerrors.Errorf("could not get base object at offset %d: %w", baseOffset, err)
		}
	}

	// The format of a delta object is:
	// - A header with:
	//   - The size of the source (x bytes)
	//   - the size of the target (x bytes)
	// - A set of instruction (x bytes)
	delta := o.Bytes()
	sourceSize, sourceSizeLen, err := pck.readSize(delta)
	if err != nil {
		return nil, xerrors.Errorf("couldn't read source size of delta: %w", err)
	}
	if int(sourceSize) != base.Size() {
		return nil, xerrors.Errorf("invalid base object size. expected %d, got %d: %w", base.Size(), sourceSize, err)
	}
	_, tartgetSizeLen, err := pck.readSize(delta[sourceSizeLen:])
	if err != nil {
		return nil, xerrors.Errorf("couldn't read target size of delta: %w", err)
	}
	headerSize := tartgetSizeLen + sourceSizeLen
	instructions := delta[headerSize:]
	baseContent := base.Bytes()

	// We loop over all instructions
	// We don't do a for-range loop because an instruction can be over
	// multiple bytes.
	var out bytes.Buffer

	for i := 0; i < len(instructions); i++ {
		instr := instructions[i]

		// there's 2 types of instruction: COPY and INSERT.
		// If the MSB of the byte is 1 it's a COPY, otherwise it's
		// an INSERT
		switch pck.isMSBSet(instr) {
		case true: // COPY
			// the last 4 bit of the byte contains information about
			// how many bytes to read to get the offset.
			// Example: if the last 4 bits are 1010, we need to read
			// 2 bytes (count the 1), and we'll have to insert to bytes
			// of 0 in the numbers. [first_byte, byte(0), second_byte, byte(0)]
			offsetInfo := uint(instr & 0b_0000_1111)
			var offset uint32
			offsetBytes := make([]byte, 4)
			byteRead := 0
			// our offset will be stored in $offsetBytes
			// We need to loop over the 4 bits of info we have, find the
			// bits that are 1 and insert the correct bytes at the correct
			// index.
			// For example, with 1010 we need to insert our bytes at
			// offsetBytes[0] and offsetBytes[2], and zeros at [1] and [3].
			for j := uint(0); j < 4; j++ {
				offsetBytes[j] = 0

				// we move the current bit to the very left and check that
				// its value is one
				if (offsetInfo >> j & 1) == 1 {
					offsetBytes[j] = instructions[i+1+byteRead]
					byteRead++
				}
			}
			offset = binary.LittleEndian.Uint32(offsetBytes)
			i += byteRead

			// the next 3 bits of the byte after the MSB contains
			// information about how many bytes to read to get the size
			// of the copy (ie. how many bytes we're copying).
			// Example: if the 3 bits are 110, we need to read
			// 2 bytes (count the 1), and we'll have to insert to bytes
			// of 0 in the numbers. [first_byte, byte(0), second_byte, byte(0)]
			copyLenInfo := uint((instr & 0b_0111_0000) >> 4)
			var copyLen uint32
			copyLenBytes := make([]byte, 4)
			byteRead = 0
			// our size will be stored in $copyLenInfo
			// We need to loop over the 3 bits of info we have, find the
			// bits that are 1 and insert the correct bytes at the correct
			// index.
			// For example, with 101 we need to insert our bytes at
			// copyLenInfo[0] and copyLenInfo[2], and a zero at copyLenInfo[1].
			for j := uint(0); j < 3; j++ {
				copyLenBytes[j] = 0

				// we move the current bit to the very left and check that
				// its value is one
				if (copyLenInfo >> j & 1) == 1 {
					copyLenBytes[j] = instructions[i+1+byteRead]
					byteRead++
				}
			}
			// we're working on a 32 bit number (4 bytes) but the size
			// is only stored on 3 bits. We need to make sure the 4th byte
			// is always set to 0
			copyLenBytes[3] = 0
			copyLen = binary.LittleEndian.Uint32(copyLenBytes)
			i += byteRead
			out.Write(baseContent[offset : offset+copyLen])
		case false: // INSERT
			// $instr contains the amount of bytes we need to copy from
			// the delta to the output
			start := i + 1
			end := start + int(instr)
			out.Write(instructions[start:end])
			i += int(instr)
		}
	}

	return object.NewWithID(oid, base.Type(), out.Bytes()), nil
}

// GetObject returns the object that has the given SHA
func (pck *Pack) GetObject(oid ginternals.Oid) (*object.Object, error) {
	pck.mu.Lock()
	defer pck.mu.Unlock()

	objectOffset, err := pck.idx.GetObjectOffset(oid)
	if err != nil {
		if !errors.Is(err, ginternals.ErrObjectNotFound) {
			return nil, xerrors.Errorf("could not get object index: %w", err)
		}
		return nil, err
	}
	return pck.getObjectAt(oid, objectOffset)
}

// ObjectCount returns the number of objects in the packfile
func (pck *Pack) ObjectCount() uint32 {
	return binary.BigEndian.Uint32(pck.header[8:])
}

// ID returns the ID of the packfile
func (pck *Pack) ID() (ginternals.Oid, error) {
	pck.mu.Lock()
	defer pck.mu.Unlock()

	if pck.id != ginternals.NullOid {
		return pck.id, nil
	}

	id := make([]byte, ginternals.OidSize)
	offset, err := pck.r.Seek(-ginternals.OidSize, os.SEEK_END)
	if err != nil {
		return ginternals.NullOid, xerrors.Errorf("could not get the offset of the ID: %w", err)
	}
	if _, err = pck.r.ReadAt(id, offset); err != nil {
		return ginternals.NullOid, xerrors.Errorf("could not read the ID: %w", err)
	}
	pck.id, err = ginternals.NewOidFromHex(id)
	if err != nil {
		return ginternals.NullOid, xerrors.Errorf("could not generate oid from %v: %w", id, err)
	}
	return pck.id, nil
}

// Close frees the resources
func (pck *Pack) Close() error {
	pck.mu.Lock()
	defer pck.mu.Unlock()

	packErr := pck.r.Close()
	idxErr := pck.idxFile.Close()
	if packErr != nil {
		return packErr
	}
	if idxErr != nil {
		return idxErr
	}
	return nil
}

// readSize reads the provided bytes to extract what's left for the
// size from an object metadata.
// This method is only to read the remaining parts of a size.
func (pck *Pack) readSize(data []byte) (objectSize uint64, bytesRead int, err error) {
	for i, b := range data {
		bytesRead++

		// We make sure to remove the MSB because it's not part of the size
		chunk := pck.unsetMSB(b)

		// Sizes are little endian encoded, because why not
		objectSize = pck.insertLittleEndian7(objectSize, chunk, uint8(i))

		// No more MSB? Then we're done reading the size
		if !pck.isMSBSet(b) {
			break
		}
	}

	// if the last byte read has its MSB set it means that we have an
	// overflow (bytesRead - 1 is also == to len(data))
	if pck.isMSBSet(data[bytesRead-1]) {
		return 0, 0, ErrIntOverflow
	}

	return objectSize, bytesRead, nil
}

// readDeltaOffset reads the provided bytes to extract a delta offset.
// The format of the each byte is:
// - 1 bit (MSB) that is used to know if we need to read the next byte
// - 7 bits that contains a chunk of offset
// The offset is big-endian encoded.
// Each chunk of offset (except the last one) are stored -1, so we need
// to add 1 back to each chunk.
func (pck *Pack) readDeltaOffset(data []byte) (offset uint64, bytesRead int, err error) {
	for _, b := range data {
		bytesRead++

		// We set the MSB to 0 since it's not part of the offset
		chunk := pck.unsetMSB(b)

		// To save more space (I guess?), all the chunks beside the last one
		// are stored with -1.
		if pck.isMSBSet(b) {
			chunk++
		}

		// Offsets are big endian encoded, because why not
		offset = pck.insertBigEndian7(offset, chunk)

		// No more MSB? Then we're done reading the offset
		if !pck.isMSBSet(b) {
			break
		}
	}
	// if the last byte read has its MSB set it means that we have an
	// overflow (bytesRead-1 is also == to len(data))
	if pck.isMSBSet(data[bytesRead-1]) {
		return 0, 0, ErrIntOverflow
	}

	return offset, bytesRead, nil
}

// insertLittleEndian7 inserts $chunk into $base from the left.
// Only the 7 most right bits will be inserted.
// Example:
// base   = 1110_1010_1111_1100
// chunk  = 1010_1011
// Result = 1010_1011_1110_1010_1111_1100 [chunk][base]
func (pck *Pack) insertLittleEndian7(base uint64, chunk, position uint8) uint64 {
	// To build the final number in little endian, we first need to
	// add x*7 new bits to the right of the new chunk with "<< position*7"
	// (7, because our chunk is encoded on 7 bits because of the MSB)
	// then we use "| base" to insert and replace all the 0s by the
	// bits we got. x*7 corresponds to the number of bits already set
	// inside $base.
	//
	// That might sound confusing so here's an example:
	// Assuming that:
	// - Our current base is 0000_0000_0011_1010
	// - We're inserting 011_0011 (position=1, because it's the second chunk)
	//
	// 011_0011 << 1*7  = 0001_1001_1000_0000    // we make enough space on the left for $base
	// | base           = 0001_1001_1011_1010 // we insert base
	return (uint64(chunk) << (position * 7)) | base
}

// insertBigEndian7 inserts $chunk into $base from the right
// Only the 7 most right bits will be inserted.
// Example:
// base   = 1110_1010_1111_1100
// chunk  = 1010_1011
// Result = 1110_1010_1111_1100_1010_1011 [base][chunk]
func (pck *Pack) insertBigEndian7(base uint64, chunk uint8) uint64 {
	return base<<7 | uint64(chunk)
}

// isMSBSet checks if the MSB of a byte is set to 1.
// The MSB is the first bit on the left
func (pck *Pack) isMSBSet(b byte) bool {
	return b >= 0b_1000_0000
}

// unsetMSB set the most left bit of the byte to 0
func (pck *Pack) unsetMSB(b byte) byte {
	// To make any bit turn to 0 we can use a mask and a AND operator.
	// Example:
	// value       : XXXX_XXXX
	// & 0111_1111 : 0XXX_XXXX
	return b & 0b_0111_1111
}
