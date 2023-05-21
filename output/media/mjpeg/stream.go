package mjpeg

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"path"
	"time"

	"github.com/icza/mjpeg"
)

type stream struct {
	frame     image.Image
	next      time.Time
	writer    mjpeg.AviWriter
	Directory string
	Quality   int
	FPS       int
	ID        int
}

func (s *stream) Write(image image.Image, t time.Time) (err error) {
	// Create a writer if needed
	if s.writer == nil {
		name := fmt.Sprintf("%s.%02d.avi", t.UTC().Format("2006-01-02T15-04-05,000"), s.ID)
		if s.writer, err = mjpeg.New(path.Join(s.Directory, name), int32(image.Bounds().Size().X), int32(image.Bounds().Size().Y), int32(s.FPS)); err != nil {
			return err
		}
	}
	// Handle first frames
	if s.next.IsZero() {
		s.next = t.Add(time.Second / time.Duration(s.FPS))
	}
	if s.next.Before(t) {
		// Encode previous frames only once
		b := &bytes.Buffer{}
		if err := jpeg.Encode(b, s.frame, &jpeg.Options{Quality: s.Quality}); err != nil {
			return err
		}
		d, err := io.ReadAll(b)
		if err != nil {
			return err
		}
		// And keep adding frames until we meet the FPS
		for s.next.Before(t) {
			if err = s.writer.AddFrame(d); err != nil {
				return err
			}
			s.next = s.next.Add(time.Second / time.Duration(s.FPS))
		}
	}
	s.frame = image
	return nil
}

func (s *stream) Close() error {
	// Make sure to always append the last frame
	b := &bytes.Buffer{}
	if err := jpeg.Encode(b, s.frame, &jpeg.Options{Quality: s.Quality}); err == nil {
		d, err := io.ReadAll(b)
		if err == nil {
			_ = s.writer.AddFrame(d)
		}
	}
	return s.writer.Close()
}
