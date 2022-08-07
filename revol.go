package vol

import (
	"encoding/binary"
)

func (v File) Store(buf *ByteBuffer) {
	hdrPayload := headerAndPayload{IsPVOL: true}
	fnFooter := filenameFooter{}
	itFooter := itemFooter{}

	const hdrLen = blockHeaderLen // length of header before payload
	for _, item := range v.Items {
		itemOff := hdrLen + len(hdrPayload.Payload)
		payloadItem{Payload: item.Payload}.Store(&hdrPayload.Payload)
		// TODO: add special case to convert 0-length block to have 1-length header in pitem?

		fnFooter.Filenames = append(fnFooter.Filenames, item.Filename)
		itFooter.Items = append(itFooter.Items, itemHeader{
			Offset:      uint32(itemOff),
			Compression: None,
			PayloadLen:  uint32(len(item.Payload)),
		})
	}

	hdrPayload.Store(buf)
	fnFooter.Store(buf)
	itFooter.Store(buf)
}

func (v headerAndPayload) Store(buf *ByteBuffer) {
	if !v.IsPVOL {
		panic("non-PVOL unsupported")
	}
	block{HeaderMagic: magicPVOL, Payload: v.Payload}.Store(true, buf)
}

func (v filenameFooter) Store(buf *ByteBuffer) {
	blk := block{HeaderMagic: magicVOLS}
	for _, fn := range v.Filenames {
		blk.Payload.AppendString(fn)
		blk.Payload.Append(0) // null terminator
	}
	blk.Store(false, buf)
}

func (v itemFooter) Store(buf *ByteBuffer) {
	blk := block{HeaderMagic: magicVOLI}
	for _, it := range v.Items {
		it.Store(&blk.Payload)
	}
	blk.Store(false, buf)
}

func (v itemHeader) Store(buf *ByteBuffer) {
	binary.LittleEndian.PutUint32(buf.Extend(4), v.Unknown1)
	binary.LittleEndian.PutUint32(buf.Extend(4), v.Unknown2)
	binary.LittleEndian.PutUint32(buf.Extend(4), v.Offset)
	binary.LittleEndian.PutUint32(buf.Extend(4), v.PayloadLen)
	buf.Extend(1)[0] = byte(v.Compression)
}

func (v payloadItem) Store(buf *ByteBuffer) {
	block{HeaderMagic: magicVBLK, Payload: v.Payload}.Store(false, buf)
}
