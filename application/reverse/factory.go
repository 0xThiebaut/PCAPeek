package reverse

import (
	"github.com/0xThiebaut/PCAPeek/application"
	"github.com/0xThiebaut/PCAPeek/transport/tcp"
	"github.com/google/gopacket"
)

func New(application application.Factory) application.Factory {
	return &factory{application: application}
}

type factory struct {
	application application.Factory
}

func (f *factory) Client(net gopacket.Flow, transport gopacket.Flow, client tcp.PeekStream) bool {
	return f.application.Server(net.Reverse(), transport.Reverse(), client)
}

func (f *factory) Handle(net gopacket.Flow, transport gopacket.Flow, client tcp.Stream, server tcp.Stream) {
	f.application.Handle(net.Reverse(), transport.Reverse(), server, client)
}

func (f *factory) Server(net gopacket.Flow, transport gopacket.Flow, server tcp.PeekStream) bool {
	return f.application.Client(net.Reverse(), transport.Reverse(), server)
}
