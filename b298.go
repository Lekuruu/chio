package chio

import (
	"bytes"
	"fmt"
	"io"
)

// b298 adds a partial implementation of multiplayer, as well as fellow spectators
type b298 struct {
	BanchoIO
	stream           io.ReadWriteCloser
	previous         *b296
	slotSize         int
	supportedPackets []uint16
}

func (client *b298) Write(p []byte) (n int, err error) {
	return client.stream.Write(p)
}

func (client *b298) Read(p []byte) (n int, err error) {
	return client.stream.Read(p)
}

func (client *b298) Close() error {
	return client.stream.Close()
}

func (client *b298) Clone() BanchoIO {
	previous := &b296{}
	clone := previous.Clone()
	return &b298{
		previous: clone.(*b296),
		slotSize: 8,
	}
}

func (client *b298) GetStream() io.ReadWriteCloser {
	return client.stream
}

func (client *b298) SetStream(stream io.ReadWriteCloser) {
	client.stream = stream
	client.previous.SetStream(stream)
}

func (client *b298) WritePacket(packetId uint16, data []byte) error {
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

func (client *b298) ReadPacket() (packet *BanchoPacket, err error) {
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

func (client *b298) SupportedPackets() []uint16 {
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
	}
	return client.supportedPackets
}

func (client *b298) ImplementsPacket(packetId uint16) bool {
	for _, id := range client.SupportedPackets() {
		if id == packetId {
			return true
		}
	}
	return false
}

func (client *b298) OverrideMatchSlotSize(amount int) {
	client.slotSize = amount
	client.previous.OverrideMatchSlotSize(amount)
}

func (client *b298) ConvertInputPacketId(packetId uint16) uint16 {
	if packetId == 11 {
		// "IrcJoin" packet
		return BanchoHandleIrcJoin
	}
	if packetId > 11 {
		return packetId - 1
	}
	return packetId
}

func (client *b298) ConvertOutputPacketId(packetId uint16) uint16 {
	if packetId == BanchoHandleIrcJoin {
		// "IrcJoin" packet
		return 11
	}
	if packetId >= 11 {
		return packetId + 1
	}
	return packetId
}

func (client *b298) ReadPacketType(packetId uint16, reader io.Reader) (any, error) {
	switch packetId {
	case OsuSendUserStatus:
		return client.ReadStatus(reader)
	case OsuSendIrcMessage:
		return client.ReadMessage(reader)
	case OsuStartSpectating:
		return readUint32(reader)
	case OsuSpectateFrames:
		return client.ReadFrameBundle(reader)
	case OsuErrorReport:
		return readString(reader)
	case OsuSendIrcMessagePrivate:
		return client.ReadMessagePrivate(reader)
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
	default:
		return nil, nil
	}
}

/* New Packets */

func (client *b298) WriteMatch(writer io.Writer, match Match) error {
	slotsOpen := make([]bool, client.slotSize)
	slotsUsed := make([]bool, client.slotSize)
	slotsReady := make([]bool, client.slotSize)

	for index, slot := range match.Slots {
		slotsOpen[index] = slot.Status == SlotStatusOpen
	}

	for index, slot := range match.Slots {
		slotsUsed[index] = slot.HasPlayer()
	}

	for index, slot := range match.Slots {
		slotsReady[index] = slot.Status == SlotStatusReady
	}

	writeUint8(writer, uint8(match.Id))
	writeUint8(writer, match.Type)
	writeString(writer, match.Name)
	writeString(writer, match.BeatmapText)
	writeInt32(writer, match.BeatmapId)
	writeString(writer, match.BeatmapChecksum)
	writeBoolList(writer, slotsOpen)
	writeBoolList(writer, slotsUsed)
	writeBoolList(writer, slotsReady)

	for _, slot := range match.Slots {
		if slot.HasPlayer() {
			writeInt32(writer, slot.UserId)
		}
	}

	return nil
}

func (client *b298) WriteMatchUpdate(match Match) error {
	writer := bytes.NewBuffer([]byte{})
	client.WriteMatch(writer, match)
	return client.WritePacket(BanchoMatchUpdate, writer.Bytes())
}

func (client *b298) WriteMatchNew(match Match) error {
	writer := bytes.NewBuffer([]byte{})
	client.WriteMatch(writer, match)
	return client.WritePacket(BanchoMatchNew, writer.Bytes())
}

func (client *b298) WriteMatchDisband(matchId int32) error {
	writer := bytes.NewBuffer([]byte{})
	writeInt32(writer, matchId)
	return client.WritePacket(BanchoMatchDisband, writer.Bytes())
}

func (client *b298) WriteLobbyJoin(userId int32) error {
	writer := bytes.NewBuffer([]byte{})
	writeInt32(writer, userId)
	return client.WritePacket(OsuLobbyJoin, writer.Bytes())
}

func (client *b298) WriteLobbyPart(userId int32) error {
	writer := bytes.NewBuffer([]byte{})
	writeInt32(writer, userId)
	return client.WritePacket(OsuLobbyPart, writer.Bytes())
}

func (client *b298) WriteMatchJoinSuccess(match Match) error {
	writer := bytes.NewBuffer([]byte{})
	client.WriteMatch(writer, match)
	return client.WritePacket(BanchoMatchJoinSuccess, writer.Bytes())
}

func (client *b298) WriteMatchJoinFail() error {
	return client.WritePacket(BanchoMatchJoinFail, []byte{})
}

func (client *b298) WriteFellowSpectatorJoined(userId int32) error {
	writer := bytes.NewBuffer([]byte{})
	writeInt32(writer, userId)
	return client.WritePacket(BanchoFellowSpectatorJoined, writer.Bytes())
}

func (client *b298) WriteFellowSpectatorLeft(userId int32) error {
	writer := bytes.NewBuffer([]byte{})
	writeInt32(writer, userId)
	return client.WritePacket(BanchoFellowSpectatorLeft, writer.Bytes())
}

func (client *b298) ReadMatch(reader io.Reader) (Match, error) {
	var err error
	errors := NewErrorCollection()
	match := Match{}

	matchId, err := readUint8(reader)
	match.Id = int32(matchId)
	errors.Add(err)
	match.Type, err = readUint8(reader)
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
	slotsOpen, err := readBoolList(reader)
	errors.Add(err)
	slotsUsed, err := readBoolList(reader)
	errors.Add(err)
	slotsReady, err := readBoolList(reader)
	errors.Add(err)

	for i := 0; i < client.slotSize; i++ {
		slot := &MatchSlot{}
		slot.Status = SlotStatusOpen
		slot.Team = SlotTeamNeutral
		slot.Mods = NoMod
		slot.UserId = -1

		if slotsOpen[i] {
			slot.Status = SlotStatusOpen
		}

		if slotsUsed[i] {
			slot.Status = SlotStatusNotReady
		}

		if slotsReady[i] {
			slot.Status = SlotStatusReady
		}

		if slot.HasPlayer() {
			slot.UserId, err = readInt32(reader)
			errors.Add(err)
		}

		match.Slots[i] = slot
	}

	return match, errors.Next()
}

/* Inherited Packets */

func (client *b298) WriteMessage(message Message) error {
	return client.previous.WriteMessage(message)
}

func (client *b298) WriteVersionUpdate() error {
	return client.previous.WriteVersionUpdate()
}

func (client *b298) WriteLoginReply(reply int32) error {
	return client.previous.WriteLoginReply(reply)
}

func (client *b298) WritePing() error {
	return client.previous.WritePing()
}

func (client *b298) WriteIrcChangeUsername(oldName string, newName string) error {
	return client.previous.WriteIrcChangeUsername(oldName, newName)
}

func (client *b298) WriteUserStats(info UserInfo) error {
	return client.previous.WriteUserStats(info)
}

func (client *b298) WriteUserQuit(quit UserQuit) error {
	return client.previous.WriteUserQuit(quit)
}

func (client *b298) WriteSpectateFrames(bundle ReplayFrameBundle) error {
	return client.previous.WriteSpectateFrames(bundle)
}

func (client *b298) WriteSpectatorJoined(userId int32) error {
	return client.previous.WriteSpectatorJoined(userId)
}

func (client *b298) WriteSpectatorLeft(userId int32) error {
	return client.previous.WriteSpectatorLeft(userId)
}

func (client *b298) WriteSpectatorCantSpectate(userId int32) error {
	return client.previous.WriteSpectatorCantSpectate(userId)
}

func (client *b298) WriteScoreFrame(writer io.Writer, frame ScoreFrame) error {
	return client.previous.WriteScoreFrame(writer, frame)
}

func (client *b298) WriteStatus(writer io.Writer, status *UserStatus) error {
	return client.previous.WriteStatus(writer, status)
}

func (client *b298) WriteStats(writer io.Writer, info UserInfo) error {
	return client.previous.WriteStats(writer, info)
}

func (client *b298) WriteUserPresence(info UserInfo) error {
	return client.previous.WriteUserPresence(info)
}

func (client *b298) WriteUserPresenceSingle(info UserInfo) error {
	return client.previous.WriteUserPresenceSingle(info)
}

func (client *b298) WriteUserPresenceBundle(infos []UserInfo) error {
	return client.previous.WriteUserPresenceBundle(infos)
}

func (client *b298) WriteGetAttention() error {
	return client.previous.WriteGetAttention()
}

func (client *b298) WriteAnnouncement(message string) error {
	return client.previous.WriteAnnouncement(message)
}

func (client *b298) ReadStatus(reader io.Reader) (any, error) {
	return client.previous.ReadStatus(reader)
}

func (client *b298) ReadReplayFrame(reader io.Reader) (*ReplayFrame, error) {
	return client.previous.ReadReplayFrame(reader)
}

func (client *b298) ReadMessage(reader io.Reader) (*Message, error) {
	return client.previous.ReadMessage(reader)
}

func (client *b298) ReadMessagePrivate(reader io.Reader) (*Message, error) {
	return client.previous.ReadMessagePrivate(reader)
}

func (client b298) ReadScoreFrame(reader io.Reader) (*ScoreFrame, error) {
	return client.previous.ReadScoreFrame(reader)
}

func (client *b298) ReadFrameBundle(reader io.Reader) (*ReplayFrameBundle, error) {
	return client.previous.ReadFrameBundle(reader)
}

/* Unsupported Packets */

func (client *b298) WriteMatchStart(match Match) error                   { return nil }
func (client *b298) WriteMatchScoreUpdate(frame ScoreFrame) error        { return nil }
func (client *b298) WriteMatchTransferHost() error                       { return nil }
func (client *b298) WriteMatchAllPlayersLoaded() error                   { return nil }
func (client *b298) WriteMatchPlayerFailed(slotId uint32) error          { return nil }
func (client *b298) WriteMatchComplete() error                           { return nil }
func (client *b298) WriteMatchSkip() error                               { return nil }
func (client *b298) WriteUnauthorized() error                            { return nil }
func (client *b298) WriteChannelJoinSuccess(channel string) error        { return nil }
func (client *b298) WriteChannelRevoked(channel string) error            { return nil }
func (client *b298) WriteChannelAvailable(channel Channel) error         { return nil }
func (client *b298) WriteChannelAvailableAutojoin(channel Channel) error { return nil }
func (client *b298) WriteBeatmapInfoReply(reply BeatmapInfoReply) error  { return nil }
func (client *b298) WriteLoginPermissions(permissions uint32) error      { return nil }
func (client *b298) WriteFriendsList(userIds []uint32) error             { return nil }
func (client *b298) WriteProtocolNegotiation(version int32) error        { return nil }
func (client *b298) WriteTitleUpdate(update TitleUpdate) error           { return nil }
func (client *b298) WriteMonitor() error                                 { return nil }
func (client *b298) WriteMatchPlayerSkipped(slotId int32) error          { return nil }
func (client *b298) WriteRestart(retryMs int32) error                    { return nil }
func (client *b298) WriteInvite(message Message) error                   { return nil }
func (client *b298) WriteChannelInfoComplete() error                     { return nil }
func (client *b298) WriteMatchChangePassword(password string) error      { return nil }
func (client *b298) WriteSilenceInfo(timeRemaining int32) error          { return nil }
func (client *b298) WriteUserSilenced(userId uint32) error               { return nil }
func (client *b298) WriteUserDMsBlocked(targetName string) error         { return nil }
func (client *b298) WriteTargetIsSilenced(targetName string) error       { return nil }
func (client *b298) WriteVersionUpdateForced() error                     { return nil }
func (client *b298) WriteSwitchServer(target int32) error                { return nil }
func (client *b298) WriteAccountRestricted() error                       { return nil }
func (client *b298) WriteRTX(message string) error                       { return nil }
func (client *b298) WriteMatchAbort() error                              { return nil }
func (client *b298) WriteSwitchTournamentServer(ip string) error         { return nil }
