package readutil

// ReadTo reads from b until to is seen and returns the bytes between the start
// and to, exclusive of to. Returns nil if it's not found
func ReadTo(b []byte, to byte) []byte {
	var i int
	for ; i < len(b) && b[i] != to; i++ {
		// the conditions handle it all!
	}

	if i == len(b) {
		return nil
	}

	return b[0:i]
}
