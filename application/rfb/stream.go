package rfb

import (
	"bytes"
	"compress/zlib"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	"image/color"
	"io"
	"math"
	"time"

	"github.com/0xThiebaut/PCAPeek/output/files"
	"github.com/0xThiebaut/PCAPeek/output/media"
	"github.com/0xThiebaut/PCAPeek/transport/tcp"
	"github.com/google/gopacket"
)

type stream struct {
	Negotiate  bool
	net        gopacket.Flow
	transport  gopacket.Flow
	client     tcp.Stream
	server     tcp.Stream
	encodings  []Encoding
	rectangles uint16
	date       time.Time
	format     PixelFormat
	frame      *image.RGBA
	zw         io.Writer
	zr         io.Reader
	id         uint64
	Media      media.Factory
	mstream    media.Stream
	Files      files.Factory
}

func (s *stream) hasExtendedClipboard() bool {
	for _, encoding := range s.encodings {
		if encoding == PseudoEncodingExtendedClipBoard {
			return true
		}
	}
	return false
}

func (s *stream) hasZlib() bool {
	for _, encoding := range s.encodings {
		if encoding == EncodingZlib {
			return true
		}
	}
	return false
}

func (s *stream) ClientInit() {
	// Expect the next server message to be a ServerInit
	s.server.Notifier(s.ServerInit)
	// Discard client data until then
	s.client.Notifier(s.client.Clear)
}

func (s *stream) ServerInit() {
	init := ServerInit{}
	t, err := tcp.Unmarshall(s.server, &init)
	if err != nil {
		tcp.Discard(s.server)
		return
	}
	s.format = init.ServerPixelFormat
	s.frame = image.NewRGBA(image.Rect(0, 0, int(init.FramebufferWidth), int(init.FramebufferHeight)))
	s.mstream = s.Media.New()
	go func() {
		<-context.Background().Done()
		_ = s.mstream.Close()
	}()
	fmt.Printf("Got ServerInit at %s from %s:%s to %s:%s named %q of ratio %dx%d\n", t.UTC().Format(time.RFC3339Nano), s.net.Dst().String(), s.transport.Dst().String(), s.net.Src().String(), s.transport.Src().String(), init.Name, init.FramebufferWidth, init.FramebufferHeight)
	// Expect the next messages to be client-to-server or server-to-client messages
	s.server.Notifier(s.Server)
	s.client.Notifier(s.Client)
}

var (
	ErrUnhandledMessage        = errors.New(`message not handled`)
	ErrUnhandledClientMessage  = fmt.Errorf(`client %w`, ErrUnhandledMessage)
	ErrUnhandledServerMessage  = fmt.Errorf(`server %w`, ErrUnhandledMessage)
	ErrUnhandledServerEncoding = fmt.Errorf(`%w due to unknown encoding`, ErrUnhandledServerMessage)
	ErrNewStreamNotifier       = fmt.Errorf("%w: new stream notifier", io.EOF)
)

func (s *stream) Client() {
	for {
		if err := s.onClient(); err != nil {
			if !errors.Is(err, io.EOF) {
				tcp.Discard(s.client)
				fmt.Println(err)
			}
			return
		}
	}
}

func (s *stream) Server() {
	for {
		if err := s.onServer(); err != nil {
			if !errors.Is(err, io.EOF) {
				tcp.Discard(s.server)
				fmt.Println(err)
			}
			return
		}
	}
}

func (s *stream) serverRectangle() {
	for ; s.rectangles > 0; s.rectangles-- {
		if err := s.onServerFramebufferUpdateRectangle(); err != nil {
			if !errors.Is(err, io.EOF) {
				tcp.Discard(s.server)
				fmt.Println(err)
			}
			return
		}
	}
	if s.rectangles == 0 {
		if err := s.mstream.Write(s.frame, s.date); err != nil {
			fmt.Println(err)
		}
		s.server.Notifier(s.Server)
	}
}

