package chunk

import (
	"bufio"
	"io"
)

// Chunk0 is a type 0 chunk.
// This type MUST be used at
// the start of a chunk stream, and whenever the stream timestamp goes
// backward (e.g., because of a backward seek).
type Chunk0 struct {
	ChunkStreamID   int
	Timestamp       uint32
	Type            MessageType
	MessageStreamID uint32
	BodyLen         uint32
	Body            []byte
}

// Read reads the chunk.
func (c *Chunk0) Read(r io.Reader, chunkMaxBodyLen uint32) error {
	br := bufio.NewReader(r)
	_, csid, err0 := ReadBasicHeader(br)
	if err0 != nil {
		return err0
	}
	c.ChunkStreamID = csid

	header := make([]byte, 11)
	_, err := io.ReadFull(br, header)
	if err != nil {
		return err
	}

	c.Timestamp = uint32(header[0])<<16 | uint32(header[1])<<8 | uint32(header[2])
	c.BodyLen = uint32(header[3])<<16 | uint32(header[4])<<8 | uint32(header[5])
	c.Type = MessageType(header[6])
	c.MessageStreamID = uint32(header[7])<<24 | uint32(header[8])<<16 | uint32(header[9])<<8 | uint32(header[10])

	chunkBodyLen := c.BodyLen
	if chunkBodyLen > chunkMaxBodyLen {
		chunkBodyLen = chunkMaxBodyLen
	}

	c.Body = make([]byte, chunkBodyLen)
	_, err = io.ReadFull(br, c.Body)
	return err
}

// Marshal writes the chunk.
func (c Chunk0) Marshal() ([]byte, error) {
	header := WriteBasicHeader(byte(0), c.ChunkStreamID)

	raw := make([]byte, len(header)+11+len(c.Body))
	copy(raw[0:], header)

	buf := raw[len(header):]

	buf[0] = byte(c.Timestamp >> 16)
	buf[1] = byte(c.Timestamp >> 8)
	buf[2] = byte(c.Timestamp)
	buf[3] = byte(c.BodyLen >> 16)
	buf[4] = byte(c.BodyLen >> 8)
	buf[5] = byte(c.BodyLen)
	buf[6] = byte(c.Type)
	buf[7] = byte(c.MessageStreamID >> 24)
	buf[8] = byte(c.MessageStreamID >> 16)
	buf[9] = byte(c.MessageStreamID >> 8)
	buf[10] = byte(c.MessageStreamID)
	copy(buf[11:], c.Body)

	return raw, nil
}
