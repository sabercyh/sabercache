package sabercache_server

type ByteView struct {
	bytes []byte
}

func cloneBytes(bytes []byte) []byte {
	copyBytes := make([]byte, len(bytes))
	copy(copyBytes, bytes)
	return copyBytes
}

func (v ByteView) Len() int {
	return len(v.bytes)
}

func (v ByteView) ByteSlice() []byte {
	return cloneBytes(v.bytes)
}

func (v ByteView) String() string {
	return string(v.bytes)
}