func (s *stream) onClient() error {
	var t ClientMessageType
	if _, err := tcp.UnmarshallPeek(s.client, &t); err != nil {
		return err
	}
	switch t {
	case TypeSetPixelFormat:
		if err := s.onClientSetPixelFormat(); err != nil {
			return err
		}
	case TypeFixColourMapEntries:
		// In some observed cases the ClientInit is observed after a ServerInit.s
		if s.client.Length() == 1 {
			s.client.Clear()
		} else {
			return fmt.Errorf(`%w (message type %d)`, ErrUnhandledClientMessage, t)
		}
	case TypeSetEncodings:
		if err := s.onClientSetEncodings(); err != nil {
			return err
		}
	case TypeFramebufferUpdateRequest:
		if err := s.onClientFramebufferUpdateRequest(); err != nil {
			return err
		}
	case TypePointerEvent:
		if err := s.onClientPointerEvent(); err != nil {
			return err
		}
	case TypeKeyEvent:
		if err := s.onClientKeyEvent(); err != nil {
			return err
		}
	case TypeClientCutTex:
		if err := s.onClientCutText(); err != nil {
			return err
		}
	default:
		return fmt.Errorf(`%w (message type %d)`, ErrUnhandledClientMessage, t)
	}
	return nil
}

func (s *stream) onClientSetPixelFormat() error {
	var message SetPixelFormat
	_, err := tcp.Unmarshall(s.client, &message)
	if err != nil {
		return err
	}
	if s.Negotiate {
		// Some servers cannot negotiate (typically backdoors) in which case we retain server preferences
		s.format = message.PixelFormat
	}
	return nil
}

func (s *stream) onClientFixColourMapEntries() error {
	var message FixColourMapEntries
	_, err := tcp.Unmarshall(s.client, &message)
	if err != nil {
		return err
	}
	return nil
}

func (s *stream) onClientSetEncodings() error {
	var message SetEncodings
	_, err := tcp.Unmarshall(s.client, &message)
	if err != nil {
		return err
	}
	s.encodings = message.Encodings
	return nil
}

func (s *stream) onClientFramebufferUpdateRequest() error {
	var message FramebufferUpdateRequest
	_, err := tcp.Unmarshall(s.client, &message)
	return err
}

func (s *stream) onClientPointerEvent() error {
	var message PointerEvent
	_, err := tcp.Unmarshall(s.client, &message)
	if err != nil {
		return err
	}
	return nil
}

func (s *stream) onClientKeyEvent() error {
	var message KeyEvent
	_, err := tcp.Unmarshall(s.client, &message)
	if err != nil {
		return err
	}
	return nil
}

func (s *stream) onClientCutText() error {
	if s.hasExtendedClipboard() {
		var header ExtendedCutTextHeader
		_, err := tcp.UnmarshallPeek(s.client, &header)
		if err != nil {
			return err
		}
		if header.Length < 0 {
			var message ExtendedCutText
			message.Text = make([]uint8, -header.Length-4)
			t, err := tcp.Unmarshall(s.client, &message)
			if err == nil {
				f := s.Files.New()
				defer f.Close()
				return f.Write(bytes.NewReader(message.Text), t)
			}
			return err
		}
	}
	var message CutText
	t, err := tcp.Unmarshall(s.client, &message)
	if err != nil {
		return err
	}
	f := s.Files.New()
	defer f.Close()
	return f.Write(bytes.NewReader(message.Text), t)
}

func (s *stream) onServer() error {
	var t ServerMessageType
	if _, err := tcp.UnmarshallPeek(s.server, &t); err != nil {
		return err
	}
	switch t {
	case TypeFramebufferUpdate:
		if err := s.onServerFramebufferUpdate(); err != nil {
			return err
		}
	case TypeServerCutText:
		if err := s.onServerCutText(); err != nil {
			return err
		}
	default:
		return fmt.Errorf(`%w (message type %d)`, ErrUnhandledServerMessage, t)
	}
	return nil
}

