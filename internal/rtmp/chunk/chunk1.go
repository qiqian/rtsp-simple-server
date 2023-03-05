package chunk

import (
	"bufio"
	"io"
)

// Chunk1 is a type 1 chunk.
// The message stream ID is not
// included; this chunk takes the same stream ID as the preceding chunk.
// Streams with variable-sized messages (for example, many video
// formats) SHOULD use this format for the first chunk of each new
// message after the first.
type Chunk1 struct {
	ChunkStreamID  int
	TimestampDelta uint32
	Type           MessageType
	BodyLen        uint32
	Body           []byte
}

// Read reads the chunk.
func (c *Chunk1) Read(r io.Reader, chunkMaxBodyLen uint32) error {
	br := bufio.NewReader(r)
	_, csid, err0 := ReadBasicHeader(br)
	if err0 != nil {
		return err0
	}
	c.ChunkStreamID = csid

	header := make([]byte, 7)
	_, err := io.ReadFull(br, header)
	if err != nil {
		return err
	}

	c.TimestampDelta = uint32(header[0])<<16 | uint32(header[1])<<8 | uint32(header[2])
	c.BodyLen = uint32(header[3])<<16 | uint32(header[4])<<8 | uint32(header[5])
	c.Type = MessageType(header[6])

	chunkBodyLen := (c.BodyLen)
	if chunkBodyLen > chunkMaxBodyLen {
		chunkBodyLen = chunkMaxBodyLen
	}

	c.Body = make([]byte, chunkBodyLen)
	_, err = io.ReadFull(br, c.Body)
	return err
}

// Marshal writes the chunk.
func (c Chunk1) Marshal() ([]byte, error) {
	header := WriteBasicHeader(byte(1), c.ChunkStreamID)

	raw := make([]byte, len(header)+7+len(c.Body))
	copy(raw[0:], header)

	buf := raw[len(header):]

	buf[0] = byte(c.TimestampDelta >> 16)
	buf[1] = byte(c.TimestampDelta >> 8)
	buf[2] = byte(c.TimestampDelta)
	buf[3] = byte(c.BodyLen >> 16)
	buf[4] = byte(c.BodyLen >> 8)
	buf[5] = byte(c.BodyLen)
	buf[6] = byte(c.Type)
	copy(buf[7:], c.Body)

	return raw, nil
}
