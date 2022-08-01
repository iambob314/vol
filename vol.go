package vol

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// Magic bytes
const (
	magicVOL  = " VOL"
	magicPVOL = "PVOL"
	magicVOLS = "vols"
	magicVOLI = "voli"
	magicVBLK = "VBLK"
)

//go:generate stringer -type=CompressionType
type CompressionType byte

const (
	None = CompressionType(0)
	RLE  = CompressionType(1)
	LZ   = CompressionType(2)
	LZH  = CompressionType(3)
)

type File struct {
	Items []Item
}

type Item struct {
	Filename    string
	Compression CompressionType
	Payload     ByteBuffer
}

func (v *File) Parse(data []byte) error {
	var header header
	var footer footer

	parseBuf := ByteBuffer(data)
	if err := header.Parse(&parseBuf); err != nil {
		return err
	} else if err := footer.Parse(header.IsPVOL, &parseBuf); err != nil {
		return err
	}

	pstart, pend := header.PayloadOffset, header.PayloadOffset+header.PayloadLen
	for i, itemHdr := range footer.Items {
		filename := footer.Filenames[i]

		start, end := itemHdr.Offset, itemHdr.Offset+blockHeaderLen+itemHdr.PayloadLen
		if start < pstart || end > pend {
			return fmt.Errorf("item %d range [%d, %d) out of bounds in payload [%d, %d)", i, start, end, pstart, pend)
		}

		// Stupid special case: for zero-length file, the itemHeader reports 0 length, but the header on the Item itself
		// reports 1 length, so we could crash to parse it. In this case, just abort early with an empty Item, don't
		// check the payload.
		if itemHdr.PayloadLen == 0 {
			v.Items = append(v.Items, Item{Filename: filename, Compression: None, Payload: nil})
			continue
		}
		itemBuf := ByteBuffer(data[start:end])

		var item Item
		if err := item.Parse(filename, itemHdr, &itemBuf); err != nil {
			return fmt.Errorf("parsing item %d (%s) at range [%d, %d): %w", i, filename, start, end, err)
		}
		v.Items = append(v.Items, item)
	}

	return nil
}

func (v *Item) Parse(filename string, hdr itemHeader, buf *ByteBuffer) error {
	v.Filename = filename
	v.Compression = hdr.Compression

	var block block
	if err := block.Parse("item", magicVBLK, 24, false, buf); err != nil {
		return err
	}

	v.Payload = block.Payload
	return nil
}

func (v *Item) Decompress() {
	switch v.Compression {
	case None:
	}
}

type header struct {
	IsPVOL        bool
	PayloadOffset uint32 // payload's offset in file
	PayloadLen    uint32
}

type footer struct {
	Filenames []string
	Items     []itemHeader
}

type itemHeader struct {
	Unknown1    uint32
	Unknown2    uint32
	Offset      uint32 // file offset of item payload, with its header
	PayloadLen  uint32 // length of item payload only, not including header
	Compression CompressionType
}

func (v *header) Parse(buf *ByteBuffer) (err error) {
	var headerAndPayload block
	if err := headerAndPayload.Parse("payload", "", 0, true, buf); err != nil {
		return err
	}

	v.PayloadOffset = headerAndPayload.HeaderLen()
	v.PayloadLen = headerAndPayload.PayloadLen

	switch headerAndPayload.HeaderMagic {
	case magicVOL:
	case magicPVOL:
		v.IsPVOL = true
	default:
		return fmt.Errorf("unexpected payload header magic: got %s, expected %s or %s", headerAndPayload.HeaderMagic, magicVOL, magicVOLS)
	}

	return nil
}

func (v *footer) Parse(isPVOL bool, buf *ByteBuffer) error {
	var filenamesBlock, itemsBlock block

	// Parse filenames block
	if !isPVOL {
		buf.Skip(2 * (2 * 4)) // non-PVOL footer has 2 extra pairs of magic/offset headers (unknown purpose) before the strings section
	}
	if err := filenamesBlock.Parse("footer filenames", magicVOLS, 0, false, buf); err != nil {
		return err
	}
	for len(filenamesBlock.Payload) > 0 {
		nulIdx := bytes.IndexByte(filenamesBlock.Payload, 0)
		if nulIdx == -1 {
			return fmt.Errorf("missing null terminator in filename list")
		}

		fnBytes := filenamesBlock.Payload.MustNext(nulIdx + 1) // guaranteed by index check above
		v.Filenames = append(v.Filenames, string(fnBytes[:nulIdx]))
	}

	// There is sometimes range garbage between the filenames and items blocks; seek forward a limited distance to find
	// the magic header
	maxSeek := 8
	if maxSeek > len(*buf) {
		maxSeek = len(*buf)
	}
	if padding := bytes.Index(*buf, []byte(magicVOLI)); padding > 0 {
		buf.Skip(padding)
	}

	// Parse items block
	if err := itemsBlock.Parse("footer items", magicVOLI, 0, false, buf); err != nil {
		return err
	}
	for len(itemsBlock.Payload) > 0 {
		var item itemHeader
		if err := item.Parse(&itemsBlock.Payload); err != nil {
			return err
		}
		v.Items = append(v.Items, item)
	}
	if len(v.Filenames) != len(v.Items) {
		return fmt.Errorf("footer contains different numbers of filenames and item headers (%d vs. %d)", len(v.Filenames), len(v.Items))
	}

	return nil
}

func (v *itemHeader) Parse(buf *ByteBuffer) error {
	if len(*buf) < 4*4+1 { // 4 uint32 + 1 byte
		return fmt.Errorf("unexpected end of item")
	}
	v.Unknown1 = binary.LittleEndian.Uint32(buf.MustNext(4))
	v.Unknown2 = binary.LittleEndian.Uint32(buf.MustNext(4))
	v.Offset = binary.LittleEndian.Uint32(buf.MustNext(4))
	v.PayloadLen = binary.LittleEndian.Uint32(buf.MustNext(4))
	v.Compression = CompressionType(buf.MustNext(1)[0])
	return nil
}
