package mp3

import (
	"fmt"
)

// Parser is a mpeg-3 parser
type Parser struct {
	samplingFrequency int
}

// NewParser returns a parser
func NewParser() *Parser {
	return &Parser{}
}

// sampling_frequency - indicates the sampling frequency, according to the following table.
// '00' 44.1 kHz
// '01' 48 kHz
// '10' 32 kHz
// '11' reserved
var mp3Rates = []int{44100, 48000, 32000}
var (
	// ErrInvalidMp3Data means invalid mp3 data
	ErrInvalidMp3Data = fmt.Errorf("invalid mp3data")
	// ErrInvalidIndex means invalid rate index
	ErrInvalidIndex = fmt.Errorf("invalid rate index")
)

// Parse parse the data
func (parser *Parser) Parse(src []byte) error {
	if len(src) < 3 {
		return ErrInvalidMp3Data
	}
	index := (src[2] >> 2) & 0x3
	if index <= byte(len(mp3Rates)-1) {
		parser.samplingFrequency = mp3Rates[index]
		return nil
	}
	return ErrInvalidIndex
}

// SampleRate returns the sampling rate
func (parser *Parser) SampleRate() int {
	if parser.samplingFrequency == 0 {
		parser.samplingFrequency = 44100
	}
	return parser.samplingFrequency
}
