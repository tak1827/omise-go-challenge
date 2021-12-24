package cipher

type Reader interface {
	Read(b []byte) (n int, err error)
	ReadAt(b []byte, off int64) (n int, err error)
}
