package vol

import "io"

type ByteBuffer []byte

func (b *ByteBuffer) Skip(n int) bool {
	_, ok := b.Next(n)
	return ok
}

func (b *ByteBuffer) MustNext(n int) (sub ByteBuffer) {
	sub, ok := b.Next(n)
	if !ok {
		panic(io.ErrUnexpectedEOF)
	}
	return sub
}

func (b *ByteBuffer) Next(n int) (sub ByteBuffer, ok bool) {
	if len(*b) < n {
		return nil, false
	}
	sub, *b = (*b)[:n], (*b)[n:]
	return sub, true
}

// Append adds bs to the end of b.
func (b *ByteBuffer) Append(bs ...byte) {
	*b = append(*b, bs...)
}

// AppendString adds the bytes of s to the end of b.
func (b *ByteBuffer) AppendString(s string) {
	*b = append(*b, s...)
}

// Extend adds n uninitialized bytes to b and returns that range as a slice.
func (b *ByteBuffer) Extend(n int) ByteBuffer {
	cur := len(*b)
	if cur+n <= cap(*b) {
		*b = (*b)[:cur+n]
	} else {
		b2 := make([]byte, cur+n, cur+2*n)
		copy(b2, *b)
		*b = b2
	}
	return (*b)[cur : cur+n]
}
