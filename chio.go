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

type Decoder struct {
	packetId    uint16
	function    func(io.Reader) (any, error)
	fromVersion uint32
	toVersion   uint32
}

var Decoders map[uint16][]Decoder = make(map[uint16][]Decoder)
var Encoders map[uint16][]Encoder = make(map[uint16][]Encoder)

// RegisterDecoder registers a new decoder for the given packetId and version range.
// The decoder function must decode the given io.Reader data to a matching data type.
func RegisterDecoder(packetId uint16, fromVersion uint32, toVersion uint32, function func(io.Reader) (any, error)) {
	Decoders[packetId] = append(
		Decoders[packetId],
		Decoder{
			packetId,
			function,
			fromVersion,
			toVersion,
		},
	)
}

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

// Decode decodes a bancho packet from the given buffer.
// It adjusts the output to the given version, so that the packet is compatible with all clients.
// The packet data is returned as a struct from chio/types.go, that matches the given packetId.
func Decode(buffer *bytes.Buffer, version uint32) (uint16, any, error) {
	if version > 20130815 {
		version = 20130815
	}

	packetId := readUint16(buffer)
	compression := true

	if version > 323 {
		// Version 323 introduced a boolean for compression
		compression = readBool(buffer)
	}

	length := readUint32(buffer)
	data := buffer.Next(int(length))

	if compression {
		// Decompress data with gzip
		data = decompressData(data)
	}

	decoder := findDecoder(packetId, version)

	if decoder == nil {
		return 0, nil, fmt.Errorf("no decoder found for packetId '%d' and version '%d'", packetId, version)
	}

	reader := io.Reader(bytes.NewReader(data))
	packetType, err := decoder.function(reader)

	if err != nil {
		return 0, nil, err
	}

	return packetId, packetType, nil
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
		compressBuffer(dataBuffer)
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
			if encoder.fromVersion <= version && encoder.toVersion <= version {
				return &encoder
			}
		}
	}
	return nil
}

// Find and return the decoder that fits the given packetId and version.
func findDecoder(packetId uint16, version uint32) *Decoder {
	if decoders, ok := Decoders[packetId]; ok {
		for _, decoder := range decoders {
			if decoder.fromVersion <= version && decoder.toVersion <= version {
				return &decoder
			}
		}
	}
	return nil
}
