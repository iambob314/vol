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
	var (
		hdrPayload headerAndPayload
		fnFooter   filenameFooter
		itFooter   itemFooter
	)

	parseBuf := ByteBuffer(data)
	if err := hdrPayload.Parse(&parseBuf); err != nil {
		return err
	} else if err := fnFooter.Parse(hdrPayload.IsPVOL, &parseBuf); err != nil {
		return err
	} else if err := itFooter.Parse(&parseBuf); err != nil {
		return err
	}

	if len(fnFooter.Filenames) != len(itFooter.Items) {
		return fmt.Errorf("filenameFooter contains different number of filenames than itemFooter's number of item headers (%d vs. %d)", len(fnFooter.Filenames), len(itFooter.Items))
	}

	pstart, pend := hdrPayload.HeaderLen(), hdrPayload.HeaderLen()+uint32(len(hdrPayload.Payload))
	for i, itemHdr := range itFooter.Items {
		filename := fnFooter.Filenames[i]

		start, end := itemHdr.Offset, itemHdr.Offset+blockHeaderLen+itemHdr.PayloadLen
		if start < pstart || end > pend {
			return fmt.Errorf("item %d range [%d, %d) out of bounds in payload [%d, %d)", i, start, end, pstart, pend)
		}

		item := Item{Filename: filename, Compression: None}

		// Stupid special case: for zero-length item, itemHeader reports length 0, but the header on the Item itself
		// reports length 1, so we may crash to parse it. So for length 0, just append an empty Item, don't read payload.
		if itemHdr.PayloadLen > 0 {
			itemBuf := ByteBuffer(data[start:end])
			var pitem payloadItem
			if err := pitem.Parse(&itemBuf); err != nil {
				return fmt.Errorf("parsing item %d (%s) at range [%d, %d): %w", i, filename, start, end, err)
			}
			item.Payload = pitem.Payload
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
	default:
		panic(fmt.Errorf("unsupported compression type %s", v.Compression))
	}
}

type headerAndPayload struct {
	IsPVOL  bool
	Payload ByteBuffer
}

type filenameFooter struct {
	Filenames []string
}

type itemFooter struct {
	Items []itemHeader
}

type itemHeader struct {
	Unknown1    uint32
	Unknown2    uint32
	Offset      uint32 // file offset of item payload, with its header
	PayloadLen  uint32 // length of item payload only, not including header
	Compression CompressionType
}

type payloadItem struct {
	Payload ByteBuffer
}

func (v *payloadItem) Parse(buf *ByteBuffer) error {
	var block block
	if err := block.Parse("item", magicVBLK, 24, false, buf); err != nil {
		return err
	}

	v.Payload = block.Payload
	return nil
}

func (v *headerAndPayload) Parse(buf *ByteBuffer) (err error) {
	var blk block
	if err := blk.Parse("payload", "", 0, true, buf); err != nil {
		return err
	}

	switch blk.HeaderMagic {
	case magicVOL:
	case magicPVOL:
		v.IsPVOL = true
	default:
		return fmt.Errorf("unexpected payload header magic: got %s, expected %s or %s", blk.HeaderMagic, magicVOL, magicVOLS)
	}

	v.Payload = blk.Payload

	return nil
}

func (v *headerAndPayload) HeaderLen() uint32 { return blockHeaderLen }

func (v *filenameFooter) Parse(isPVOL bool, buf *ByteBuffer) error {
	if !isPVOL {
		buf.Skip(2 * (2 * 4)) // non-PVOL filenameFooter has 2 extra pairs of magic/offset headers (unknown purpose) before the strings section
	}

	var filenamesBlock block
	if err := filenamesBlock.Parse("filenameFooter filenames", magicVOLS, 0, false, buf); err != nil {
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

	return nil
}

func (v *itemFooter) Parse(buf *ByteBuffer) error {
	// There is padding between the filenames and items footers; seek forward a limited distance to find the magic header
	const maxSeek = 8
	if padding := bytes.Index(*buf, []byte(magicVOLI)); padding > maxSeek {
		return fmt.Errorf("filenameFooter could not find magic %s within %d bytes", magicVOLI, maxSeek)
	} else if padding > 0 {
		buf.Skip(padding)
	}

	// Parse items block
	var itemsBlock block
	if err := itemsBlock.Parse("filenameFooter items", magicVOLI, 0, false, buf); err != nil {
		return err
	}
	for len(itemsBlock.Payload) > 0 {
		var item itemHeader
		if err := item.Parse(&itemsBlock.Payload); err != nil {
			return err
		}
		v.Items = append(v.Items, item)
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
