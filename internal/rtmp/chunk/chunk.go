// Package chunk implements RTMP chunks.
package chunk

import (
	"bufio"
	"io"
)

// Chunk is a chunk.
type Chunk interface {
	Read(io.Reader, uint32) error
	Marshal() ([]byte, error)
}

func ReadBasicHeader(br *bufio.Reader) (byte, int, error) {
	byt, err := br.ReadByte()
	if err != nil {
		return 0, 0, err
	}

	typ := byt >> 6
	chunkStreamID := int(byt & 0x3F)
	if chunkStreamID == 0 {
		code, err := br.ReadByte()
		if err != nil {
			return 0, 0, err
		}
		chunkStreamID = int(code) + 64
	} else if chunkStreamID == 1 {
		code1, err1 := br.ReadByte()
		if err1 != nil {
			return 0, 0, err1
		}
		code2, err2 := br.ReadByte()
		if err2 != nil {
			return 0, 0, err2
		}
		chunkStreamID = ((int(code1) << 8) | int(code2)) + 64
	}

	return typ, chunkStreamID, nil
}

func WriteBasicHeader(typ byte, csid int) []byte {
	basicHeaderLen := 1
	if csid > 319 {
		basicHeaderLen = 3
	} else if csid >= 64 {
		basicHeaderLen = 2
	}

	header := make([]byte, basicHeaderLen)
	header[0] = typ << 6

	if basicHeaderLen == 2 {
		header[1] = byte(csid - 64)
	} else if basicHeaderLen == 3 {
		csid -= 64
		header[1] = byte(csid >> 8)
		header[2] = byte(csid & 0xff)
	} else {
		header[0] |= byte(csid)
	}

	return header
}
