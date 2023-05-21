package rfb

import (
	"encoding/json"
	"strconv"
	"unicode"
)

type PixelFormat struct {
	BitsPerPixel  uint8
	Depth         uint8
	BigEndianFlag uint8
	TrueColorFlag uint8
	RedMax        uint16
	GreenMax      uint16
	BlueMax       uint16
	RedShift      uint8
	GreenShift    uint8
	BlueShift     uint8
	Padding       [3]uint8
}

type ServerInit struct {
	FramebufferWidth  uint16
	FramebufferHeight uint16
	ServerPixelFormat PixelFormat
	NameLength        uint32 `tcp:",Name"`
	Name              string
}

type ClientMessageType uint8

func (t ClientMessageType) MarshalJSON() ([]byte, error) {
	if s, ok := map[ClientMessageType]string{
		TypeSetPixelFormat:           `SetPixelFormat`,
		TypeSetEncodings:             `SetEncodings`,
		TypeFramebufferUpdateRequest: `FramebufferUpdateRequest`,
		TypeKeyEvent:                 `KeyEvent`,
		TypePointerEvent:             `PointerEvent`,
		TypeClientCutTex:             `CutText`,
	}[t]; ok {
		return json.Marshal(s)
	} else {
		return json.Marshal(uint8(t))
	}
}

const (
	TypeSetPixelFormat           ClientMessageType = 0
	TypeFixColourMapEntries      ClientMessageType = 1
	TypeSetEncodings             ClientMessageType = 2
	TypeFramebufferUpdateRequest ClientMessageType = 3
	TypeKeyEvent                 ClientMessageType = 4
	TypePointerEvent             ClientMessageType = 5
	TypeClientCutTex             ClientMessageType = 6
)

type SetPixelFormat struct {
	MessageType ClientMessageType
	Padding     [3]uint8
	PixelFormat PixelFormat
}

type FixColourMapEntries struct {
	MessageType     ClientMessageType
	Padding         uint8
	FirstColour     uint16
	NumberOfColours uint16 `tcp:",RGBIntensities"`
	RGBIntensities  []RGBIntensity
}

type RGBIntensity struct {
	Red   uint16
	Green uint16
	Blue  uint16
}

type Encoding int32

func (e Encoding) MarshalJSON() ([]byte, error) {
	if s, ok := map[Encoding]string{
		EncodingRaw:                     `Raw`,
		EncodingCopyRect:                `CopyRect`,
		EncodingRRE:                     `RRE`,
		EncodingHextile:                 `Hextile`,
		EncodingZlib:                    `ZLIB`,
		EncodingTRLE:                    `TRLE`,
		EncodingZRLE:                    `ZRLE`,
		EncodingJPEG:                    `JPEG`,
		EncodingJRLE:                    `JRLE`,
		EncodingZRLE2:                   `ZRLE2`,
		PseudoEncodingDesktopSize:       `Desktop Size`,
		PseudoEncodingCursor:            `Cursor`,
		PseudoEncodingCursorWithAlpha:   `Cursor with Alpha`,
		PseudoEncodingExtendedClipBoard: `Extended Clipboard`,
	}[e]; ok {
		return json.Marshal(s)
	} else {
		return json.Marshal(int32(e))
	}
}

