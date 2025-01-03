package chio

import (
	"bytes"
	"fmt"
	"io"
)

// b312 adds the match start & update packets, as well
// as the "InProgress" boolean inside the match struct
type b312 struct {
	BanchoIO
	stream           io.ReadWriteCloser
	previous         *b298
	slotSize         int
	supportedPackets []uint16
}

func (client *b312) Write(p []byte) (n int, err error) {
	return client.stream.Write(p)
}

func (client *b312) Read(p []byte) (n int, err error) {
	return client.stream.Read(p)
}

func (client *b312) Close() error {
	return client.stream.Close()
}

func (client *b312) Clone() BanchoIO {
	previous := &b298{}
	clone := previous.Clone()
	return &b312{
		previous: clone.(*b298),
		slotSize: 8,
	}
}

func (client *b312) GetStream() io.ReadWriteCloser {
	return client.stream
}

func (client *b312) SetStream(stream io.ReadWriteCloser) {
	client.stream = stream
	client.previous.SetStream(stream)
}

func (client *b312) WritePacket(packetId uint16, data []byte) error {
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

func (client *b312) ReadPacket() (packet *BanchoPacket, err error) {
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

func (client *b312) SupportedPackets() []uint16 {
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
	}
	return client.supportedPackets
}

func (client *b312) ImplementsPacket(packetId uint16) bool {
	for _, id := range client.SupportedPackets() {
		if id == packetId {
			return true
		}
	}
	return false
}

func (client *b312) OverrideMatchSlotSize(amount int) {
	client.slotSize = amount
	client.previous.OverrideMatchSlotSize(amount)
}

func (client *b312) ConvertInputPacketId(packetId uint16) uint16 {
	if packetId == 11 {
		// "IrcJoin" packet
		return BanchoHandleIrcJoin
	}
	if packetId > 11 {
		return packetId - 1
	}
	return packetId
}

func (client *b312) ConvertOutputPacketId(packetId uint16) uint16 {
	if packetId == BanchoHandleIrcJoin {
		// "IrcJoin" packet
		return 11
	}
	if packetId >= 11 {
		return packetId + 1
	}
	return packetId
}

func (client *b312) ReadPacketType(packetId uint16, reader io.Reader) (any, error) {
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
	case OsuMatchScoreUpdate:
		return client.ReadScoreFrame(reader)
	default:
		return nil, nil
	}
}

/* New Packets */

func (client *b312) WriteMatchStart(match Match) error {
	writer := bytes.NewBuffer([]byte{})
	client.WriteMatch(writer, match)
	return client.WritePacket(BanchoMatchStart, writer.Bytes())
}

func (client *b312) WriteMatchScoreUpdate(frame ScoreFrame) error {
	writer := bytes.NewBuffer([]byte{})
	client.WriteScoreFrame(writer, frame)
	return client.WritePacket(BanchoMatchScoreUpdate, writer.Bytes())
}

func (client *b312) WriteMatchUpdate(match Match) error {
	writer := bytes.NewBuffer([]byte{})
	client.WriteMatch(writer, match)
	return client.WritePacket(BanchoMatchUpdate, writer.Bytes())
}

func (client *b312) WriteMatchNew(match Match) error {
	writer := bytes.NewBuffer([]byte{})
	client.WriteMatch(writer, match)
	return client.WritePacket(BanchoMatchNew, writer.Bytes())
}

func (client *b312) WriteMatchJoinSuccess(match Match) error {
	writer := bytes.NewBuffer([]byte{})
	client.WriteMatch(writer, match)
	return client.WritePacket(BanchoMatchJoinSuccess, writer.Bytes())
}

