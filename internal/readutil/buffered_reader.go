package readutil

// BufferedReader is an interface that represents a reader with an
// internal buffer.
type BufferedReader interface {
	Discard(n int) (discarded int, err error)
	Read(p []byte) (n int, err error)
}
