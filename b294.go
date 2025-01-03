package chio

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

// b294 implements private messages, as well as score frames in spectating
type b294 struct {
	BanchoIO
	stream           io.ReadWriteCloser
	supportedPackets []uint16
	previous         *b291
}

func (client *b294) Write(p []byte) (n int, err error) {
	return client.stream.Write(p)
}

func (client *b294) Read(p []byte) (n int, err error) {
	return client.stream.Read(p)
}

func (client *b294) Close() error {
	return client.stream.Close()
}

func (client *b294) Clone() BanchoIO {
	previous := &b291{}
	clone := previous.Clone()
	return &b294{previous: clone.(*b291)}
}

func (client *b294) GetStream() io.ReadWriteCloser {
	return client.stream
}

func (client *b294) SetStream(stream io.ReadWriteCloser) {
	client.stream = stream
	client.previous.SetStream(stream)
}

func (client *b294) WritePacket(packetId uint16, data []byte) error {
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

func (client *b294) ReadPacket() (packet *BanchoPacket, err error) {
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

func (client *b294) SupportedPackets() []uint16 {
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

func (client *b294) ImplementsPacket(packetId uint16) bool {
	for _, id := range client.SupportedPackets() {
		if id == packetId {
			return true
		}
	}
	return false
}

func (client *b294) OverrideMatchSlotSize(amount int) {
	// Multiplayer is not supported in this version
}

func (client *b294) ConvertInputPacketId(packetId uint16) uint16 {
	if packetId == 11 {
		// "IrcJoin" packet
		return BanchoHandleIrcJoin
	}
	if packetId > 11 {
		return packetId - 1
	}
	return packetId
}

func (client *b294) ConvertOutputPacketId(packetId uint16) uint16 {
	if packetId == BanchoHandleIrcJoin {
		// "IrcJoin" packet
		return 11
	}
	if packetId >= 11 {
		return packetId + 1
	}
	return packetId
}

func (client *b294) ReadPacketType(packetId uint16, reader io.Reader) (any, error) {
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

func (client *b294) WriteMessage(message Message) error {
	isChannel := strings.HasPrefix(message.Target, "#")

	if isChannel && message.Target != "#osu" {
		// Channel selection was not implemented yet
		return nil
	}

	writer := bytes.NewBuffer([]byte{})
	writeString(writer, message.Sender)
	writeString(writer, message.Content)
	writeBoolean(writer, message.Target != "#osu") // IsPrivate
	return client.WritePacket(BanchoSendMessage, writer.Bytes())
}

func (client *b294) ReadMessage(reader io.Reader) (*Message, error) {
	content, err := readString(reader)
	if err != nil {
		return nil, err
	}
	return &Message{Content: content, Target: "#osu"}, nil
}

func (client *b294) ReadMessagePrivate(reader io.Reader) (*Message, error) {
	var err error
	message := &Message{}
	message.Sender = ""

	message.Target, err = readString(reader)
	if err != nil {
		return nil, err
	}

	message.Content, err = readString(reader)
	if err != nil {
		return nil, err
	}

	isPrivate, err := readBoolean(reader)
	if err != nil {
		return nil, err
	}

	if !isPrivate {
		panic("expected private message")
	}

	return message, nil
}

func (client *b294) WriteSpectateFrames(bundle ReplayFrameBundle) error {
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

func (client *b294) WriteScoreFrame(writer io.Writer, frame ScoreFrame) error {
	writeString(writer, frame.Checksum())
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

func (client *b294) ReadFrameBundle(reader io.Reader) (*ReplayFrameBundle, error) {
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

func (client b294) ReadScoreFrame(reader io.Reader) (*ScoreFrame, error) {
	var err error
	errors := NewErrorCollection()
	frame := &ScoreFrame{}
	_, err = readString(reader) // Checksum
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
	return frame, errors.Next()
}

/* Inherited Packets */

func (client *b294) WriteVersionUpdate() error {
	return client.previous.WriteVersionUpdate()
}

func (client *b294) WriteSpectatorCantSpectate(userId int32) error {
	return client.previous.WriteSpectatorCantSpectate(userId)
}

func (client *b294) WriteLoginReply(reply int32) error {
	return client.previous.WriteLoginReply(reply)
}

func (client *b294) WritePing() error {
	return client.previous.WritePing()
}

func (client *b294) WriteIrcChangeUsername(oldName string, newName string) error {
	return client.previous.WriteIrcChangeUsername(oldName, newName)
}

func (client *b294) WriteUserStats(info UserInfo) error {
	return client.previous.WriteUserStats(info)
}

func (client *b294) WriteUserQuit(quit UserQuit) error {
	return client.previous.WriteUserQuit(quit)
}

func (client *b294) WriteSpectatorJoined(userId int32) error {
	return client.previous.WriteSpectatorJoined(userId)
}

func (client *b294) WriteSpectatorLeft(userId int32) error {
	return client.previous.WriteSpectatorLeft(userId)
}

func (client *b294) WriteStatus(writer io.Writer, status *UserStatus) error {
	return client.previous.WriteStatus(writer, status)
}

func (client *b294) WriteStats(writer io.Writer, info UserInfo) error {
	return client.previous.WriteStats(writer, info)
}

func (client *b294) WriteUserPresence(info UserInfo) error {
	return client.previous.WriteUserPresence(info)
}

func (client *b294) WriteUserPresenceSingle(info UserInfo) error {
	return client.previous.WriteUserPresenceSingle(info)
}

func (client *b294) WriteUserPresenceBundle(infos []UserInfo) error {
	return client.previous.WriteUserPresenceBundle(infos)
}

func (client *b294) WriteGetAttention() error {
	return client.previous.WriteGetAttention()
}

func (client *b294) WriteAnnouncement(message string) error {
	return client.previous.WriteAnnouncement(message)
}

func (client *b294) ReadStatus(reader io.Reader) (any, error) {
	return client.previous.ReadStatus(reader)
}

func (client *b294) ReadReplayFrame(reader io.Reader) (*ReplayFrame, error) {
	return client.previous.ReadReplayFrame(reader)
}

/* Unsupported Packets */

func (client *b294) WriteMatchUpdate(match Match) error                  { return nil }
func (client *b294) WriteMatchNew(match Match) error                     { return nil }
func (client *b294) WriteMatchDisband(matchId int32) error               { return nil }
func (client *b294) WriteLobbyJoin(userId int32) error                   { return nil }
func (client *b294) WriteLobbyPart(userId int32) error                   { return nil }
func (client *b294) WriteMatchJoinSuccess(match Match) error             { return nil }
func (client *b294) WriteMatchJoinFail() error                           { return nil }
func (client *b294) WriteFellowSpectatorJoined(userId int32) error       { return nil }
func (client *b294) WriteFellowSpectatorLeft(userId int32) error         { return nil }
func (client *b294) WriteMatchStart(match Match) error                   { return nil }
func (client *b294) WriteMatchScoreUpdate(frame ScoreFrame) error        { return nil }
func (client *b294) WriteMatchTransferHost() error                       { return nil }
func (client *b294) WriteMatchAllPlayersLoaded() error                   { return nil }
func (client *b294) WriteMatchPlayerFailed(slotId uint32) error          { return nil }
func (client *b294) WriteMatchComplete() error                           { return nil }
func (client *b294) WriteMatchSkip() error                               { return nil }
func (client *b294) WriteUnauthorized() error                            { return nil }
func (client *b294) WriteChannelJoinSuccess(channel string) error        { return nil }
func (client *b294) WriteChannelRevoked(channel string) error            { return nil }
func (client *b294) WriteChannelAvailable(channel Channel) error         { return nil }
func (client *b294) WriteChannelAvailableAutojoin(channel Channel) error { return nil }
func (client *b294) WriteBeatmapInfoReply(reply BeatmapInfoReply) error  { return nil }
func (client *b294) WriteLoginPermissions(permissions uint32) error      { return nil }
func (client *b294) WriteFriendsList(userIds []uint32) error             { return nil }
func (client *b294) WriteProtocolNegotiation(version int32) error        { return nil }
func (client *b294) WriteTitleUpdate(update TitleUpdate) error           { return nil }
func (client *b294) WriteMonitor() error                                 { return nil }
func (client *b294) WriteMatchPlayerSkipped(slotId int32) error          { return nil }
func (client *b294) WriteRestart(retryMs int32) error                    { return nil }
func (client *b294) WriteInvite(message Message) error                   { return nil }
func (client *b294) WriteChannelInfoComplete() error                     { return nil }
func (client *b294) WriteMatchChangePassword(password string) error      { return nil }
func (client *b294) WriteSilenceInfo(timeRemaining int32) error          { return nil }
func (client *b294) WriteUserSilenced(userId uint32) error               { return nil }
func (client *b294) WriteUserDMsBlocked(targetName string) error         { return nil }
func (client *b294) WriteTargetIsSilenced(targetName string) error       { return nil }
func (client *b294) WriteVersionUpdateForced() error                     { return nil }
func (client *b294) WriteSwitchServer(target int32) error                { return nil }
func (client *b294) WriteAccountRestricted() error                       { return nil }
func (client *b294) WriteRTX(message string) error                       { return nil }
func (client *b294) WriteMatchAbort() error                              { return nil }
func (client *b294) WriteSwitchTournamentServer(ip string) error         { return nil }
