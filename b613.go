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

func (client *b613) ReadPacketType(packetId uint16, reader io.Reader) (any, error) {
	switch packetId {
	case OsuSendUserStatus:
		return client.ReadStatus(reader)
	case OsuSendIrcMessage:
		return client.ReadMessage(reader)
	case OsuSendIrcMessagePrivate:
		return client.ReadMessage(reader)
	case OsuStartSpectating:
		return readUint32(reader)
	case OsuSpectateFrames:
		return client.ReadFrameBundle(reader)
	case OsuErrorReport:
		return readString(reader)
	case OsuMatchCreate:
		return client.ReadMatch(reader)
	case OsuMatchJoin:
		return client.ReadMatchJoin(reader)
	case OsuMatchChangeSettings:
		return client.ReadMatch(reader)
	case OsuMatchChangeSlot:
		return readUint32(reader)
	case OsuMatchLock:
		return readUint32(reader)
	case OsuMatchScoreUpdate:
		return client.ReadScoreFrame(reader)
	case OsuMatchChangeMods:
		return readUint32(reader)
	case OsuChannelJoin:
		return readString(reader)
	case OsuChannelLeave:
		return readString(reader)
	case OsuBeatmapInfoRequest:
		return client.ReadBeatmapInfoRequest(reader)
	case OsuMatchTransferHost:
		return readInt32(reader)
	case OsuFriendsAdd:
		return readInt32(reader)
	case OsuFriendsRemove:
		return readInt32(reader)
	default:
		return nil, nil
	}
}
