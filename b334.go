package chio

import (
	"bytes"
	"fmt"
	"io"
)

// b334 adds the "compression" boolean to the
// packet header. It also adds mods to the
// match struct, as well as the packet IDs 50-58
type b334 struct {
	*b323
}

func (client *b334) Clone() BanchoIO {
	previous := b323{}
	return &b334{previous.Clone().(*b323)}
}

func (client *b334) WritePacket(packetId uint16, data []byte) error {
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

func (client *b334) ReadPacket() (packet *BanchoPacket, err error) {
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

func (client *b334) SupportedPackets() []uint16 {
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
		BanchoMatchTransferHost,
		OsuMatchChangeMods,
		OsuMatchLoadComplete,
		BanchoMatchAllPlayersLoaded,
		OsuMatchNoBeatmap,
		OsuMatchNotReady,
		OsuMatchFailed,
		BanchoMatchPlayerFailed,
		BanchoMatchComplete,
	}
	return client.supportedPackets
}

func (client *b334) ImplementsPacket(packetId uint16) bool {
	for _, id := range client.SupportedPackets() {
		if id == packetId {
			return true
		}
	}
	return false
}

func (client *b334) ConvertInputPacketId(packetId uint16) uint16 {
	if packetId == 11 {
		// "IrcJoin" packet
		return BanchoHandleIrcJoin
	}
	if packetId == 51 {
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

func (client *b334) ConvertOutputPacketId(packetId uint16) uint16 {
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
	if packetId > 50 {
		packetId += 1
	}
	return packetId
}

func (client *b334) ReadPacketType(packetId uint16, reader io.Reader) (any, error) {
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

/* New Packets */

func (client *b334) WriteMatchTransferHost() error {
	return client.WritePacket(BanchoMatchTransferHost, []byte{})
}

func (client *b334) WriteMatchAllPlayersLoaded() error {
	return client.WritePacket(BanchoMatchAllPlayersLoaded, []byte{})
}

func (client *b334) WriteMatchComplete() error {
	return client.WritePacket(BanchoMatchComplete, []byte{})
}

func (client *b334) WriteMatchPlayerFailed(slotId uint32) error {
	writer := bytes.NewBuffer([]byte{})
	writeUint32(writer, slotId)
	return client.WritePacket(BanchoMatchPlayerFailed, writer.Bytes())
}

func (client *b334) WriteMatch(writer io.Writer, match Match) error {
	writeUint8(writer, uint8(match.Id))
	writeBoolean(writer, match.InProgress)
	writeUint8(writer, match.Type)
	writeUint16(writer, uint16(match.Mods))
	writeString(writer, match.Name)
	writeString(writer, match.BeatmapText)
	writeInt32(writer, match.BeatmapId)
	writeString(writer, match.BeatmapChecksum)

	for _, slot := range match.Slots {
		writeUint8(writer, slot.Status)
	}

	for _, slot := range match.Slots {
		if slot.HasPlayer() {
			writeInt32(writer, slot.UserId)
		}
	}

	return nil
}

func (client *b334) ReadMatch(reader io.Reader) (*Match, error) {
	var err error
	errors := NewErrorCollection()
	match := &Match{}

	matchId, err := readUint8(reader)
	match.Id = int32(matchId)
	errors.Add(err)
	match.InProgress, err = readBoolean(reader)
	errors.Add(err)
	match.Type, err = readUint8(reader)
	errors.Add(err)
	mods, err := readUint16(reader)
	match.Mods = uint32(mods)
	errors.Add(err)
	match.Name, err = readString(reader)
	errors.Add(err)
	match.BeatmapText, err = readString(reader)
	errors.Add(err)
	match.BeatmapId, err = readInt32(reader)
	errors.Add(err)
	match.BeatmapChecksum, err = readString(reader)
	errors.Add(err)

	match.Slots = make([]*MatchSlot, client.slotSize)

	for i := 0; i < client.slotSize; i++ {
		slot := &MatchSlot{}
		slot.Status, err = readUint8(reader)
		slot.Team = SlotTeamNeutral
		slot.Mods = NoMod
		slot.UserId = -1
		match.Slots[i] = slot
		errors.Add(err)
	}

	for i := 0; i < client.slotSize; i++ {
		if !match.Slots[i].HasPlayer() {
			continue
		}

		match.Slots[i].UserId, err = readInt32(reader)
		if err != nil {
			errors.Add(err)
			continue
		}
	}

	return match, errors.Next()
}

/* Inherited Packets */

// We need to re-define these functions, because the `WritePacket`
// function has been overridden to include compression
// TODO: Refactor this to avoid code duplication

func (client *b334) WriteLoginReply(reply int32) error {
	writer := bytes.NewBuffer([]byte{})
	writeInt32(writer, reply)
	return client.WritePacket(BanchoLoginReply, writer.Bytes())
}

func (client *b334) WriteMessage(message Message) error {
	writer := bytes.NewBuffer([]byte{})
	writeString(writer, message.Sender)
	writeString(writer, message.Content)
	writeString(writer, message.Target)
	return client.WritePacket(BanchoSendMessage, writer.Bytes())
}

func (client *b334) WritePing() error {
	return client.WritePacket(BanchoPing, []byte{})
}

func (client *b334) WriteIrcChangeUsername(oldName string, newName string) error {
	writer := bytes.NewBuffer([]byte{})
	writeString(writer, fmt.Sprintf("%s>>>>%s", oldName, newName))
	return client.WritePacket(BanchoHandleIrcChangeUsername, writer.Bytes())
}

func (client *b334) WriteUserStats(info UserInfo) error {
	writer := bytes.NewBuffer([]byte{})

	if info.Presence.IsIrc {
		writeString(writer, info.Name)
		return client.WritePacket(BanchoHandleIrcJoin, writer.Bytes())
	}

	client.WriteStats(writer, info)
	return client.WritePacket(BanchoHandleOsuUpdate, writer.Bytes())
}

func (client *b334) WriteUserQuit(quit UserQuit) error {
	writer := bytes.NewBuffer([]byte{})

	if quit.Info.Presence.IsIrc && quit.QuitState != QuitStateIrcRemaining {
		writeString(writer, quit.Info.Name)
		return client.WritePacket(BanchoHandleIrcQuit, writer.Bytes())
	}

	if quit.QuitState == QuitStateOsuRemaining {
		return nil
	}

	// Remove from user map
	delete(client.b323.userMap, quit.Info.Id)

	client.WriteStats(writer, *quit.Info)
	return client.WritePacket(BanchoHandleOsuQuit, writer.Bytes())
}

func (client *b334) WriteSpectatorJoined(userId int32) error {
	writer := bytes.NewBuffer([]byte{})
	writeInt32(writer, userId)
	return client.WritePacket(BanchoSpectatorJoined, writer.Bytes())
}

func (client *b334) WriteSpectatorLeft(userId int32) error {
	writer := bytes.NewBuffer([]byte{})
	writeInt32(writer, userId)
	return client.WritePacket(BanchoSpectatorLeft, writer.Bytes())
}

func (client *b334) WriteSpectateFrames(bundle ReplayFrameBundle) error {
	writer := bytes.NewBuffer([]byte{})
	writeUint16(writer, uint16(len(bundle.Frames)))

	for _, frame := range bundle.Frames {
		// Convert button state
		leftMouse := ButtonStateLeft1&frame.ButtonState > 0 || ButtonStateLeft2&frame.ButtonState > 0
		rightMouse := ButtonStateRight1&frame.ButtonState > 0 || ButtonStateRight2&frame.ButtonState > 0

		writeBoolean(writer, leftMouse)
		writeBoolean(writer, rightMouse)
		writeFloat32(writer, frame.MouseX)
		writeFloat32(writer, frame.MouseY)
		writeInt32(writer, frame.Time)
	}

	writeUint8(writer, bundle.Action)

	if bundle.Frame != nil {
		client.WriteScoreFrame(writer, *bundle.Frame)
	}

	return client.WritePacket(BanchoSpectateFrames, writer.Bytes())
}

func (client *b334) WriteVersionUpdate() error {
	return client.WritePacket(BanchoVersionUpdate, []byte{})
}

func (client *b334) WriteSpectatorCantSpectate(userId int32) error {
	writer := bytes.NewBuffer([]byte{})
	writeInt32(writer, userId)
	return client.WritePacket(BanchoSpectatorCantSpectate, writer.Bytes())
}

// Redirect UserPresence packets to UserStats
func (client *b334) WriteUserPresence(info UserInfo) error {
	return client.WriteUserStats(info)
}

func (client *b334) WriteUserPresenceSingle(info UserInfo) error {
	return client.WriteUserPresence(info)
}

func (client *b334) WriteUserPresenceBundle(infos []UserInfo) error {
	for _, info := range infos {
		err := client.WriteUserPresence(info)
		if err != nil {
			return err
		}
	}
	return nil
}

func (client *b334) WriteGetAttention() error {
	return client.WritePacket(BanchoGetAttention, []byte{})
}

func (client *b334) WriteAnnouncement(message string) error {
	writer := bytes.NewBuffer([]byte{})
	writeString(writer, message)
	return client.WritePacket(BanchoAnnounce, writer.Bytes())
}

func (client *b334) WriteMatchUpdate(match Match) error {
	writer := bytes.NewBuffer([]byte{})
	client.WriteMatch(writer, match)
	return client.WritePacket(BanchoMatchUpdate, writer.Bytes())
}

func (client *b334) WriteMatchNew(match Match) error {
	writer := bytes.NewBuffer([]byte{})
	client.WriteMatch(writer, match)
	return client.WritePacket(BanchoMatchNew, writer.Bytes())
}

func (client *b334) WriteMatchDisband(matchId int32) error {
	writer := bytes.NewBuffer([]byte{})
	writeInt32(writer, matchId)
	return client.WritePacket(BanchoMatchDisband, writer.Bytes())
}

func (client *b334) WriteLobbyJoin(userId int32) error {
	writer := bytes.NewBuffer([]byte{})
	writeInt32(writer, userId)
	return client.WritePacket(BanchoLobbyJoin, writer.Bytes())
}

func (client *b334) WriteLobbyPart(userId int32) error {
	writer := bytes.NewBuffer([]byte{})
	writeInt32(writer, userId)
	return client.WritePacket(BanchoLobbyPart, writer.Bytes())
}

func (client *b334) WriteMatchJoinSuccess(match Match) error {
	writer := bytes.NewBuffer([]byte{})
	client.WriteMatch(writer, match)
	return client.WritePacket(BanchoMatchJoinSuccess, writer.Bytes())
}

func (client *b334) WriteMatchJoinFail() error {
	return client.WritePacket(BanchoMatchJoinFail, []byte{})
}

func (client *b334) WriteFellowSpectatorJoined(userId int32) error {
	writer := bytes.NewBuffer([]byte{})
	writeInt32(writer, userId)
	return client.WritePacket(BanchoFellowSpectatorJoined, writer.Bytes())
}

func (client *b334) WriteFellowSpectatorLeft(userId int32) error {
	writer := bytes.NewBuffer([]byte{})
	writeInt32(writer, userId)
	return client.WritePacket(BanchoFellowSpectatorLeft, writer.Bytes())
}

func (client *b334) WriteMatchStart(match Match) error {
	writer := bytes.NewBuffer([]byte{})
	client.WriteMatch(writer, match)
	return client.WritePacket(BanchoMatchStart, writer.Bytes())
}

func (client *b334) WriteMatchScoreUpdate(frame ScoreFrame) error {
	writer := bytes.NewBuffer([]byte{})
	client.WriteScoreFrame(writer, frame)
	return client.WritePacket(BanchoMatchScoreUpdate, writer.Bytes())
}

/* Unsupported Packets */

func (client *b334) WriteMatchSkip() error                               { return nil }
func (client *b334) WriteUnauthorized() error                            { return nil }
func (client *b334) WriteChannelJoinSuccess(channel string) error        { return nil }
func (client *b334) WriteChannelRevoked(channel string) error            { return nil }
func (client *b334) WriteChannelAvailable(channel Channel) error         { return nil }
func (client *b334) WriteChannelAvailableAutojoin(channel Channel) error { return nil }
func (client *b334) WriteBeatmapInfoReply(reply BeatmapInfoReply) error  { return nil }
func (client *b334) WriteLoginPermissions(permissions uint32) error      { return nil }
func (client *b334) WriteFriendsList(userIds []uint32) error             { return nil }
func (client *b334) WriteProtocolNegotiation(version int32) error        { return nil }
func (client *b334) WriteTitleUpdate(update TitleUpdate) error           { return nil }
func (client *b334) WriteMonitor() error                                 { return nil }
func (client *b334) WriteMatchPlayerSkipped(slotId int32) error          { return nil }
func (client *b334) WriteRestart(retryMs int32) error                    { return nil }
func (client *b334) WriteInvite(message Message) error                   { return nil }
func (client *b334) WriteChannelInfoComplete() error                     { return nil }
func (client *b334) WriteMatchChangePassword(password string) error      { return nil }
func (client *b334) WriteSilenceInfo(timeRemaining int32) error          { return nil }
func (client *b334) WriteUserSilenced(userId uint32) error               { return nil }
func (client *b334) WriteUserDMsBlocked(targetName string) error         { return nil }
func (client *b334) WriteTargetIsSilenced(targetName string) error       { return nil }
func (client *b334) WriteVersionUpdateForced() error                     { return nil }
func (client *b334) WriteSwitchServer(target int32) error                { return nil }
func (client *b334) WriteAccountRestricted() error                       { return nil }
func (client *b334) WriteRTX(message string) error                       { return nil }
func (client *b334) WriteMatchAbort() error                              { return nil }
func (client *b334) WriteSwitchTournamentServer(ip string) error         { return nil }
