// Package errutil contains methods to simplify working with error
package errutil

import (
	"io"
	"log"
)

// Close closes the closer and sets the error to err if err is nil
func Close(c io.Closer, err *error) { //nolint: gocritic // the pointer of pointer is on purpose so we can change the value if it's nil
	e := c.Close()
	switch *err { //nolint: errorlint,gocritic // no need to use errors.Is here since we're only checking for "is nil or not"
	case nil:
		*err = e
	default:
		if e != nil {
			log.Println("Close() failed:", e)
		}
	}
}
