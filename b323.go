package chio

import (
	"bytes"
	"fmt"
	"io"
)

// b323 changes the structure of user stats
type b323 struct {
	BanchoIO
	stream           io.ReadWriteCloser
	previous         *b320
	slotSize         int
	supportedPackets []uint16
}

func (client *b323) Write(p []byte) (n int, err error) {
	return client.stream.Write(p)
}

func (client *b323) Read(p []byte) (n int, err error) {
	return client.stream.Read(p)
}

func (client *b323) Close() error {
	return client.stream.Close()
}

func (client *b323) Clone() BanchoIO {
	previous := &b320{}
	clone := previous.Clone()
	return &b323{
		previous: clone.(*b320),
		slotSize: 8,
	}
}

func (client *b323) GetStream() io.ReadWriteCloser {
	return client.stream
}

func (client *b323) SetStream(stream io.ReadWriteCloser) {
	client.stream = stream
	client.previous.SetStream(stream)
}

func (client *b323) WritePacket(packetId uint16, data []byte) error {
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

	_, err = client.Write(writer.Bytes())
	return err
}

func (client *b323) ReadPacket() (packet *BanchoPacket, err error) {
	packet = &BanchoPacket{}
	packet.Id, err = readUint16(client.stream)
	if err != nil {
		return nil, err
	}

	// Convert packet ID to a usable value
	packet.Id = client.ConvertInputPacketId(packet.Id)

	if !client.ImplementsPacket(packet.Id) {
		return nil, nil
	}

	length, err := readInt32(client.stream)
	if err != nil {
		return nil, err
	}

	compressedData := make([]byte, length)
	n, err := client.stream.Read(compressedData)
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
		OsuMatchChangeSettings,
		BanchoFellowSpectatorJoined,
		BanchoFellowSpectatorLeft,
		OsuMatchStart,
		BanchoMatchStart,
		OsuMatchScoreUpdate,
		BanchoMatchScoreUpdate,
		OsuMatchComplete,
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

func (client *b323) OverrideMatchSlotSize(amount int) {
	client.slotSize = amount
	client.previous.OverrideMatchSlotSize(amount)
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
	if packetId > 11 {
		packetId -= 1
	}
	if packetId > 50 {
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
	if packetId >= 11 {
		packetId += 1
	}
	if packetId >= 50 {
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

/* New Packets */

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
		writeString(writer, info.Presence.City)
	}

	client.WriteStatus(writer, info.Status)
	return nil
}

func (client *b323) WriteUserStats(info UserInfo) error {
	writer := bytes.NewBuffer([]byte{})

	if info.Presence.IsIrc {
		writeString(writer, info.Name)
		return client.WritePacket(BanchoHandleIrcJoin, writer.Bytes())
	}

	client.WriteStats(writer, info)
	return client.WritePacket(BanchoHandleOsuUpdate, writer.Bytes())
}

func (client *b323) WriteUserQuit(quit UserQuit) error {
	writer := bytes.NewBuffer([]byte{})

	if quit.Info.Presence.IsIrc && quit.QuitState != QuitStateIrcRemaining {
		writeString(writer, quit.Info.Name)
		return client.WritePacket(BanchoHandleIrcQuit, writer.Bytes())
	}

	if quit.QuitState == QuitStateOsuRemaining {
		return nil
	}

	client.WriteStats(writer, *quit.Info)
	return client.WritePacket(BanchoHandleOsuQuit, writer.Bytes())
}

// Redirect UserPresence packets to UserStats
func (client *b323) WriteUserPresence(info UserInfo) error {
	return client.WriteUserStats(info)
}

func (client *b323) WriteUserPresenceSingle(info UserInfo) error {
	return client.WriteUserPresence(info)
}

func (client *b323) WriteUserPresenceBundle(infos []UserInfo) error {
	for _, info := range infos {
		err := client.WriteUserPresence(info)
		if err != nil {
			return err
		}
	}
	return nil
}

/* Inherited Packets */

func (client *b323) WriteMessage(message Message) error {
	return client.previous.WriteMessage(message)
}

func (client *b323) WriteVersionUpdate() error {
	return client.previous.WriteVersionUpdate()
}

func (client *b323) WriteLoginReply(reply int32) error {
	return client.previous.WriteLoginReply(reply)
}

func (client *b323) WritePing() error {
	return client.previous.WritePing()
}

func (client *b323) WriteIrcChangeUsername(oldName string, newName string) error {
	return client.previous.WriteIrcChangeUsername(oldName, newName)
}

func (client *b323) WriteSpectateFrames(bundle ReplayFrameBundle) error {
	return client.previous.WriteSpectateFrames(bundle)
}

func (client *b323) WriteSpectatorJoined(userId int32) error {
	return client.previous.WriteSpectatorJoined(userId)
}

func (client *b323) WriteSpectatorLeft(userId int32) error {
	return client.previous.WriteSpectatorLeft(userId)
}

func (client *b323) WriteSpectatorCantSpectate(userId int32) error {
	return client.previous.WriteSpectatorCantSpectate(userId)
}

func (client *b323) WriteScoreFrame(writer io.Writer, frame ScoreFrame) error {
	return client.previous.WriteScoreFrame(writer, frame)
}

func (client *b323) WriteStatus(writer io.Writer, status *UserStatus) error {
	return client.previous.WriteStatus(writer, status)
}

func (client *b323) WriteMatch(writer io.Writer, match Match) error {
	return client.previous.WriteMatch(writer, match)
}

func (client *b323) WriteMatchStart(match Match) error {
	return client.previous.WriteMatchStart(match)
}

func (client *b323) WriteMatchScoreUpdate(frame ScoreFrame) error {
	return client.previous.WriteMatchScoreUpdate(frame)
}

func (client *b323) WriteGetAttention() error {
	return client.previous.WriteGetAttention()
}

func (client *b323) WriteAnnouncement(message string) error {
	return client.previous.WriteAnnouncement(message)
}

func (client *b323) WriteMatchUpdate(match Match) error {
	return client.previous.WriteMatchUpdate(match)
}

func (client *b323) WriteMatchNew(match Match) error {
	return client.previous.WriteMatchNew(match)
}

func (client *b323) WriteMatchDisband(matchId int32) error {
	return client.previous.WriteMatchDisband(matchId)
}

func (client *b323) WriteLobbyJoin(userId int32) error {
	return client.previous.WriteLobbyJoin(userId)
}

func (client *b323) WriteLobbyPart(userId int32) error {
	return client.previous.WriteLobbyPart(userId)
}

func (client *b323) WriteMatchJoinSuccess(match Match) error {
	return client.previous.WriteMatchJoinSuccess(match)
}

func (client *b323) WriteMatchJoinFail() error {
	return client.previous.WriteMatchJoinFail()
}

func (client *b323) WriteFellowSpectatorJoined(userId int32) error {
	return client.previous.WriteFellowSpectatorJoined(userId)
}

func (client *b323) WriteFellowSpectatorLeft(userId int32) error {
	return client.previous.WriteFellowSpectatorLeft(userId)
}

func (client *b323) ReadMessage(reader io.Reader) (*Message, error) {
	return client.previous.ReadMessage(reader)
}

func (client *b323) ReadStatus(reader io.Reader) (any, error) {
	return client.previous.ReadStatus(reader)
}

func (client *b323) ReadReplayFrame(reader io.Reader) (*ReplayFrame, error) {
	return client.previous.ReadReplayFrame(reader)
}

func (client b323) ReadScoreFrame(reader io.Reader) (*ScoreFrame, error) {
	return client.previous.ReadScoreFrame(reader)
}

func (client *b323) ReadFrameBundle(reader io.Reader) (*ReplayFrameBundle, error) {
	return client.previous.ReadFrameBundle(reader)
}

func (client *b323) ReadMatch(reader io.Reader) (Match, error) {
	return client.previous.ReadMatch(reader)
}

/* Unsupported Packets */

func (client *b323) WriteMatchTransferHost() error                       { return nil }
func (client *b323) WriteMatchAllPlayersLoaded() error                   { return nil }
func (client *b323) WriteMatchPlayerFailed(slotId uint32) error          { return nil }
func (client *b323) WriteMatchComplete() error                           { return nil }
func (client *b323) WriteMatchSkip() error                               { return nil }
func (client *b323) WriteUnauthorized() error                            { return nil }
func (client *b323) WriteChannelJoinSuccess(channel string) error        { return nil }
func (client *b323) WriteChannelRevoked(channel string) error            { return nil }
func (client *b323) WriteChannelAvailable(channel Channel) error         { return nil }
func (client *b323) WriteChannelAvailableAutojoin(channel Channel) error { return nil }
func (client *b323) WriteBeatmapInfoReply(reply BeatmapInfoReply) error  { return nil }
func (client *b323) WriteLoginPermissions(permissions uint32) error      { return nil }
func (client *b323) WriteFriendsList(userIds []uint32) error             { return nil }
func (client *b323) WriteProtocolNegotiation(version int32) error        { return nil }
func (client *b323) WriteTitleUpdate(update TitleUpdate) error           { return nil }
func (client *b323) WriteMonitor() error                                 { return nil }
func (client *b323) WriteMatchPlayerSkipped(slotId int32) error          { return nil }
func (client *b323) WriteRestart(retryMs int32) error                    { return nil }
func (client *b323) WriteInvite(message Message) error                   { return nil }
func (client *b323) WriteChannelInfoComplete() error                     { return nil }
func (client *b323) WriteMatchChangePassword(password string) error      { return nil }
func (client *b323) WriteSilenceInfo(timeRemaining int32) error          { return nil }
func (client *b323) WriteUserSilenced(userId uint32) error               { return nil }
func (client *b323) WriteUserDMsBlocked(targetName string) error         { return nil }
func (client *b323) WriteTargetIsSilenced(targetName string) error       { return nil }
func (client *b323) WriteVersionUpdateForced() error                     { return nil }
func (client *b323) WriteSwitchServer(target int32) error                { return nil }
func (client *b323) WriteAccountRestricted() error                       { return nil }
func (client *b323) WriteRTX(message string) error                       { return nil }
func (client *b323) WriteMatchAbort() error                              { return nil }
func (client *b323) WriteSwitchTournamentServer(ip string) error         { return nil }
