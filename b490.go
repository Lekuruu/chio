package chio

import (
	"bytes"
	"fmt"
	"io"
)

// b490 now sends the beatmap ID in user status updates and
// can request a list of set IDs inside of the beatmap info request.
type b490 struct {
	*b489
}

func (client *b490) ReadPacket(stream io.Reader) (packet *BanchoPacket, err error) {
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

func (client *b490) ReadPacketType(packetId uint16, reader io.Reader) (any, error) {
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

func (client *b490) ReadStatus(reader io.Reader) (*UserStatus, error) {
	var err error
	errors := NewErrorCollection()
	status := &UserStatus{}
	status.Action, err = readUint8(reader)
	errors.Add(err)
	beatmapUpdate, err := readBoolean(reader)
	errors.Add(err)

	if beatmapUpdate {
		status.Text, err = readString(reader)
		errors.Add(err)
		status.BeatmapChecksum, err = readString(reader)
		errors.Add(err)
		mods, err := readUint16(reader)
		errors.Add(err)
		status.Mods = uint32(mods)

		if client.ProtocolVersion() >= 1 {
			status.Mode, err = readUint8(reader)
			errors.Add(err)
			status.BeatmapId, err = readInt32(reader)
			errors.Add(err)
		}
	}

	return status, errors.Next()
}

func (client *b490) WriteStatus(writer io.Writer, status *UserStatus) error {
	// Convert action enum
	action := status.Action

	if status.UpdateStats {
		// This will make the client update the user's stats
		// It will not be present in later versions
		action = StatusStatsUpdate
	}

	beatmapUpdate := true
	writeUint8(writer, action)
	writeBoolean(writer, beatmapUpdate)

	if beatmapUpdate {
		writeString(writer, status.Text)
		writeString(writer, status.BeatmapChecksum)
		writeUint16(writer, uint16(status.Mods))

		if client.ProtocolVersion() >= 1 {
			writeUint8(writer, status.Mode)
			writeInt32(writer, status.BeatmapId)
		}
	}

	return nil
}

func (client *b490) ReadBeatmapInfoRequest(reader io.Reader) (*BeatmapInfoRequest, error) {
	var err error
	errors := NewErrorCollection()
	request := &BeatmapInfoRequest{}

	filenameCount, err := readInt32(reader)
	errors.Add(err)

	for range filenameCount {
		filename, err := readString(reader)
		errors.Add(err)
		request.Filenames = append(request.Filenames, filename)
	}

	idCount, err := readInt32(reader)
	errors.Add(err)

	for range idCount {
		id, err := readInt32(reader)
		errors.Add(err)
		request.Ids = append(request.Ids, id)
	}

	return request, errors.Next()
}

func (client *b490) WriteUserPresence(stream io.Writer, info UserInfo) error {
	writer := bytes.NewBuffer([]byte{})

	if info.Presence.IsIrc {
		writeString(writer, info.Name)
		return client.WritePacket(stream, BanchoHandleIrcJoin, writer.Bytes())
	}

	writeInt32(writer, info.Id)
	writeUint8(writer, CompletenessFull)
	client.WriteStatus(writer, info.Status)
	writeUint64(writer, info.Stats.Rscore)
	writeFloat32(writer, float32(info.Stats.Accuracy))
	writeInt32(writer, info.Stats.Playcount)
	writeUint64(writer, info.Stats.Tscore)
	writeUint16(writer, uint16(info.Stats.Rank))
	writeString(writer, info.Name)
	writeString(writer, info.AvatarFilename())
	writeUint8(writer, uint8(info.Presence.Timezone+24))
	writeString(writer, info.Presence.Location())
	writeUint8(writer, info.Presence.Permissions)
	return client.WritePacket(stream, BanchoHandleOsuUpdate, writer.Bytes())
}

func (client *b490) WriteStats(writer io.Writer, info UserInfo) error {
	writeInt32(writer, info.Id)
	writeUint8(writer, CompletenessStatistics)
	client.WriteStatus(writer, info.Status)
	writeUint64(writer, info.Stats.Rscore)
	writeFloat32(writer, float32(info.Stats.Accuracy))
	writeInt32(writer, info.Stats.Playcount)
	writeUint64(writer, info.Stats.Tscore)
	writeUint16(writer, uint16(info.Stats.Rank))
	return nil
}

func (client *b490) WriteUserStats(stream io.Writer, info UserInfo) error {
	writer := bytes.NewBuffer([]byte{})

	if info.Presence.IsIrc {
		writeString(writer, info.Name)
		return client.WritePacket(stream, BanchoHandleIrcJoin, writer.Bytes())
	}

	client.WriteStats(writer, info)
	return client.WritePacket(stream, BanchoHandleOsuUpdate, writer.Bytes())
}

func (client *b490) WriteUserQuit(stream io.Writer, quit UserQuit) error {
	writer := bytes.NewBuffer([]byte{})

	if quit.Info.Presence.IsIrc && quit.QuitState != QuitStateIrcRemaining {
		writeString(writer, quit.Info.Name)
		return client.WritePacket(stream, BanchoHandleIrcQuit, writer.Bytes())
	}

	if quit.QuitState == QuitStateOsuRemaining {
		return nil
	}

	client.WriteStats(writer, *quit.Info)
	return client.WritePacket(stream, BanchoHandleOsuQuit, writer.Bytes())
}
