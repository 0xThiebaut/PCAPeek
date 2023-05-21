package rfb

import (
	"unicode/utf8"

	"github.com/0xThiebaut/PCAPeek/application"
	"github.com/0xThiebaut/PCAPeek/output/files"
	"github.com/0xThiebaut/PCAPeek/output/media"
	"github.com/0xThiebaut/PCAPeek/transport/tcp"
	"github.com/google/gopacket"
)

func New(mo media.Factory, fo files.Factory) application.Factory {
	return &factory{
		Media: mo,
		Files: fo,
	}
}

type factory struct {
	id    uint64
	Media media.Factory
	Files files.Factory
}

func (f *factory) Client(net gopacket.Flow, transport gopacket.Flow, client tcp.PeekStream) bool {
	// By default, we don't recognize ClientInit messages.
	// The protocol only defines these as 1 byte messages which is too small for confident decisions.
	// TODO: In the future we may cache these to not lose the byte's meaning.
	return false
}

func (f *factory) Server(net gopacket.Flow, transport gopacket.Flow, server tcp.PeekStream) bool {
	init := ServerInit{}
	_, err := tcp.UnmarshallPeek(server, &init)
	// TODO: Currently mark the stream as supported if the message seems like a ServerInit message
	// where the frame buffer has a landscape orientation and the name is a valid UTF-8 string.
	// Both of these assertions are NOT required per the protocol and only match a subset of possible values.
	return err == nil && init.FramebufferWidth > init.FramebufferHeight && utf8.ValidString(init.Name)
}

func (f *factory) Handle(net gopacket.Flow, transport gopacket.Flow, client tcp.Stream, server tcp.Stream) {
	s := stream{
		id:        f.id,
		net:       net,
		transport: transport,
		client:    client,
		server:    server,
		Media:     f.Media,
		Files:     f.Files,
	}
	f.id++
	// TODO: In the future we can peek the client and server streams to deduce the connection's stage.
	// Currently, we assume IcedID's state which omits handshakes.
	client.Notifier(s.ClientInit)
	server.Notifier(s.ServerInit)
}
