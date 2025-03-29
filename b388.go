package chio

import (
	"bytes"
	"fmt"
	"io"
)

// b388 changes ranked status from bool->int in beatmap info packets
type b388 struct {
	*b354
}

func (client *b388) WritePacket(stream io.Writer, packetId uint16, data []byte) error {
	// Convert packetId back for the client
	packetId = client.ConvertOutputPacketId(packetId)
	compressionEnabled := len(data) >= 150
	writer := bytes.NewBuffer([]byte{})

	err := writeUint16(writer, packetId)
	if err != nil {
		return err
	}

	err = writeBoolean(writer, compressionEnabled)
	if err != nil {
		return err
	}

	if compressionEnabled {
		data = compressData(data)
	}

	err = writeUint32(writer, uint32(len(data)))
	if err != nil {
		return err
	}

	_, err = writer.Write(data)
	if err != nil {
		return err
	}

	_, err = stream.Write(writer.Bytes())
	return err
}

func (client *b388) ReadPacket(stream io.Reader) (packet *BanchoPacket, err error) {
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

func (client *b388) ReadPacketType(packetId uint16, reader io.Reader) (any, error) {
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
		return readUint32(reader)
	case OsuMatchChangeSettings:
		return client.ReadMatch(reader)
	case OsuMatchChangeSlot:
		return readUint32(reader)
	case OsuMatchLock:
		return readUint32(reader)
	case OsuMatchScoreUpdate:
		return client.ReadScoreFrame(reader)
	case OsuMatchChangeBeatmap:
		return client.ReadMatch(reader)
	case OsuMatchChangeMods:
		return readUint32(reader)
	case OsuChannelJoin:
		return readString(reader)
	case OsuChannelLeave:
		return readString(reader)
	case OsuBeatmapInfoRequest:
		return client.ReadBeatmapInfoRequest(reader)
	default:
		return nil, nil
	}
}

func (client *b388) WriteBeatmapInfoReply(stream io.Writer, reply BeatmapInfoReply) error {
	buffer := bytes.NewBuffer([]byte{})
	writeInt32(buffer, int32(len(reply.Beatmaps)))

	for _, info := range reply.Beatmaps {
		writeInt16(buffer, info.Index)
		writeInt32(buffer, info.BeatmapId)
		writeInt32(buffer, info.BeatmapSetId)
		writeInt32(buffer, info.ThreadId)
		writeInt8(buffer, info.RankedStatus)
		writeInt8(buffer, info.OsuRank)
		writeString(buffer, info.Checksum)
	}

	return nil
}
