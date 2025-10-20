package chio

import (
	"bytes"
	"fmt"
	"io"
)

// b613 allows the client to leave channels.
type b613 struct {
	*b591
}

func (client *b613) ReadPacket(stream io.Reader) (packet *BanchoPacket, err error) {
	packet = &BanchoPacket{}
	packet.Id, err = readUint16(stream)
	if err != nil {
		return nil, err
	}

	// Convert packet ID to a usable value
	packet.Id = client.ConvertInputPacketId(packet.Id)

	if !client.ImplementsPacket(packet.Id) {
		return nil, fmt.Errorf("packet '%d' not implemented", packet.Id)
	}

	compressionEnabled, err := readBoolean(stream)
	if err != nil {
		return nil, err
	}

	length, err := readInt32(stream)
	if err != nil {
		return nil, err
	}

	data := make([]byte, length)
	n, err := stream.Read(data)
	if err != nil {
		return nil, err
	}

	if n != int(length) {
		return nil, fmt.Errorf("expected %d bytes, got %d", length, n)
	}

	if compressionEnabled {
		data, err = decompressData(data)
		if err != nil {
			return nil, err
		}
	}

	packet.Data, err = client.ReadPacketType(packet.Id, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	return packet, nil
}

func (client *b613) ReadChannelLeave(reader io.Reader) (string, error) {
	return readString(reader)
}
