package chio

import (
	"bytes"
	"fmt"
	"io"
)

// b334 has a lot of breaking changes:
//   - Compression boolean inside packet header
//   - Removal of checksums in score frames
//   - Mods inside match struct
//   - Packet IDs 50-58
type b334 struct {
	*b323
}

func (client *b334) WritePacket(stream io.Writer, packetId uint16, data []byte) error {
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

func (client *b334) ReadPacket(stream io.Reader) (packet *BanchoPacket, err error) {
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
		return 51
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

func (client *b334) WriteMatchTransferHost(stream io.Writer) error {
	return client.WritePacket(stream, BanchoMatchTransferHost, []byte{})
}

func (client *b334) WriteMatchAllPlayersLoaded(stream io.Writer) error {
	return client.WritePacket(stream, BanchoMatchAllPlayersLoaded, []byte{})
}

func (client *b334) WriteMatchComplete(stream io.Writer) error {
	return client.WritePacket(stream, BanchoMatchComplete, []byte{})
}

func (client *b334) WriteMatchPlayerFailed(stream io.Writer, slotId uint32) error {
	writer := bytes.NewBuffer([]byte{})
	writeUint32(writer, slotId)
	return client.WritePacket(stream, BanchoMatchPlayerFailed, writer.Bytes())
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

func (client *b334) WriteScoreFrame(writer io.Writer, frame ScoreFrame) error {
	if frame.Hp == 0 {
		// Used by old clients to determine
		// if the player is passing
		frame.Hp = 254
	}

	writeInt32(writer, frame.Time)
	writeUint8(writer, frame.Id)
	writeUint16(writer, frame.Total300)
	writeUint16(writer, frame.Total100)
	writeUint16(writer, frame.Total50)
	writeUint16(writer, frame.TotalGeki)
	writeUint16(writer, frame.TotalKatu)
	writeUint16(writer, frame.TotalMiss)
	writeUint32(writer, frame.TotalScore)
	writeUint16(writer, frame.MaxCombo)
	writeUint16(writer, frame.CurrentCombo)
	writeBoolean(writer, frame.Perfect)
	writeUint8(writer, frame.Hp)
	return nil
}

func (client *b334) ReadFrameBundle(reader io.Reader) (*ReplayFrameBundle, error) {
	count, err := readUint16(reader)
	if err != nil {
		return nil, err
	}

	frames := make([]*ReplayFrame, count)
	for i := 0; i < int(count); i++ {
		frame, err := client.ReadReplayFrame(reader)
		if err != nil {
			return nil, err
		}
		frames[i] = frame
	}

	action, err := readUint8(reader)
	if err != nil {
		return nil, err
	}

	frame, err := client.ReadScoreFrame(reader)
	if err != nil && err.Error() != "EOF" {
		return nil, err
	}

	return &ReplayFrameBundle{Frames: frames, Action: action, Frame: frame}, nil
}

func (client *b334) ReadScoreFrame(reader io.Reader) (*ScoreFrame, error) {
	var err error
	errors := NewErrorCollection()
	frame := &ScoreFrame{}
	frame.Time, err = readInt32(reader)
	errors.Add(err)
	frame.Id, err = readUint8(reader)
	errors.Add(err)
	frame.Total300, err = readUint16(reader)
	errors.Add(err)
	frame.Total100, err = readUint16(reader)
	errors.Add(err)
	frame.Total50, err = readUint16(reader)
	errors.Add(err)
	frame.TotalGeki, err = readUint16(reader)
	errors.Add(err)
	frame.TotalKatu, err = readUint16(reader)
	errors.Add(err)
	frame.TotalMiss, err = readUint16(reader)
	errors.Add(err)
	frame.TotalScore, err = readUint32(reader)
	errors.Add(err)
	frame.MaxCombo, err = readUint16(reader)
	errors.Add(err)
	frame.CurrentCombo, err = readUint16(reader)
	errors.Add(err)
	frame.Perfect, err = readBoolean(reader)
	errors.Add(err)
	frame.Hp, err = readUint8(reader)
	errors.Add(err)

	if frame.Hp == 254 {
		// Used by old clients to determine
		// if the player is passing
		frame.Hp = 0
	}

	return frame, errors.Next()
}

/* Inherited Packets */

// We need to re-define these functions, because the `WritePacket`
// function has been overridden to include compression
// TODO: Refactor this to avoid code duplication

func (client *b334) WriteLoginReply(stream io.Writer, reply int32) error {
	writer := bytes.NewBuffer([]byte{})
	writeInt32(writer, reply)
	return client.WritePacket(stream, BanchoLoginReply, writer.Bytes())
}

func (client *b334) WriteMessage(stream io.Writer, message Message) error {
	writer := bytes.NewBuffer([]byte{})
	writeString(writer, message.Sender)
	writeString(writer, message.Content)
	writeString(writer, message.Target)
	return client.WritePacket(stream, BanchoSendMessage, writer.Bytes())
}

func (client *b334) WritePing(stream io.Writer) error {
	return client.WritePacket(stream, BanchoPing, []byte{})
}

func (client *b334) WriteIrcChangeUsername(stream io.Writer, oldName string, newName string) error {
	writer := bytes.NewBuffer([]byte{})
	writeString(writer, fmt.Sprintf("%s>>>>%s", oldName, newName))
	return client.WritePacket(stream, BanchoHandleIrcChangeUsername, writer.Bytes())
}

func (client *b334) WriteUserStats(stream io.Writer, info UserInfo) error {
	writer := bytes.NewBuffer([]byte{})

	if info.Presence.IsIrc {
		writeString(writer, info.Name)
		return client.WritePacket(stream, BanchoHandleIrcJoin, writer.Bytes())
	}

	client.WriteStats(writer, info)
	return client.WritePacket(stream, BanchoHandleOsuUpdate, writer.Bytes())
}

func (client *b334) WriteUserQuit(stream io.Writer, quit UserQuit) error {
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

func (client *b334) WriteSpectatorJoined(stream io.Writer, userId int32) error {
	writer := bytes.NewBuffer([]byte{})
	writeInt32(writer, userId)
	return client.WritePacket(stream, BanchoSpectatorJoined, writer.Bytes())
}

func (client *b334) WriteSpectatorLeft(stream io.Writer, userId int32) error {
	writer := bytes.NewBuffer([]byte{})
	writeInt32(writer, userId)
	return client.WritePacket(stream, BanchoSpectatorLeft, writer.Bytes())
}

func (client *b334) WriteSpectateFrames(stream io.Writer, bundle ReplayFrameBundle) error {
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

	return client.WritePacket(stream, BanchoSpectateFrames, writer.Bytes())
}

func (client *b334) WriteVersionUpdate(stream io.Writer) error {
	return client.WritePacket(stream, BanchoVersionUpdate, []byte{})
}

func (client *b334) WriteSpectatorCantSpectate(stream io.Writer, userId int32) error {
	writer := bytes.NewBuffer([]byte{})
	writeInt32(writer, userId)
	return client.WritePacket(stream, BanchoSpectatorCantSpectate, writer.Bytes())
}

// Redirect UserPresence packets to UserStats
func (client *b334) WriteUserPresence(stream io.Writer, info UserInfo) error {
	return client.WriteUserStats(stream, info)
}

func (client *b334) WriteUserPresenceSingle(stream io.Writer, info UserInfo) error {
	return client.WriteUserPresence(stream, info)
}

func (client *b334) WriteUserPresenceBundle(stream io.Writer, infos []UserInfo) error {
	for _, info := range infos {
		err := client.WriteUserPresence(stream, info)
		if err != nil {
			return err
		}
	}
	return nil
}

func (client *b334) WriteGetAttention(stream io.Writer) error {
	return client.WritePacket(stream, BanchoGetAttention, []byte{})
}

func (client *b334) WriteAnnouncement(stream io.Writer, message string) error {
	writer := bytes.NewBuffer([]byte{})
	writeString(writer, message)
	return client.WritePacket(stream, BanchoAnnounce, writer.Bytes())
}

func (client *b334) WriteMatchUpdate(stream io.Writer, match Match) error {
	writer := bytes.NewBuffer([]byte{})
	client.WriteMatch(writer, match)
	return client.WritePacket(stream, BanchoMatchUpdate, writer.Bytes())
}

func (client *b334) WriteMatchNew(stream io.Writer, match Match) error {
	writer := bytes.NewBuffer([]byte{})
	client.WriteMatch(writer, match)
	return client.WritePacket(stream, BanchoMatchNew, writer.Bytes())
}

func (client *b334) WriteMatchDisband(stream io.Writer, matchId int32) error {
	writer := bytes.NewBuffer([]byte{})
	writeInt32(writer, matchId)
	return client.WritePacket(stream, BanchoMatchDisband, writer.Bytes())
}

func (client *b334) WriteLobbyJoin(stream io.Writer, userId int32) error {
	writer := bytes.NewBuffer([]byte{})
	writeInt32(writer, userId)
	return client.WritePacket(stream, BanchoLobbyJoin, writer.Bytes())
}

func (client *b334) WriteLobbyPart(stream io.Writer, userId int32) error {
	writer := bytes.NewBuffer([]byte{})
	writeInt32(writer, userId)
	return client.WritePacket(stream, BanchoLobbyPart, writer.Bytes())
}

func (client *b334) WriteMatchJoinSuccess(stream io.Writer, match Match) error {
	writer := bytes.NewBuffer([]byte{})
	client.WriteMatch(writer, match)
	return client.WritePacket(stream, BanchoMatchJoinSuccess, writer.Bytes())
}

func (client *b334) WriteMatchJoinFail(stream io.Writer) error {
	return client.WritePacket(stream, BanchoMatchJoinFail, []byte{})
}

func (client *b334) WriteFellowSpectatorJoined(stream io.Writer, userId int32) error {
	writer := bytes.NewBuffer([]byte{})
	writeInt32(writer, userId)
	return client.WritePacket(stream, BanchoFellowSpectatorJoined, writer.Bytes())
}

func (client *b334) WriteFellowSpectatorLeft(stream io.Writer, userId int32) error {
	writer := bytes.NewBuffer([]byte{})
	writeInt32(writer, userId)
	return client.WritePacket(stream, BanchoFellowSpectatorLeft, writer.Bytes())
}

func (client *b334) WriteMatchStart(stream io.Writer, match Match) error {
	writer := bytes.NewBuffer([]byte{})
	client.WriteMatch(writer, match)
	return client.WritePacket(stream, BanchoMatchStart, writer.Bytes())
}

func (client *b334) WriteMatchScoreUpdate(stream io.Writer, frame ScoreFrame) error {
	writer := bytes.NewBuffer([]byte{})
	client.WriteScoreFrame(writer, frame)
	return client.WritePacket(stream, BanchoMatchScoreUpdate, writer.Bytes())
}

/* Unsupported Packets */

func (client *b334) WriteMatchSkip(stream io.Writer) error                          { return nil }
func (client *b334) WriteUnauthorized(stream io.Writer) error                       { return nil }
func (client *b334) WriteChannelJoinSuccess(stream io.Writer, channel string) error { return nil }
func (client *b334) WriteChannelRevoked(stream io.Writer, channel string) error     { return nil }
func (client *b334) WriteChannelAvailable(stream io.Writer, channel Channel) error  { return nil }
func (client *b334) WriteChannelAvailableAutojoin(stream io.Writer, channel Channel) error {
	return nil
}
func (client *b334) WriteBeatmapInfoReply(stream io.Writer, reply BeatmapInfoReply) error { return nil }
func (client *b334) WriteLoginPermissions(stream io.Writer, permissions uint32) error     { return nil }
func (client *b334) WriteFriendsList(stream io.Writer, userIds []int32) error             { return nil }
func (client *b334) WriteProtocolNegotiation(stream io.Writer, version int32) error       { return nil }
func (client *b334) WriteTitleUpdate(stream io.Writer, update TitleUpdate) error          { return nil }
func (client *b334) WriteMonitor(stream io.Writer) error                                  { return nil }
func (client *b334) WriteMatchPlayerSkipped(stream io.Writer, slotId int32) error         { return nil }
func (client *b334) WriteRestart(stream io.Writer, retryMs int32) error                   { return nil }
func (client *b334) WriteInvite(stream io.Writer, message Message) error                  { return nil }
func (client *b334) WriteChannelInfoComplete(stream io.Writer) error                      { return nil }
func (client *b334) WriteMatchChangePassword(stream io.Writer, password string) error     { return nil }
func (client *b334) WriteSilenceInfo(stream io.Writer, timeRemaining int32) error         { return nil }
func (client *b334) WriteUserSilenced(stream io.Writer, userId uint32) error              { return nil }
func (client *b334) WriteUserDMsBlocked(stream io.Writer, targetName string) error        { return nil }
func (client *b334) WriteTargetIsSilenced(stream io.Writer, targetName string) error      { return nil }
func (client *b334) WriteVersionUpdateForced(stream io.Writer) error                      { return nil }
func (client *b334) WriteSwitchServer(stream io.Writer, target int32) error               { return nil }
func (client *b334) WriteAccountRestricted(stream io.Writer) error                        { return nil }
func (client *b334) WriteRTX(stream io.Writer, message string) error                      { return nil }
func (client *b334) WriteMatchAbort(stream io.Writer) error                               { return nil }
func (client *b334) WriteSwitchTournamentServer(stream io.Writer, ip string) error        { return nil }