func (client *b312) WriteMatch(writer io.Writer, match Match) error {
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
	writeBoolean(writer, match.InProgress)
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

func (client *b312) ReadMatch(reader io.Reader) (Match, error) {
	var err error
	errors := NewErrorCollection()
	match := Match{}

	matchId, err := readUint8(reader)
	match.Id = int32(matchId)
	errors.Add(err)
	match.InProgress, err = readBoolean(reader)
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

func (client *b312) WriteMessage(message Message) error {
	return client.previous.WriteMessage(message)
}

func (client *b312) WriteVersionUpdate() error {
	return client.previous.WriteVersionUpdate()
}

func (client *b312) WriteLoginReply(reply int32) error {
	return client.previous.WriteLoginReply(reply)
}

func (client *b312) WritePing() error {
	return client.previous.WritePing()
}

func (client *b312) WriteIrcChangeUsername(oldName string, newName string) error {
	return client.previous.WriteIrcChangeUsername(oldName, newName)
}

func (client *b312) WriteUserStats(info UserInfo) error {
	return client.previous.WriteUserStats(info)
}

func (client *b312) WriteUserQuit(quit UserQuit) error {
	return client.previous.WriteUserQuit(quit)
}

func (client *b312) WriteSpectateFrames(bundle ReplayFrameBundle) error {
	return client.previous.WriteSpectateFrames(bundle)
}

func (client *b312) WriteSpectatorJoined(userId int32) error {
	return client.previous.WriteSpectatorJoined(userId)
}

func (client *b312) WriteSpectatorLeft(userId int32) error {
	return client.previous.WriteSpectatorLeft(userId)
}

func (client *b312) WriteSpectatorCantSpectate(userId int32) error {
	return client.previous.WriteSpectatorCantSpectate(userId)
}

func (client *b312) WriteScoreFrame(writer io.Writer, frame ScoreFrame) error {
	return client.previous.WriteScoreFrame(writer, frame)
}

func (client *b312) WriteStatus(writer io.Writer, status *UserStatus) error {
	return client.previous.WriteStatus(writer, status)
}

func (client *b312) WriteStats(writer io.Writer, info UserInfo) error {
	return client.previous.WriteStats(writer, info)
}

func (client *b312) WriteUserPresence(info UserInfo) error {
	return client.previous.WriteUserPresence(info)
}

func (client *b312) WriteUserPresenceSingle(info UserInfo) error {
	return client.previous.WriteUserPresenceSingle(info)
}

func (client *b312) WriteUserPresenceBundle(infos []UserInfo) error {
	return client.previous.WriteUserPresenceBundle(infos)
}

func (client *b312) WriteGetAttention() error {
	return client.previous.WriteGetAttention()
}

func (client *b312) WriteAnnouncement(message string) error {
	return client.previous.WriteAnnouncement(message)
}

func (client *b312) WriteMatchDisband(matchId uint32) error {
	return client.previous.WriteMatchDisband(matchId)
}

func (client *b312) WriteLobbyJoin(userId int32) error {
	return client.previous.WriteLobbyJoin(userId)
}

func (client *b312) WriteLobbyPart(userId int32) error {
	return client.previous.WriteLobbyPart(userId)
}

func (client *b312) WriteMatchJoinFail() error {
	return client.previous.WriteMatchJoinFail()
}

func (client *b312) WriteFellowSpectatorJoined(userId int32) error {
	return client.previous.WriteFellowSpectatorJoined(userId)
}

func (client *b312) WriteFellowSpectatorLeft(userId int32) error {
	return client.previous.WriteFellowSpectatorLeft(userId)
}

func (client *b312) ReadStatus(reader io.Reader) (any, error) {
	return client.previous.ReadStatus(reader)
}

func (client *b312) ReadReplayFrame(reader io.Reader) (*ReplayFrame, error) {
	return client.previous.ReadReplayFrame(reader)
}

func (client *b312) ReadMessage(reader io.Reader) (*Message, error) {
	return client.previous.ReadMessage(reader)
}

func (client *b312) ReadMessagePrivate(reader io.Reader) (*Message, error) {
	return client.previous.ReadMessagePrivate(reader)
}

func (client b312) ReadScoreFrame(reader io.Reader) (*ScoreFrame, error) {
	return client.previous.ReadScoreFrame(reader)
}

func (client *b312) ReadFrameBundle(reader io.Reader) (*ReplayFrameBundle, error) {
	return client.previous.ReadFrameBundle(reader)
}

/* Unsupported Packets */

func (client *b312) WriteMatchTransferHost() error                       { return nil }
func (client *b312) WriteMatchAllPlayersLoaded() error                   { return nil }
func (client *b312) WriteMatchPlayerFailed(slotId uint32) error          { return nil }
func (client *b312) WriteMatchComplete() error                           { return nil }
func (client *b312) WriteMatchSkip() error                               { return nil }
func (client *b312) WriteUnauthorized() error                            { return nil }
func (client *b312) WriteChannelJoinSuccess(channel string) error        { return nil }
func (client *b312) WriteChannelRevoked(channel string) error            { return nil }
func (client *b312) WriteChannelAvailable(channel Channel) error         { return nil }
func (client *b312) WriteChannelAvailableAutojoin(channel Channel) error { return nil }
func (client *b312) WriteBeatmapInfoReply(reply BeatmapInfoReply) error  { return nil }
func (client *b312) WriteLoginPermissions(permissions uint32) error      { return nil }
func (client *b312) WriteFriendsList(userIds []uint32) error             { return nil }
func (client *b312) WriteProtocolNegotiation(version int32) error        { return nil }
func (client *b312) WriteTitleUpdate(update TitleUpdate) error           { return nil }
func (client *b312) WriteMonitor() error                                 { return nil }
func (client *b312) WriteMatchPlayerSkipped(slotId int32) error          { return nil }
func (client *b312) WriteRestart(retryMs int32) error                    { return nil }
func (client *b312) WriteInvite(message Message) error                   { return nil }
func (client *b312) WriteChannelInfoComplete() error                     { return nil }
func (client *b312) WriteMatchChangePassword(password string) error      { return nil }
func (client *b312) WriteSilenceInfo(timeRemaining int32) error          { return nil }
func (client *b312) WriteUserSilenced(userId uint32) error               { return nil }
func (client *b312) WriteUserDMsBlocked(targetName string) error         { return nil }
func (client *b312) WriteTargetIsSilenced(targetName string) error       { return nil }
func (client *b312) WriteVersionUpdateForced() error                     { return nil }
func (client *b312) WriteSwitchServer(target int32) error                { return nil }
func (client *b312) WriteAccountRestricted() error                       { return nil }
func (client *b312) WriteRTX(message string) error                       { return nil }
func (client *b312) WriteMatchAbort() error                              { return nil }
func (client *b312) WriteSwitchTournamentServer(ip string) error         { return nil }
