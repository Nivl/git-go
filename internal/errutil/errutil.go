// Package errutil contains methods to simplify working with error
package errutil

import "io"

// Close closes the closer and sets the error to err if err is nil
func Close(c io.Closer, err *error) {
	e := c.Close()
	if *err == nil && e != nil {
		*err = e
	}
}
