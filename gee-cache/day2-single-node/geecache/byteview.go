package geecache

type ByteView struct {
	b []byte
}

func (v ByteView) Len() int {
	return len(v.b)
}

func (v ByteView) view() {

}