func (s *stream) onServerFramebufferUpdate() error {
	var message FramebufferUpdate
	t, err := tcp.Unmarshall(s.server, &message)
	if err != nil {
		return err
	}
	s.date = t
	s.rectangles = message.NumberOfRectangles
	s.server.Notifier(s.serverRectangle)
	return ErrNewStreamNotifier
}

func (s *stream) onServerCutText() error {
	if s.hasExtendedClipboard() {
		var header ExtendedCutTextHeader
		_, err := tcp.UnmarshallPeek(s.server, &header)
		if err != nil {
			return err
		}
		if header.Length < 0 {
			var message ExtendedCutText
			message.Text = make([]uint8, -header.Length-4)
			t, err := tcp.Unmarshall(s.server, &message)
			if err == nil {
				f := s.Files.New()
				defer f.Close()
				return f.Write(bytes.NewReader(message.Text), t)
			}
			return err
		}
	}
	var message CutText
	t, err := tcp.Unmarshall(s.server, &message)
	if err != nil {
		return err
	}
	f := s.Files.New()
	defer f.Close()
	return f.Write(bytes.NewReader(message.Text), t)
}

func (s *stream) onServerFramebufferUpdateRectangle() error {
	var header Rectangle
	_, err := tcp.UnmarshallPeek(s.server, &header)
	if err != nil {
		return err
	}
	switch header.Encoding {
	case EncodingZlib:
		// Parse the ZLIB rectangle
		var rectangle ZlibRectangle
		_, err = tcp.Unmarshall(s.server, &rectangle)
		if err != nil {
			return err
		}
		// Initiate or populate the ZLIB stream
		if s.zw == nil && s.hasZlib() {
			b := &bytes.Buffer{}
			s.zw = b
			if _, err = s.zw.Write(rectangle.Data); err != nil {
				return err
			}
			if s.zr, err = zlib.NewReader(b); err != nil {
				return err
			}
		} else if _, err = s.zw.Write(rectangle.Data); err != nil {
			return err
		}

		// Define the byte order
		var byteorder binary.ByteOrder
		if s.format.BigEndianFlag == 0 {
			byteorder = binary.LittleEndian
		} else {
			byteorder = binary.BigEndian
		}

		for i := uint16(0); i < rectangle.Height; i++ {
			for j := uint16(0); j < rectangle.Width; j++ {
				// Read the pixel data
				data := make([]byte, s.format.BitsPerPixel/8)
				if n, err := s.zr.Read(data); err != nil {
					return err
				} else if n != len(data) {
					return io.EOF
				}

				// Format the data
				switch s.format.BitsPerPixel {
				case 32:
					pixel := byteorder.Uint32(data)
					s.frame.Set(int(rectangle.X+j), int(rectangle.Y+i), color.RGBA{
						R: uint8((pixel >> s.format.RedShift) & uint32(s.format.RedMax)),
						G: uint8((pixel >> s.format.GreenShift) & uint32(s.format.GreenMax)),
						B: uint8((pixel >> s.format.BlueShift) & uint32(s.format.BlueMax)),
						A: math.MaxUint8,
					})
				default:
					return errors.New(`unhandled bits per pixel value`)
				}
			}
		}
	case PseudoEncodingCursor:
		var rectangle CursorRectangle
		rectangle.Pixels = make([]byte, uint64(header.Width)*uint64(header.Height)*(uint64(s.format.BitsPerPixel)/8))
		rectangle.Bitmask = make([]uint8, ((uint64(header.Width)+7)/8)*uint64(header.Height))
		_, err = tcp.Unmarshall(s.server, &rectangle)
	case PseudoEncodingXCursor:
		var rectangle XCursorRectangle
		rectangle.Bitmap = make([]uint8, ((uint64(header.Width)+7)/8)*uint64(header.Height))
		rectangle.Bitmask = make([]uint8, ((uint64(header.Width)+7)/8)*uint64(header.Height))
		_, err = tcp.Unmarshall(s.server, &rectangle)
	default:
		return fmt.Errorf("%w of type %d", ErrUnhandledServerEncoding, header.Encoding)
	}
	return err
}
