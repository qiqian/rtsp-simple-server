package chunk

import (
	"bufio"
	"io"
)

// Chunk2 is a type 2 chunk.
// Neither the stream ID nor the
// message length is included; this chunk has the same stream ID and
// message length as the preceding chunk.
type Chunk2 struct {
	ChunkStreamID  int
	TimestampDelta uint32
	Body           []byte
}

// Read reads the chunk.
func (c *Chunk2) Read(r io.Reader, chunkBodyLen uint32) error {
	br := bufio.NewReader(r)
	_, csid, err0 := ReadBasicHeader(br)
	if err0 != nil {
		return err0
	}
	c.ChunkStreamID = csid

	header := make([]byte, 3)
	_, err := io.ReadFull(br, header)
	if err != nil {
		return err
	}

	c.TimestampDelta = uint32(header[0])<<16 | uint32(header[1])<<8 | uint32(header[2])

	c.Body = make([]byte, chunkBodyLen)
	_, err = io.ReadFull(br, c.Body)
	return err
}

// Marshal writes the chunk.
func (c Chunk2) Marshal() ([]byte, error) {
	header := WriteBasicHeader(byte(2), c.ChunkStreamID)

	raw := make([]byte, len(header)+3+len(c.Body))
	copy(raw[0:], header)

	buf := raw[len(header):]
	buf[0] = byte(c.TimestampDelta >> 16)
	buf[1] = byte(c.TimestampDelta >> 8)
	buf[2] = byte(c.TimestampDelta)
	copy(buf[3:], c.Body)

	return raw, nil
}
