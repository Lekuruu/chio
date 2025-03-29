package chio

import (
	"bytes"
	"fmt"
	"io"
)

// b323 changes the structure of user stats
// and adds the "MatchChangeBeatmap" packet
type b323 struct {
	*b320
}

func (client *b323) WritePacket(stream io.Writer, packetId uint16, data []byte) error {
	// Convert packetId back for the client
	packetId = client.ConvertOutputPacketId(packetId)
	writer := bytes.NewBuffer([]byte{})

	err := writeUint16(writer, packetId)
	if err != nil {
		return err
	}

	compressed := compressData(data)
	err = writeUint32(writer, uint32(len(compressed)))
	if err != nil {
		return err
	}

	_, err = writer.Write(compressed)
	if err != nil {
		return err
	}

	_, err = stream.Write(writer.Bytes())
	return err
}

func (client *b323) ReadPacket(stream io.Reader) (packet *BanchoPacket, err error) {
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

	length, err := readInt32(stream)
	if err != nil {
		return nil, err
	}

	compressedData := make([]byte, length)
	n, err := stream.Read(compressedData)
	if err != nil {
		return nil, err
	}

	if n != int(length) {
		return nil, fmt.Errorf("expected %d bytes, got %d", length, n)
	}

	data, err := decompressData(compressedData)
	if err != nil {
		return nil, err
	}

	packet.Data, err = client.ReadPacketType(packet.Id, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	return packet, nil
}

func (client *b323) SupportedPackets() []uint16 {
	if client.supportedPackets != nil {
		return client.supportedPackets
	}

	client.supportedPackets = []uint16{
		OsuSendUserStatus,
		OsuSendIrcMessage,
		OsuExit,
		OsuRequestStatusUpdate,
		OsuPong,
		BanchoLoginReply,
		BanchoCommandError,
		BanchoSendMessage,
		BanchoPing,
		BanchoHandleIrcChangeUsername,
		BanchoHandleIrcQuit,
		BanchoHandleOsuUpdate,
		BanchoHandleOsuQuit,
		BanchoSpectatorJoined,
		BanchoSpectatorLeft,
		BanchoSpectateFrames,
		OsuStartSpectating,
		OsuStopSpectating,
		OsuSpectateFrames,
		BanchoVersionUpdate,
		OsuErrorReport,
		OsuCantSpectate,
		BanchoSpectatorCantSpectate,
		BanchoGetAttention,
		BanchoAnnounce,
		OsuSendIrcMessagePrivate,
		BanchoMatchUpdate,
		BanchoMatchNew,
		BanchoMatchDisband,
		OsuLobbyPart,
		OsuLobbyJoin,
		OsuMatchCreate,
		OsuMatchJoin,
		OsuMatchPart,
		BanchoLobbyJoin,
		BanchoLobbyPart,
		BanchoMatchJoinSuccess,
		BanchoMatchJoinFail,
		OsuMatchChangeSlot,
		OsuMatchReady,
		OsuMatchLock,
		BanchoFellowSpectatorJoined,
		BanchoFellowSpectatorLeft,
		OsuMatchStart,
		BanchoMatchStart,
		OsuMatchScoreUpdate,
		BanchoMatchScoreUpdate,
		OsuMatchComplete,
		OsuMatchChangeSettings,
		OsuMatchChangeBeatmap,
	}
	return client.supportedPackets
}

func (client *b323) ImplementsPacket(packetId uint16) bool {
	for _, id := range client.SupportedPackets() {
		if id == packetId {
			return true
		}
	}
	return false
}

func (client *b323) ConvertInputPacketId(packetId uint16) uint16 {
	if packetId == 11 {
		// "IrcJoin" packet
		return BanchoHandleIrcJoin
	}
	if packetId == 50 {
		// "MatchChangeBeatmap" packet
		return OsuMatchChangeBeatmap
	}
	if packetId > 11 && packetId <= 45 {
		packetId -= 1
	}
	return packetId
}

func (client *b323) ConvertOutputPacketId(packetId uint16) uint16 {
	if packetId == BanchoHandleIrcJoin {
		// "IrcJoin" packet
		return 11
	}
	if packetId == OsuMatchChangeBeatmap {
		// "MatchChangeBeatmap" packet
		return 50
	}
	if packetId >= 11 && packetId < 45 {
		packetId += 1
	}
	return packetId
}

func (client *b323) ReadPacketType(packetId uint16, reader io.Reader) (any, error) {
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
	default:
		return nil, nil
	}
}

func (client *b323) WriteStats(writer io.Writer, info UserInfo) error {
	writeInt32(writer, info.Id)
	writeBoolean(writer, info.Status.UpdateStats)

	if info.Status.UpdateStats {
		writeString(writer, info.Name)
		writeUint64(writer, info.Stats.Rscore)
		writeFloat32(writer, float32(info.Stats.Accuracy))
		writeInt32(writer, info.Stats.Playcount)
		writeUint64(writer, info.Stats.Tscore)
		writeInt32(writer, info.Stats.Rank)
		writeString(writer, info.AvatarFilename())
		writeUint8(writer, uint8(info.Presence.Timezone+24))
		writeString(writer, info.Presence.Location())
	}

	client.WriteStatus(writer, info.Status)
	return nil
}

func (client *b323) WriteUserStats(stream io.Writer, info UserInfo) error {
	writer := bytes.NewBuffer([]byte{})

	if info.Presence.IsIrc {
		writeString(writer, info.Name)
		return client.WritePacket(stream, BanchoHandleIrcJoin, writer.Bytes())
	}

	// We assume that the client has not seen this user before, so
	// we send two packets: one for the user stats, and one for the "presence".
	info.Status.UpdateStats = true
	client.WriteStats(writer, info)

	if err := client.WritePacket(stream, BanchoHandleOsuUpdate, writer.Bytes()); err != nil {
		return err
	}

	writer.Reset()
	info.Status.UpdateStats = false
	client.WriteStats(writer, info)
	return client.WritePacket(stream, BanchoHandleOsuUpdate, writer.Bytes())
}

func (client *b323) WriteUserQuit(stream io.Writer, quit UserQuit) error {
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

func (client *b323) WriteUserPresence(stream io.Writer, info UserInfo) error {
	return client.WriteUserStats(stream, info)
}

func (client *b323) WriteUserPresenceSingle(stream io.Writer, info UserInfo) error {
	return client.WriteUserPresence(stream, info)
}

func (client *b323) WriteUserPresenceBundle(stream io.Writer, infos []UserInfo) error {
	for _, info := range infos {
		err := client.WriteUserPresence(stream, info)
		if err != nil {
			return err
		}
	}
	return nil
}

func (client *b323) WriteMatchScoreUpdate(stream io.Writer, frame ScoreFrame) error {
	writer := bytes.NewBuffer([]byte{})
	client.WriteScoreFrame(writer, frame)
	return client.WritePacket(stream, BanchoMatchScoreUpdate, writer.Bytes())
}
