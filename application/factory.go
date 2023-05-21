package application

import (
	"github.com/0xThiebaut/PCAPeek/transport/tcp"
	"github.com/google/gopacket"
	"github.com/google/gopacket/tcpassembly"
	"sync"
)

type Factory interface {
	Client(net gopacket.Flow, transport gopacket.Flow, client tcp.PeekStream) bool
	Handle(net gopacket.Flow, transport gopacket.Flow, client tcp.Stream, server tcp.Stream)
	Server(net gopacket.Flow, transport gopacket.Flow, server tcp.PeekStream) bool
}

func NewApplicationStreamFactory(strict bool, applications ...Factory) tcpassembly.StreamFactory {
	return &factory{
		Applications: applications,
		Streams:      map[uint64]map[uint64]tcp.Stream{},
		Strict:       strict,
	}
}

type factory struct {
	Mutex        sync.Mutex
	Streams      map[uint64]map[uint64]tcp.Stream
	Applications []Factory
	Strict       bool
}

func (f *factory) New(net, transport gopacket.Flow) tcpassembly.Stream {
	// Lock the cache for editing
	f.Mutex.Lock()
	defer f.Mutex.Unlock()
	// Ensure the network cache exists
	nh := net.FastHash()
	if _, ok := f.Streams[nh]; !ok {
		f.Streams[nh] = map[uint64]tcp.Stream{}
	}
	// Ensure the transport cache exists
	th := transport.FastHash()
	if cached, ok := f.Streams[nh][th]; !ok {
		client, server := NewRouterStreams(net, transport, f.Strict, f.Applications...)
		f.Streams[nh][th] = server
		return client
	} else {
		delete(f.Streams[nh], th)
		if len(f.Streams[nh]) == 0 {
			delete(f.Streams, nh)
		}
		return cached
	}
}
