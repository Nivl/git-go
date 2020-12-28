// Package errutil contains methods to simplify working with error
package errutil

import "io"

// Close closes the closer and sets the error to err if err is nil
func Close(c io.Closer, err *error) { //nolint: gocritic // the pointer of pointer is on purpose so we can change the value if it's nil
	e := c.Close()
	if *err == nil && e != nil {
		*err = e
	}
}
