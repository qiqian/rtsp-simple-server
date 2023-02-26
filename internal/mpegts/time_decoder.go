package mpegts

import (
	"time"
)

const (
	maximum           = 0x1FFFFFFFF // 33 bits
	negativeThreshold = 0x1FFFFFFFF / 2
	clockRate         = 90000
)

// TimeDecoder is a MPEG-TS timestamp decoder.
type TimeDecoder struct {
	overall time.Duration
	prev    int64
}

// NewTimeDecoder allocates a TimeDecoder.
func NewTimeDecoder(start int64) *TimeDecoder {
	return &TimeDecoder{
		prev: start,
	}
}

// Decode decodes a MPEG-TS timestamp.
func (d *TimeDecoder) Decode(ts int64) time.Duration {
	diff := (ts - d.prev) & maximum

	// negative difference
	if diff > negativeThreshold {
		diff = (d.prev - ts) & maximum
		d.prev = ts
		d.overall -= time.Duration(diff)
	} else {
		d.prev = ts
		d.overall += time.Duration(diff)
	}

	// avoid an int64 overflow and preserve resolution by splitting division into two parts:
	// first add the integer part, then the decimal part.
	secs := d.overall / clockRate
	dec := d.overall % clockRate
	return secs*time.Second + dec*time.Second/clockRate
}
