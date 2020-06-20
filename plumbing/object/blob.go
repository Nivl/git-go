package object

// Blob represents a blob object
type Blob struct {
	*Object
}

// Type returns the ObjectType for this object
func (o *Blob) Type() Type {
	return TypeBlob
}
