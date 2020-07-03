package av

import (
	"fmt"
	"io"
)

// Tag definitions
const (
	// TagAudio denotes the audio tag
	TagAudio = 8
	// TagVideo denotes the video tag
	TagVideo = 9
	// TagScriptDataAMF0 denotes the script data AMF0 tag
	TagScriptDataAMF0 = 18
	// TagScriptDataAMF3 denotes the script data AMF3 tag
	TagScriptDataAMF3 = 0xf
)

// Metadata definitions
const (
	// MetadataAMF0 denotes the metadata of AMF0
	MetadataAMF0 = 0x12
	// MetadataAMF3 denotes the metadata of AMF3
	MetadataAMF3 = 0xf
)

// Sound definitions
const (
	// SoundMP3 denotes the codec of sound is MP3
	SoundMP3 = 2
	// SoundNellymoser16kHzMono denotes the codec of sound is Nellymoser 16kHz mono
	SoundNellymoser16kHzMono = 4
	// SoundNellymoser8kHzMono denotes the codec of sound is Nellymoser 8kHz mono
	SoundNellymoser8kHzMono = 5
	// SoundNellymoser denotes the codec of sound is Nellymoser
	SoundNellymoser = 6
	// SoundALaw denotes the codec of sound is A-law
	SoundALaw = 7
	// SoundMuLaw denotes the codec of sound is Mu-law
	SoundMuLaw = 8
	// SoundAAC denotes the codec of sound is acc
	SoundAAC = 10
	// SoundSpeex denotes the codec of sound is speex
	SoundSpeex = 11

	// Sound55kHz denotes the sampling of sound is 5 5kHz
	Sound55kHz = 0
	// Sound11kHz denotes the sampling of sound is 11kHz
	Sound11kHz = 1
	// Sound22kHz denotes the sampling of sound is 22kHz
	Sound22kHz = 2
	// Sound44kHz denotes the sampling of sound is 44kHz
	Sound44kHz = 3

	// Sound8Bit denotes sound is 8bit
	Sound8Bit = 0
	// Sound16Bit denotes sound is 16bit
	Sound16Bit = 1

	// SoundMono denotes sound is mono
	SoundMono = 0
	// SoundStereo denotes sound is stereo
	SoundStereo = 1

	// AACSeqHeader denotes the AAC Sequence Header
	AACSeqHeader = 0
	// AACRaw denotes the AAC raw
	AACRaw = 1
)

// H.264/AVC definitions
const (
	// AVCSeqHeader denotes the AVC Sequence Header
	AVCSeqHeader = 0
	// AVCNalu denotes the AVC NALU
	AVCNalu = 1
	// AVCEos denotes the AVC EOS
	AVCEos = 2

	// FrameKey denotes the frame key
	FrameKey = 1
	// FrameInter denotes the frame inter
	FrameInter = 2

	// VideoH264 denotes the video is H.264
	VideoH264 = 7
)

var (
	// PUBLISH denotes publish
	PUBLISH = "publish"
	// PLAY denotes play
	PLAY = "play"
)

// Packet is the av packet
type Packet struct {
	IsAudio    bool
	IsVideo    bool
	IsMetadata bool
	TimeStamp  uint32 // dts
	StreamID   uint32
	Header     PacketHeader
	Data       []byte
}

// PacketHeader can be converted to AudioHeaderInfo or VideoHeaderInfo
type PacketHeader interface {
}

// AudioPacketHeader is the packet header of audio
type AudioPacketHeader interface {
	PacketHeader
	SoundFormat() uint8
	AACPacketType() uint8
}

// VideoPacketHeader is the packet header of video
type VideoPacketHeader interface {
	PacketHeader
	IsKeyFrame() bool
	IsSeq() bool
	CodecID() uint8
	CompositionTime() int32
}

// Demuxer demux the packet
type Demuxer interface {
	Demux(*Packet) (ret *Packet, err error)
}

// Muxer mux the packet to Writer
type Muxer interface {
	Mux(*Packet, io.Writer) error
}

// SampleRater can return sample rate
type SampleRater interface {
	SampleRate() (int, error)
}

// CodecParser parse the packet
type CodecParser interface {
	SampleRater
	Parse(*Packet, io.Writer) error
}

// GetWriter get WriteCloser from Info
type GetWriter interface {
	Writer(Info) WriteCloser
}

// Handler handle reader and writer
type Handler interface {
	HandleReader(ReadCloser)
	HandleWriter(WriteCloser)
}

// Aliver can return if this is alive
type Aliver interface {
	// Alive return if this is alive
	Alive() bool
}

// Closer returns Info and can close
type Closer interface {
	// Info return the info
	Info() Info
	// Close close current
	Close(error)
}

// CalcTimer calculate base timestamp
type CalcTimer interface {
	CalcBaseTimestamp()
}

// Info is the information
type Info struct {
	Key   string
	URL   string
	UID   string
	Inter bool
}

// IsInterval returns if this is interval
func (info Info) IsInterval() bool {
	return info.Inter
}

// String returns the representation string
func (info Info) String() string {
	return fmt.Sprintf("<key: %s, URL: %s, UID: %s, Inter: %v>",
		info.Key, info.URL, info.UID, info.Inter)
}

// ReadCloser read to packet
type ReadCloser interface {
	Closer
	Aliver
	Read(*Packet) error
}

// WriteCloser write to packet
type WriteCloser interface {
	Closer
	Aliver
	CalcTimer
	Write(*Packet) error
}
