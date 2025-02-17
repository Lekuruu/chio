package chio

import (
	"bytes"
	"fmt"
	"io"
)

// b338 changes the structure of statuses
type b338 struct {
	*b334
}

func (client *b338) Clone() BanchoIO {
	previous := b334{}
	return &b338{previous.Clone().(*b334)}
}

func (client *b338) WritePacket(packetId uint16, data []byte) error {
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

	_, err = client.Write(writer.Bytes())
	return err
}

func (client *b338) ReadPacket() (packet *BanchoPacket, err error) {
	packet = &BanchoPacket{}
	packet.Id, err = readUint16(client.stream)
	if err != nil {
		return nil, err
	}

	// Convert packet ID to a usable value
	packet.Id = client.ConvertInputPacketId(packet.Id)

	if !client.ImplementsPacket(packet.Id) {
		return nil, fmt.Errorf("packet '%d' not implemented", packet.Id)
	}

	compressionEnabled, err := readBoolean(client.stream)
	if err != nil {
		return nil, err
	}

	length, err := readInt32(client.stream)
	if err != nil {
		return nil, err
	}

	data := make([]byte, length)
	n, err := client.stream.Read(data)
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

func (client *b338) ReadPacketType(packetId uint16, reader io.Reader) (any, error) {
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
	default:
		return nil, nil
	}
}

func (client *b338) ReadStatus(reader io.Reader) (*UserStatus, error) {
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
	}

	return status, errors.Next()
}

func (client *b338) WriteStatus(writer io.Writer, status *UserStatus) error {
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
	}

	return nil
}

func (client *b338) WriteStats(writer io.Writer, info UserInfo) error {
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

func (client *b338) WriteUserStats(info UserInfo) error {
	writer := bytes.NewBuffer([]byte{})

	if info.Presence.IsIrc {
		writeString(writer, info.Name)
		return client.WritePacket(BanchoHandleIrcJoin, writer.Bytes())
	}

	client.WriteStats(writer, info)
	return client.WritePacket(BanchoHandleOsuUpdate, writer.Bytes())
}

func (client *b338) WriteUserQuit(quit UserQuit) error {
	writer := bytes.NewBuffer([]byte{})

	if quit.Info.Presence.IsIrc && quit.QuitState != QuitStateIrcRemaining {
		writeString(writer, quit.Info.Name)
		return client.WritePacket(BanchoHandleIrcQuit, writer.Bytes())
	}

	if quit.QuitState == QuitStateOsuRemaining {
		return nil
	}

	// Remove from user map
	delete(client.userMap, quit.Info.Id)

	client.WriteStats(writer, *quit.Info)
	return client.WritePacket(BanchoHandleOsuQuit, writer.Bytes())
}

// Use CompletenessFull to achieve the same effect as presence
func (client *b338) WriteUserPresence(info UserInfo) error {
	writer := bytes.NewBuffer([]byte{})

	if info.Presence.IsIrc {
		writeString(writer, info.Name)
		return client.WritePacket(BanchoHandleIrcJoin, writer.Bytes())
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
	writeString(writer, info.Presence.City)
	return client.WritePacket(BanchoHandleOsuUpdate, writer.Bytes())
}

func (client *b338) WriteUserPresenceSingle(info UserInfo) error {
	return client.WriteUserPresence(info)
}

func (client *b338) WriteUserPresenceBundle(infos []UserInfo) error {
	for _, info := range infos {
		err := client.WriteUserPresence(info)
		if err != nil {
			return err
		}
	}
	return nil
}
