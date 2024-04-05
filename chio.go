package chio

import (
	"bytes"
	"fmt"
	"io"
)

type Encoder struct {
	packetId    uint16
	function    func(any, io.Writer)
	fromVersion uint32
	toVersion   uint32
}

var Encoders map[uint16][]Encoder = make(map[uint16][]Encoder)

// RegisterEncoder registers a new encoder for the given packetId and version range.
// The encoder function must encode the given data type to the given io.Writer.
func RegisterEncoder(packetId uint16, fromVersion uint32, toVersion uint32, function func(any, io.Writer)) {
	Encoders[packetId] = append(
		Encoders[packetId],
		Encoder{
			packetId,
			function,
			fromVersion,
			toVersion,
		},
	)
}

// Encode encodes a bancho packet with the given packetId and packetData.
// It adjusts the output to the given version, so that the packet is compatible with all clients.
// The packet data must be a struct from chio/types.go, that matches the given packetId.
func Encode[T any](packetId uint16, packetData T, version uint32) ([]byte, error) {
	if version > 20130815 {
		version = 20130815
	}

	encoder := findEncoder(packetId, version)
	if encoder == nil {
		return nil, fmt.Errorf("no encoder found for packetId '%d' and version '%d'", packetId, version)
	}

	// Encode data type
	dataBuffer := new(bytes.Buffer)
	encoder.function(packetData, dataBuffer)

	// Encode packet
	packetBuffer := new(bytes.Buffer)
	compression := true
	writeUint16(packetId, packetBuffer)

	if version > 323 {
		// Version 323 introduced a boolean for compression
		compression = false
		writeBool(compression, packetBuffer)
	}

	if compression {
		// Compress data with gzip
		compressBuffer(*dataBuffer)
	}

	data := dataBuffer.Bytes()
	length := len(data)

	writeUint32(uint32(length), packetBuffer)
	packetBuffer.Write(data)
	return packetBuffer.Bytes(), nil
}

// Find and return the encoder that fits the given packetId and version.
func findEncoder(packetId uint16, version uint32) *Encoder {
	if encoders, ok := Encoders[packetId]; ok {
		for _, encoder := range encoders {
			if encoder.fromVersion >= version && encoder.toVersion <= version {
				return &encoder
			}
		}
	}
	return nil
}
