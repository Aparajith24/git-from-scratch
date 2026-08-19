package objects

// Blob is git's simplest object: raw file content, no metadata beyond the
// "blob <size>\0" header that WriteObject/ReadObject already handle.
type Blob struct {
	Data []byte
}

func NewBlob(data []byte) Blob { //create new blob
	return Blob{Data: data}
}

func (b Blob) Type() string { //return type of object as blob if given
	return "blob"
}

// Save writes the blob to the object store and returns its SHA-1.
func (b Blob) Save(gitDir string) (string, error) { //save blob
	return WriteObject(gitDir, "blob", b.Data)
}
