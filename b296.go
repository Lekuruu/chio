package chio

import (
	"bytes"
	"fmt"
	"io"
)

// b296 adds the "Time" value to score frames
type b296 struct {
	BanchoIO
	stream           io.ReadWriteCloser
	supportedPackets []uint16
	previous         *b294
}

func (client *b296) Write(p []byte) (n int, err error) {
	return client.stream.Write(p)
}

func (client *b296) Read(p []byte) (n int, err error) {
	return client.stream.Read(p)
}

func (client *b296) Close() error {
	return client.stream.Close()
}

func (client *b296) Clone() BanchoIO {
	previous := &b294{}
	clone := previous.Clone()
	return &b296{previous: clone.(*b294)}
}

func (client *b296) GetStream() io.ReadWriteCloser {
	return client.stream
}

func (client *b296) SetStream(stream io.ReadWriteCloser) {
	client.stream = stream
	client.previous.SetStream(stream)
}

func (client *b296) WritePacket(packetId uint16, data []byte) error {
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

func (client *b296) ReadPacket() (packet *BanchoPacket, err error) {
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

func (client *b296) SupportedPackets() []uint16 {
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
	}
	return client.supportedPackets
}

func (client *b296) ImplementsPacket(packetId uint16) bool {
	for _, id := range client.SupportedPackets() {
		if id == packetId {
			return true
		}
	}
	return false
}

func (client *b296) OverrideMatchSlotSize(amount int) {
	// Multiplayer is not supported in this version
}

func (client *b296) ConvertInputPacketId(packetId uint16) uint16 {
	if packetId == 11 {
		// "IrcJoin" packet
		return BanchoHandleIrcJoin
	}
	if packetId > 11 {
		return packetId - 1
	}
	return packetId
}

func (client *b296) ConvertOutputPacketId(packetId uint16) uint16 {
	if packetId == BanchoHandleIrcJoin {
		// "IrcJoin" packet
		return 11
	}
	if packetId >= 11 {
		return packetId + 1
	}
	return packetId
}

func (client *b296) ReadPacketType(packetId uint16, reader io.Reader) (any, error) {
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
	default:
		return nil, nil
	}
}

/* New Packets */

