package env

import (
	"os"
	"strings"
)

// Env represents the environment
type Env struct {
	env map[string]string
}

// NewFromOs builds and returns an Env using os.Environ
func NewFromOs() *Env {
	return NewFromKVList(os.Environ())
}

// NewFromKVList builds and returns an Env using a provided list of
// string in the form "key=value"
func NewFromKVList(env []string) *Env {
	e := &Env{
		make(map[string]string, len(env)),
	}
	for _, kv := range env {
		data := strings.Split(kv, "=")
		e.env[data[0]] = data[1]
	}
	return e
}

// Has returns whether the given key has a value set.
// Has is case-sensitive.
func (e *Env) Has(key string) bool {
	_, ok := e.env[key]
	return ok
}

// Get returns the value of the given key, or en empty string if the key
// has no values set.
// Get is case-sensitive.
func (e *Env) Get(key string) string {
	v, ok := e.env[key]
	if !ok {
		return ""
	}
	return v
}