// https://www.iana.org/assignments/rfb/rfb.xml#table-rfb-4
// https://github.com/ultravnc/UltraVNC/blob/ee9954b90ab6b52a2332b349d55f6a98af3f7424/rfb/rfbproto.h#L460-L503
const (
	EncodingRaw                               Encoding = 0
	EncodingCopyRect                          Encoding = 1
	EncodingRRE                               Encoding = 2
	EncodingCoRRE                             Encoding = 4
	EncodingHextile                           Encoding = 5
	EncodingZlib                              Encoding = 6
	EncodingTight                             Encoding = 7
	EncodingZlibHex                           Encoding = 8
	EncodingUltraVNC                          Encoding = 9
	EncodingUltraVNC2                         Encoding = 10
	EncodingTRLE                              Encoding = 15
	EncodingZRLE                              Encoding = 16
	EncodingXZ                                Encoding = 18
	EncodingXZYW                              Encoding = 19
	EncodingJPEG                              Encoding = 21
	EncodingJRLE                              Encoding = 22
	EncodingZRLE2                             Encoding = 24
	EncodingZSTD                              Encoding = 25
	EncodingTightZSTD                         Encoding = 26
	EncodingZSTDHex                           Encoding = 27
	EncodingZSTDRLE                           Encoding = 28
	EncodingZSTDYWRLE                         Encoding = 29
	EncodingOpenH264                          Encoding = 50
	EncodingTightPNG                          Encoding = -260
	PseudoEncodingJPEGQuality10               Encoding = -23
	PseudoEncodingJPEGQuality9                Encoding = -24
	PseudoEncodingJPEGQuality8                Encoding = -25
	PseudoEncodingJPEGQuality7                Encoding = -26
	PseudoEncodingJPEGQuality6                Encoding = -27
	PseudoEncodingJPEGQuality5                Encoding = -28
	PseudoEncodingJPEGQuality4                Encoding = -29
	PseudoEncodingJPEGQuality3                Encoding = -20
	PseudoEncodingJPEGQuality2                Encoding = -31
	PseudoEncodingJPEGQuality1                Encoding = -32
	PseudoEncodingDesktopSize                 Encoding = -223
	PseudoEncodingLastRect                    Encoding = -224
	PseudoEncodingTightPointerPosition        Encoding = -232
	PseudoEncodingCursor                      Encoding = -239
	PseudoEncodingXCursor                     Encoding = -240
	PseudoEncodingCompressionLevel10          Encoding = -247
	PseudoEncodingCompressionLevel9           Encoding = -248
	PseudoEncodingCompressionLevel8           Encoding = -249
	PseudoEncodingCompressionLevel7           Encoding = -250
	PseudoEncodingCompressionLevel6           Encoding = -251
	PseudoEncodingCompressionLevel5           Encoding = -252
	PseudoEncodingCompressionLevel4           Encoding = -253
	PseudoEncodingCompressionLevel3           Encoding = -254
	PseudoEncodingCompressionLevel2           Encoding = -255
	PseudoEncodingCompressionLevel1           Encoding = -256
	PseudoEncodingExtendedDesktopSize         Encoding = -308
	PseudoEncodingCursorWithAlpha             Encoding = -314
	PseudoEncodingUltraVNCEnableIdleTime      Encoding = -32764
	PseudoEncodingUltraVNCPseudoSession       Encoding = -32765
	PseudoEncodingUltraVNCFTProtocolVersion   Encoding = -32766
	PseudoEncodingUltraVNCEnableKeepAlive     Encoding = -32767
	PseudoEncodingUltraVNCServerState         Encoding = -32768
	PseudoEncodingUltraVNCEncodingQueueEnable Encoding = -65525
	PseudoEncodingExtendedClipBoard           Encoding = -1063131698
)

type SetEncodings struct {
	MessageType       ClientMessageType
	Padding           uint8
	NumberOfEncodings uint16 `tcp:",Encodings"`
	Encodings         []Encoding
}

type FramebufferUpdateRequest struct {
	MessageType ClientMessageType
	Incremental uint8
	X           uint16
	Y           uint16
	Width       uint16
	Height      uint16
}

type PointerEvent struct {
	MessageType ClientMessageType
	ButtonMask  uint8
	X           uint16
	Y           uint16
}

type KeyEvent struct {
	MessageType ClientMessageType
	DownFlag    uint8
	Padding     [2]uint8
	Key         KeySym
}

type KeySym uint32

func (k KeySym) String() string {
	if r := rune(k); unicode.IsLetter(r) || unicode.IsNumber(r) || unicode.IsSpace(r) {
		return string(rune(k))
	} else {
		return `\u` + strconv.FormatInt(int64(k), 16)
	}
}

type CutText struct {
	MessageType ClientMessageType
	Padding     [3]uint8
	Length      uint32 `tcp:",Text"`
	Text        []uint8
}

type ExtendedCutTextHeader struct {
	MessageType ClientMessageType
	Padding     [3]uint8
	Length      int32
}

type ExtendedCutText struct {
	ExtendedCutTextHeader
	Flags uint32
	Text  []uint8
}

type ServerMessageType uint8

func (t ServerMessageType) MarshalJSON() ([]byte, error) {
	if s, ok := map[ServerMessageType]string{
		TypeFramebufferUpdate:   `FramebufferUpdate`,
		TypeSetColourMapEntries: `SetColourMapEntries`,
		TypeBell:                `Bell`,
		TypeServerCutText:       `ServerCutText`,
	}[t]; ok {
		return json.Marshal(s)
	} else {
		return json.Marshal(uint8(t))
	}
}

const (
	TypeFramebufferUpdate   ServerMessageType = 0
	TypeSetColourMapEntries ServerMessageType = 1
	TypeBell                ServerMessageType = 2
	TypeServerCutText       ServerMessageType = 3
)

type FramebufferUpdate struct {
	MessageType        ServerMessageType
	Padding            uint8
	NumberOfRectangles uint16
}

type Rectangle struct {
	X        uint16
	Y        uint16
	Width    uint16
	Height   uint16
	Encoding Encoding
}

type ZlibRectangle struct {
	Rectangle
	Length uint32 `tcp:",Data"`
	Data   []byte
}

type CursorRectangle struct {
	Rectangle
	Pixels  []byte
	Bitmask []uint8
}

type XCursorRectangle struct {
	Rectangle
	PrimaryRed     uint8
	PrimaryGreen   uint8
	PrimaryBlue    uint8
	SecondaryRed   uint8
	SecondaryGreen uint8
	SecondaryBlue  uint8
	Bitmap         []uint8
	Bitmask        []uint8
}
