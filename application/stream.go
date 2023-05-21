package application

import (
	"github.com/0xThiebaut/PCAPeek/transport/tcp"
	"github.com/google/gopacket"
	"github.com/google/gopacket/tcpassembly"
)

func NewRouterStreams(net gopacket.Flow, transport gopacket.Flow, strict bool, applications ...Factory) (client tcp.Stream, server tcp.Stream) {
	// Create a new streams
	client = tcp.NewStream()
	server = tcp.NewStream()
	// Route clients and servers to the first application accepting any of their traffic
	client.Notifier(func() {
		// Route to the first application accepting the client data
		for _, application := range applications {
			if application.Client(net, transport, client) {
				application.Handle(net, transport, client, server)
				// Trigger the client notification
				client.Reassembled([]tcpassembly.Reassembly{})
				return
			}
		}
		// If nothing matches either discard the stream (strict mode) or just clear the buffer (loose) and hope
		// a later packet might match.
		if strict {
			tcp.Discard(client)
		} else {
			client.Clear()
		}
	})
	// Set server routing logic
	server.Notifier(func() {
		// Route to the first application accepting the server data
		for _, application := range applications {
			if application.Server(net.Reverse(), transport.Reverse(), server) {
				application.Handle(net, transport, client, server)
				// Trigger the server notification
				server.Reassembled([]tcpassembly.Reassembly{})
				return
			}
		}

		// If nothing matches either discard the stream (strict mode) or just clear the buffer (loose) and hope
		// a later packet might match.
		if strict {
			tcp.Discard(server)
		} else {
			server.Clear()
		}
	})
	return client, server
}
