package chio

import (
	"bytes"
	"fmt"
	"io"
)

// b291 implements the GetAttension & Announce packets
type b291 struct {
	BanchoIO
	stream           io.ReadWriteCloser
	supportedPackets []uint16
	previous         *b282
}

func (client *b291) Write(p []byte) (n int, err error) {
	return client.stream.Write(p)
}

func (client *b291) Read(p []byte) (n int, err error) {
	return client.stream.Read(p)
}

func (client *b291) Close() error {
	return client.stream.Close()
}

func (client *b291) Clone() BanchoIO {
	return &b291{previous: &b282{}}
}

func (client *b291) GetStream() io.ReadWriteCloser {
	return client.stream
}

func (client *b291) SetStream(stream io.ReadWriteCloser) {
	client.stream = stream
	client.previous.SetStream(stream)
}

func (client *b291) WritePacket(packetId uint16, data []byte) error {
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

func (client *b291) ReadPacket() (packet *BanchoPacket, err error) {
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

func (client *b291) SupportedPackets() []uint16 {
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
	}
	return client.supportedPackets
}

func (client *b291) ImplementsPacket(packetId uint16) bool {
	for _, id := range client.SupportedPackets() {
		if id == packetId {
			return true
		}
	}
	return false
}

func (client *b291) OverrideMatchSlotSize(amount int) {
	// Multiplayer is not supported in this version
}

func (client *b291) ConvertInputPacketId(packetId uint16) uint16 {
	if packetId == 11 {
		// "IrcJoin" packet
		return BanchoHandleIrcJoin
	}
	if packetId > 11 {
		return packetId - 1
	}
	return packetId
}

func (client *b291) ConvertOutputPacketId(packetId uint16) uint16 {
	if packetId == BanchoHandleIrcJoin {
		// "IrcJoin" packet
		return 11
	}
	if packetId >= 11 {
		return packetId + 1
	}
	return packetId
}

func (client *b291) ReadPacketType(packetId uint16, reader io.Reader) (any, error) {
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
	default:
		return nil, nil
	}
}

/* New Packets */

func (client *b291) WriteGetAttention() error {
	return client.WritePacket(BanchoGetAttention, []byte{})
}

func (client *b291) WriteAnnouncement(message string) error {
	writer := bytes.NewBuffer([]byte{})
	writeString(writer, message)
	return client.WritePacket(BanchoAnnounce, writer.Bytes())
}

/* Inherited Packets */

func (client *b291) ReadStatus(reader io.Reader) (any, error) {
	return client.previous.ReadStatus(reader)
}

func (client *b291) ReadMessage(reader io.Reader) (*Message, error) {
	return client.previous.ReadMessage(reader)
}

func (client *b291) ReadFrameBundle(reader io.Reader) (*ReplayFrameBundle, error) {
	return client.previous.ReadFrameBundle(reader)
}

func (client *b291) ReadReplayFrame(reader io.Reader) (*ReplayFrame, error) {
	return client.previous.ReadReplayFrame(reader)
}

func (client *b291) WriteLoginReply(reply int32) error {
	return client.previous.WriteLoginReply(reply)
}

func (client *b291) WriteMessage(message Message) error {
	return client.previous.WriteMessage(message)
}

func (client *b291) WritePing() error {
	return client.previous.WritePing()
}

func (client *b291) WriteIrcChangeUsername(oldName string, newName string) error {
	return client.previous.WriteIrcChangeUsername(oldName, newName)
}

func (client *b291) WriteUserStats(info UserInfo) error {
	return client.previous.WriteUserStats(info)
}

func (client *b291) WriteUserQuit(quit UserQuit) error {
	return client.previous.WriteUserQuit(quit)
}

func (client *b291) WriteSpectatorJoined(userId int32) error {
	return client.previous.WriteSpectatorJoined(userId)
}

func (client *b291) WriteSpectatorLeft(userId int32) error {
	return client.previous.WriteSpectatorLeft(userId)
}

func (client *b291) WriteSpectateFrames(bundle ReplayFrameBundle) error {
	return client.previous.WriteSpectateFrames(bundle)
}

func (client *b291) WriteVersionUpdate() error {
	return client.previous.WriteVersionUpdate()
}

