package vol

import (
	"encoding/binary"
	"fmt"
)

// block is a standard format used as a subunit of the vol file format.
type block struct {
	HeaderMagic string
	PayloadLen  uint32
	Payload     ByteBuffer
}

const blockHeaderLen = 8

func (m *block) HeaderLen() uint32 { return blockHeaderLen }

// Parse parses a block of the form:
//   HeaderMagic [4]byte
//   _ [gap]byte // discard gap bytes
//   PayloadLen uint32 // after parsing, keep only lowest lenBits bits (or all bits if lenBits = 0), then subtract
//   Payload [PayloadLen]byte
//
// Before parsing Payload, PayloadLen is adjusted as follows:
// * If lenBits != 0, all but the lowest lenBits are cleared
// * If lenInclHeader, PayloadLen is decremented by (HeaderLen = 8+gap)
//
// An error is returned HeaderMagic != expectMagic (not checked if expectMagic == "") or other
// parse error occurs. If non-nil, the error will include blockName in the message.
func (m *block) Parse(blockName string, expectMagic string, lenBits int, lenInclHeader bool, buf *ByteBuffer) error {
	hdr, ok := buf.Next(blockHeaderLen)
	if !ok {
		return fmt.Errorf("unexpected end of %s header (expected %d bytes, only %d left)", blockName, m.HeaderLen, len(*buf))
	}

	m.HeaderMagic = string(hdr[:4])
	m.PayloadLen = binary.LittleEndian.Uint32(hdr[4:])

	if expectMagic != "" && m.HeaderMagic != expectMagic {
		return fmt.Errorf("unexpected %s header magic: got %s, expected %s", blockName, m.HeaderMagic, expectMagic)
	}
	if lenBits != 0 {
		m.PayloadLen &^= ^uint32(0) << lenBits // mask out high bits
	}
	if lenInclHeader {
		m.PayloadLen -= blockHeaderLen
	}

	m.Payload, ok = buf.Next(int(m.PayloadLen))
	if !ok {
		return fmt.Errorf("unexpected end of %s payload (expected %d bytes, only %d left)", blockName, m.PayloadLen, len(*buf))
	}

	return nil
}

func (m *block) Store(lenInclHeader bool, buf *ByteBuffer) error {
	payloadLen := m.PayloadLen
	if lenInclHeader {
		payloadLen += blockHeaderLen
	}

	buf.Append([]byte(m.HeaderMagic))
	binary.LittleEndian.PutUint32(buf.Extend(4), payloadLen)
	buf.Append(m.Payload)
	return nil
}
