package chio

import "io"

// PacketReader is a function that reads packet data from a reader
type PacketReader func(client BanchoIO, reader io.Reader) (any, error)

// ReaderRegistry holds all packet readers for a specific protocol version
type ReaderRegistry map[uint16]PacketReader

// InheritReaders creates a copy of the parent reader registry
func InheritReaders(parent ReaderRegistry) ReaderRegistry {
	registry := make(ReaderRegistry)
	for k, v := range parent {
		registry[k] = v
	}
	return registry
}
