package tcp

import (
	"errors"
	"io"
	"time"

	"github.com/google/gopacket/tcpassembly"
)

func NewStream() Stream {
	return &stream{}
}

type PeekStream interface {
	Peek(p []byte) (n int, f time.Time, t time.Time, err error)
	Length() int
}

type Stream interface {
	tcpassembly.Stream
	PeekStream
	Notifier(notifier func())
	Read(p []byte) (n int, f time.Time, t time.Time, err error)
	Consume(n int) (err error)
	Clear()
}

type stream struct {
	notifier func()
	chunks   []tcpassembly.Reassembly
}

func (s *stream) Notifier(notifier func()) {
	s.notifier = notifier
}

func (s *stream) Length() (n int) {
	for _, chunk := range s.chunks {
		n += len(chunk.Bytes)
	}
	return n
}

func (s *stream) Clear() {
	s.chunks = s.chunks[:0]
}

func (s *stream) Reassembled(chunks []tcpassembly.Reassembly) {
	s.chunks = append(s.chunks, chunks...)
	if s.notifier != nil {
		s.notifier()
	}
}

func (s *stream) ReassemblyComplete() {
	s.notifier()
}

func (s *stream) Read(p []byte) (n int, f time.Time, t time.Time, err error) {
	n, f, t, err = s.Peek(p)
	if eof := s.Consume(n); eof != err {
		err = errors.New(`consumption discrepancy`)
	}
	return n, f, t, err
}

func (s *stream) Consume(n int) (err error) {
	for n > 0 && len(s.chunks) > 0 {
		l := len(s.chunks[0].Bytes)
		if n < l {
			s.chunks[0].Bytes = s.chunks[0].Bytes[n:]
			n -= n
		} else {
			s.chunks = s.chunks[1:]
			n -= l
		}
	}
	if n > 0 {
		err = io.EOF
	}
	return err
}

func (s *stream) Peek(p []byte) (n int, start time.Time, end time.Time, err error) {
	for i := 0; i < len(s.chunks) && n < len(p); i++ {
		// Copy data
		m := copy(p[n:], s.chunks[i].Bytes)
		// Set the start time for initial reads
		if n == 0 {
			start = s.chunks[i].Seen
		}
		// Set the end time
		end = s.chunks[i].Seen
		// Increment the read bytes
		n += m
	}
	// Generate an io.EOF if there are no more chunks
	if n < len(p) {
		err = io.EOF
	}
	return n, start, end, err
}

func Discard(stream Stream) {
	stream.Notifier(stream.Clear)
	stream.Clear()
}