func (client *b296) WriteSpectateFrames(bundle ReplayFrameBundle) error {
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

func (client *b296) WriteScoreFrame(writer io.Writer, frame ScoreFrame) error {
	writeString(writer, frame.Checksum())
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

func (client *b296) ReadFrameBundle(reader io.Reader) (*ReplayFrameBundle, error) {
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

func (client b296) ReadScoreFrame(reader io.Reader) (*ScoreFrame, error) {
	var err error
	errors := NewErrorCollection()
	frame := &ScoreFrame{}
	_, err = readString(reader) // Checksum
	errors.Add(err)
	frame.Id, err = readUint8(reader)
	errors.Add(err)
	frame.Time, err = readInt32(reader)
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
	return frame, errors.Next()
}

/* Inherited Packets */

func (client *b296) WriteMessage(message Message) error {
	return client.previous.WriteMessage(message)
}

func (client *b296) WriteVersionUpdate() error {
	return client.previous.WriteVersionUpdate()
}

func (client *b296) WriteSpectatorCantSpectate(userId int32) error {
	return client.previous.WriteSpectatorCantSpectate(userId)
}

func (client *b296) WriteLoginReply(reply int32) error {
	return client.previous.WriteLoginReply(reply)
}

func (client *b296) WritePing() error {
	return client.previous.WritePing()
}

func (client *b296) WriteIrcChangeUsername(oldName string, newName string) error {
	return client.previous.WriteIrcChangeUsername(oldName, newName)
}

func (client *b296) WriteUserStats(info UserInfo) error {
	return client.previous.WriteUserStats(info)
}

func (client *b296) WriteUserQuit(quit UserQuit) error {
	return client.previous.WriteUserQuit(quit)
}

func (client *b296) WriteSpectatorJoined(userId int32) error {
	return client.previous.WriteSpectatorJoined(userId)
}

func (client *b296) WriteSpectatorLeft(userId int32) error {
	return client.previous.WriteSpectatorLeft(userId)
}

func (client *b296) WriteStatus(writer io.Writer, status *UserStatus) error {
	return client.previous.WriteStatus(writer, status)
}

func (client *b296) WriteStats(writer io.Writer, info UserInfo) error {
	return client.previous.WriteStats(writer, info)
}

func (client *b296) WriteUserPresence(info UserInfo) error {
	return client.previous.WriteUserPresence(info)
}

func (client *b296) WriteUserPresenceSingle(info UserInfo) error {
	return client.previous.WriteUserPresenceSingle(info)
}

func (client *b296) WriteUserPresenceBundle(infos []UserInfo) error {
	return client.previous.WriteUserPresenceBundle(infos)
}

func (client *b296) WriteGetAttention() error {
	return client.previous.WriteGetAttention()
}

func (client *b296) WriteAnnouncement(message string) error {
	return client.previous.WriteAnnouncement(message)
}

func (client *b296) ReadStatus(reader io.Reader) (*UserStatus, error) {
	return client.previous.ReadStatus(reader)
}

func (client *b296) ReadReplayFrame(reader io.Reader) (*ReplayFrame, error) {
	return client.previous.ReadReplayFrame(reader)
}

func (client *b296) ReadMessage(reader io.Reader) (*Message, error) {
	return client.previous.ReadMessage(reader)
}

func (client *b296) ReadMessagePrivate(reader io.Reader) (*Message, error) {
	return client.previous.ReadMessagePrivate(reader)
}

/* Unsupported Packets */

func (client *b296) WriteMatchUpdate(match Match) error                  { return nil }
func (client *b296) WriteMatchNew(match Match) error                     { return nil }
func (client *b296) WriteMatchDisband(matchId int32) error               { return nil }
func (client *b296) WriteLobbyJoin(userId int32) error                   { return nil }
func (client *b296) WriteLobbyPart(userId int32) error                   { return nil }
func (client *b296) WriteMatchJoinSuccess(match Match) error             { return nil }
func (client *b296) WriteMatchJoinFail() error                           { return nil }
func (client *b296) WriteFellowSpectatorJoined(userId int32) error       { return nil }
func (client *b296) WriteFellowSpectatorLeft(userId int32) error         { return nil }
func (client *b296) WriteMatchStart(match Match) error                   { return nil }
func (client *b296) WriteMatchScoreUpdate(frame ScoreFrame) error        { return nil }
func (client *b296) WriteMatchTransferHost() error                       { return nil }
func (client *b296) WriteMatchAllPlayersLoaded() error                   { return nil }
func (client *b296) WriteMatchPlayerFailed(slotId uint32) error          { return nil }
func (client *b296) WriteMatchComplete() error                           { return nil }
func (client *b296) WriteMatchSkip() error                               { return nil }
func (client *b296) WriteUnauthorized() error                            { return nil }
func (client *b296) WriteChannelJoinSuccess(channel string) error        { return nil }
func (client *b296) WriteChannelRevoked(channel string) error            { return nil }
func (client *b296) WriteChannelAvailable(channel Channel) error         { return nil }
func (client *b296) WriteChannelAvailableAutojoin(channel Channel) error { return nil }
func (client *b296) WriteBeatmapInfoReply(reply BeatmapInfoReply) error  { return nil }
func (client *b296) WriteLoginPermissions(permissions uint32) error      { return nil }
func (client *b296) WriteFriendsList(userIds []uint32) error             { return nil }
func (client *b296) WriteProtocolNegotiation(version int32) error        { return nil }
func (client *b296) WriteTitleUpdate(update TitleUpdate) error           { return nil }
func (client *b296) WriteMonitor() error                                 { return nil }
func (client *b296) WriteMatchPlayerSkipped(slotId int32) error          { return nil }
func (client *b296) WriteRestart(retryMs int32) error                    { return nil }
func (client *b296) WriteInvite(message Message) error                   { return nil }
func (client *b296) WriteChannelInfoComplete() error                     { return nil }
func (client *b296) WriteMatchChangePassword(password string) error      { return nil }
func (client *b296) WriteSilenceInfo(timeRemaining int32) error          { return nil }
func (client *b296) WriteUserSilenced(userId uint32) error               { return nil }
func (client *b296) WriteUserDMsBlocked(targetName string) error         { return nil }
func (client *b296) WriteTargetIsSilenced(targetName string) error       { return nil }
func (client *b296) WriteVersionUpdateForced() error                     { return nil }
func (client *b296) WriteSwitchServer(target int32) error                { return nil }
func (client *b296) WriteAccountRestricted() error                       { return nil }
func (client *b296) WriteRTX(message string) error                       { return nil }
func (client *b296) WriteMatchAbort() error                              { return nil }
func (client *b296) WriteSwitchTournamentServer(ip string) error         { return nil }