func (client *b291) WriteSpectatorCantSpectate(userId int32) error {
	return client.previous.WriteSpectatorCantSpectate(userId)
}

func (client *b291) WriteUserPresence(info UserInfo) error {
	return client.previous.WriteUserPresence(info)
}

func (client *b291) WriteUserPresenceSingle(info UserInfo) error {
	return client.previous.WriteUserPresenceSingle(info)
}

func (client *b291) WriteUserPresenceBundle(infos []UserInfo) error {
	return client.previous.WriteUserPresenceBundle(infos)
}

func (client *b291) WriteStatus(writer io.Writer, status *UserStatus) error {
	return client.previous.WriteStatus(writer, status)
}

func (client *b291) WriteStats(writer io.Writer, info UserInfo) error {
	return client.previous.WriteStats(writer, info)
}

/* Unsupported Packets */

func (client *b291) WriteMatchUpdate(match Match) error                  { return nil }
func (client *b291) WriteMatchNew(match Match) error                     { return nil }
func (client *b291) WriteMatchDisband(matchId uint32) error              { return nil }
func (client *b291) WriteLobbyJoin(userId int32) error                   { return nil }
func (client *b291) WriteLobbyPart(userId int32) error                   { return nil }
func (client *b291) WriteMatchJoinSuccess(match Match) error             { return nil }
func (client *b291) WriteMatchJoinFail() error                           { return nil }
func (client *b291) WriteFellowSpectatorJoined(userId int32) error       { return nil }
func (client *b291) WriteFellowSpectatorLeft(userId int32) error         { return nil }
func (client *b291) WriteMatchStart(match Match) error                   { return nil }
func (client *b291) WriteMatchScoreUpdate(frame ScoreFrame) error        { return nil }
func (client *b291) WriteMatchTransferHost() error                       { return nil }
func (client *b291) WriteMatchAllPlayersLoaded() error                   { return nil }
func (client *b291) WriteMatchPlayerFailed(slotId uint32) error          { return nil }
func (client *b291) WriteMatchComplete() error                           { return nil }
func (client *b291) WriteMatchSkip() error                               { return nil }
func (client *b291) WriteUnauthorized() error                            { return nil }
func (client *b291) WriteChannelJoinSuccess(channel string) error        { return nil }
func (client *b291) WriteChannelRevoked(channel string) error            { return nil }
func (client *b291) WriteChannelAvailable(channel Channel) error         { return nil }
func (client *b291) WriteChannelAvailableAutojoin(channel Channel) error { return nil }
func (client *b291) WriteBeatmapInfoReply(reply BeatmapInfoReply) error  { return nil }
func (client *b291) WriteLoginPermissions(permissions uint32) error      { return nil }
func (client *b291) WriteFriendsList(userIds []uint32) error             { return nil }
func (client *b291) WriteProtocolNegotiation(version int32) error        { return nil }
func (client *b291) WriteTitleUpdate(update TitleUpdate) error           { return nil }
func (client *b291) WriteMonitor() error                                 { return nil }
func (client *b291) WriteMatchPlayerSkipped(slotId int32) error          { return nil }
func (client *b291) WriteRestart(retryMs int32) error                    { return nil }
func (client *b291) WriteInvite(message Message) error                   { return nil }
func (client *b291) WriteChannelInfoComplete() error                     { return nil }
func (client *b291) WriteMatchChangePassword(password string) error      { return nil }
func (client *b291) WriteSilenceInfo(timeRemaining int32) error          { return nil }
func (client *b291) WriteUserSilenced(userId uint32) error               { return nil }
func (client *b291) WriteUserDMsBlocked(targetName string) error         { return nil }
func (client *b291) WriteTargetIsSilenced(targetName string) error       { return nil }
func (client *b291) WriteVersionUpdateForced() error                     { return nil }
func (client *b291) WriteSwitchServer(target int32) error                { return nil }
func (client *b291) WriteAccountRestricted() error                       { return nil }
func (client *b291) WriteRTX(message string) error                       { return nil }
func (client *b291) WriteMatchAbort() error                              { return nil }
func (client *b291) WriteSwitchTournamentServer(ip string) error         { return nil }
