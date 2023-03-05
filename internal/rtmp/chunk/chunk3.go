package chunk

import (
	"bufio"
	"io"
)

// Chunk3 is a type 3 chunk.
// Type 3 chunks have no message header. The stream ID, message length
// and timestamp delta fields are not present; chunks of this type take
// values from the preceding chunk for the same Chunk Stream ID. When a
// single message is split into chunks, all chunks of a message except
// the first one SHOULD use this type.
type Chunk3 struct {
	ChunkStreamID int
	Body          []byte
}

// Read reads the chunk.
func (c *Chunk3) Read(r io.Reader, chunkBodyLen uint32) error {
	br := bufio.NewReader(r)
	_, csid, err0 := ReadBasicHeader(br)
	if err0 != nil {
		return err0
	}
	c.ChunkStreamID = csid

	c.Body = make([]byte, chunkBodyLen)
	_, err := io.ReadFull(br, c.Body)
	return err
}

// Marshal writes the chunk.
func (c Chunk3) Marshal() ([]byte, error) {
	header := WriteBasicHeader(byte(3), c.ChunkStreamID)

	buf := make([]byte, len(header)+len(c.Body))
	copy(buf[0:], header)
	copy(buf[len(header):], c.Body)
	return buf, nil
}
