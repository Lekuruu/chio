package chio

import (
	"bytes"
	"fmt"
	"io"
)

// b320 adds support for multiple channels
type b320 struct {
	BanchoIO
	stream           io.ReadWriteCloser
	previous         *b312
	slotSize         int
	supportedPackets []uint16
}

func (client *b320) Write(p []byte) (n int, err error) {
	return client.stream.Write(p)
}

func (client *b320) Read(p []byte) (n int, err error) {
	return client.stream.Read(p)
}

func (client *b320) Close() error {
	return client.stream.Close()
}

func (client *b320) Clone() BanchoIO {
	previous := &b312{}
	clone := previous.Clone()
	return &b320{
		previous: clone.(*b312),
		slotSize: 8,
	}
}

func (client *b320) GetStream() io.ReadWriteCloser {
	return client.stream
}

func (client *b320) SetStream(stream io.ReadWriteCloser) {
	client.stream = stream
	client.previous.SetStream(stream)
}

func (client *b320) WritePacket(packetId uint16, data []byte) error {
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

func (client *b320) ReadPacket() (packet *BanchoPacket, err error) {
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

func (client *b320) SupportedPackets() []uint16 {
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

func (client *b320) ImplementsPacket(packetId uint16) bool {
	for _, id := range client.SupportedPackets() {
		if id == packetId {
			return true
		}
	}
	return false
}

func (client *b320) OverrideMatchSlotSize(amount int) {
	client.slotSize = amount
	client.previous.OverrideMatchSlotSize(amount)
}

func (client *b320) ConvertInputPacketId(packetId uint16) uint16 {
	if packetId == 11 {
		// "IrcJoin" packet
		return BanchoHandleIrcJoin
	}
	if packetId > 11 {
		return packetId - 1
	}
	return packetId
}

func (client *b320) ConvertOutputPacketId(packetId uint16) uint16 {
	if packetId == BanchoHandleIrcJoin {
		// "IrcJoin" packet
		return 11
	}
	if packetId >= 11 {
		return packetId + 1
	}
	return packetId
}

func (client *b320) ReadPacketType(packetId uint16, reader io.Reader) (any, error) {
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
	default:
		return nil, nil
	}
}

/* New Packets */

func (client *b320) WriteMessage(message Message) error {
	writer := bytes.NewBuffer([]byte{})
	writeString(writer, message.Sender)
	writeString(writer, message.Content)
	writeString(writer, message.Target)
	return client.WritePacket(BanchoSendMessage, writer.Bytes())
}

func (client *b320) ReadMessage(reader io.Reader) (*Message, error) {
	content, err := readString(reader)
	if err != nil {
		return nil, err
	}

	sender, err := readString(reader)
	if err != nil {
		return nil, err
	}

	target, err := readString(reader)
	if err != nil {
		return nil, err
	}

	return &Message{
		Sender:  sender,
		Content: content,
		Target:  target,
	}, nil
}

/* Inherited Packets */

func (client *b320) WriteVersionUpdate() error {
	return client.previous.WriteVersionUpdate()
}

func (client *b320) WriteLoginReply(reply int32) error {
	return client.previous.WriteLoginReply(reply)
}

func (client *b320) WritePing() error {
	return client.previous.WritePing()
}

func (client *b320) WriteIrcChangeUsername(oldName string, newName string) error {
	return client.previous.WriteIrcChangeUsername(oldName, newName)
}

func (client *b320) WriteUserStats(info UserInfo) error {
	return client.previous.WriteUserStats(info)
}

func (client *b320) WriteUserQuit(quit UserQuit) error {
	return client.previous.WriteUserQuit(quit)
}

func (client *b320) WriteSpectateFrames(bundle ReplayFrameBundle) error {
	return client.previous.WriteSpectateFrames(bundle)
}

func (client *b320) WriteSpectatorJoined(userId int32) error {
	return client.previous.WriteSpectatorJoined(userId)
}

func (client *b320) WriteSpectatorLeft(userId int32) error {
	return client.previous.WriteSpectatorLeft(userId)
}

func (client *b320) WriteSpectatorCantSpectate(userId int32) error {
	return client.previous.WriteSpectatorCantSpectate(userId)
}

func (client *b320) WriteScoreFrame(writer io.Writer, frame ScoreFrame) error {
	return client.previous.WriteScoreFrame(writer, frame)
}

func (client *b320) WriteStatus(writer io.Writer, status *UserStatus) error {
	return client.previous.WriteStatus(writer, status)
}

func (client *b320) WriteStats(writer io.Writer, info UserInfo) error {
	return client.previous.WriteStats(writer, info)
}

func (client *b320) WriteMatch(writer io.Writer, match Match) error {
	return client.previous.WriteMatch(writer, match)
}

func (client *b320) WriteMatchStart(match Match) error {
	return client.previous.WriteMatchStart(match)
}

func (client *b320) WriteMatchScoreUpdate(frame ScoreFrame) error {
	return client.previous.WriteMatchScoreUpdate(frame)
}

func (client *b320) WriteUserPresence(info UserInfo) error {
	return client.previous.WriteUserPresence(info)
}

func (client *b320) WriteUserPresenceSingle(info UserInfo) error {
	return client.previous.WriteUserPresenceSingle(info)
}

func (client *b320) WriteUserPresenceBundle(infos []UserInfo) error {
	return client.previous.WriteUserPresenceBundle(infos)
}

func (client *b320) WriteGetAttention() error {
	return client.previous.WriteGetAttention()
}

func (client *b320) WriteAnnouncement(message string) error {
	return client.previous.WriteAnnouncement(message)
}

func (client *b320) WriteMatchUpdate(match Match) error {
	return client.previous.WriteMatchUpdate(match)
}

func (client *b320) WriteMatchNew(match Match) error {
	return client.previous.WriteMatchNew(match)
}

func (client *b320) WriteMatchDisband(matchId uint32) error {
	return client.previous.WriteMatchDisband(matchId)
}

func (client *b320) WriteLobbyJoin(userId int32) error {
	return client.previous.WriteLobbyJoin(userId)
}

func (client *b320) WriteLobbyPart(userId int32) error {
	return client.previous.WriteLobbyPart(userId)
}

func (client *b320) WriteMatchJoinSuccess(match Match) error {
	return client.previous.WriteMatchJoinSuccess(match)
}

func (client *b320) WriteMatchJoinFail() error {
	return client.previous.WriteMatchJoinFail()
}

func (client *b320) WriteFellowSpectatorJoined(userId int32) error {
	return client.previous.WriteFellowSpectatorJoined(userId)
}

func (client *b320) WriteFellowSpectatorLeft(userId int32) error {
	return client.previous.WriteFellowSpectatorLeft(userId)
}

func (client *b320) ReadStatus(reader io.Reader) (any, error) {
	return client.previous.ReadStatus(reader)
}

func (client *b320) ReadReplayFrame(reader io.Reader) (*ReplayFrame, error) {
	return client.previous.ReadReplayFrame(reader)
}

func (client b320) ReadScoreFrame(reader io.Reader) (*ScoreFrame, error) {
	return client.previous.ReadScoreFrame(reader)
}

func (client *b320) ReadFrameBundle(reader io.Reader) (*ReplayFrameBundle, error) {
	return client.previous.ReadFrameBundle(reader)
}

func (client *b320) ReadMatch(reader io.Reader) (Match, error) {
	return client.previous.ReadMatch(reader)
}

/* Unsupported Packets */

func (client *b320) WriteMatchTransferHost() error                       { return nil }
func (client *b320) WriteMatchAllPlayersLoaded() error                   { return nil }
func (client *b320) WriteMatchPlayerFailed(slotId uint32) error          { return nil }
func (client *b320) WriteMatchComplete() error                           { return nil }
func (client *b320) WriteMatchSkip() error                               { return nil }
func (client *b320) WriteUnauthorized() error                            { return nil }
func (client *b320) WriteChannelJoinSuccess(channel string) error        { return nil }
func (client *b320) WriteChannelRevoked(channel string) error            { return nil }
func (client *b320) WriteChannelAvailable(channel Channel) error         { return nil }
func (client *b320) WriteChannelAvailableAutojoin(channel Channel) error { return nil }
func (client *b320) WriteBeatmapInfoReply(reply BeatmapInfoReply) error  { return nil }
func (client *b320) WriteLoginPermissions(permissions uint32) error      { return nil }
func (client *b320) WriteFriendsList(userIds []uint32) error             { return nil }
func (client *b320) WriteProtocolNegotiation(version int32) error        { return nil }
func (client *b320) WriteTitleUpdate(update TitleUpdate) error           { return nil }
func (client *b320) WriteMonitor() error                                 { return nil }
func (client *b320) WriteMatchPlayerSkipped(slotId int32) error          { return nil }
func (client *b320) WriteRestart(retryMs int32) error                    { return nil }
func (client *b320) WriteInvite(message Message) error                   { return nil }
func (client *b320) WriteChannelInfoComplete() error                     { return nil }
func (client *b320) WriteMatchChangePassword(password string) error      { return nil }
func (client *b320) WriteSilenceInfo(timeRemaining int32) error          { return nil }
func (client *b320) WriteUserSilenced(userId uint32) error               { return nil }
func (client *b320) WriteUserDMsBlocked(targetName string) error         { return nil }
func (client *b320) WriteTargetIsSilenced(targetName string) error       { return nil }
func (client *b320) WriteVersionUpdateForced() error                     { return nil }
func (client *b320) WriteSwitchServer(target int32) error                { return nil }
func (client *b320) WriteAccountRestricted() error                       { return nil }
func (client *b320) WriteRTX(message string) error                       { return nil }
func (client *b320) WriteMatchAbort() error                              { return nil }
func (client *b320) WriteSwitchTournamentServer(ip string) error         { return nil }
