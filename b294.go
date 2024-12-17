package chio

import (
	"bytes"
	"fmt"
	"io"
)

type b294 struct {
	BanchoIO
	stream           io.ReadWriteCloser
	supportedPackets []uint16
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
	return &b294{}
}

func (client *b294) GetStream() io.ReadWriteCloser {
	return client.stream
}

func (client *b294) SetStream(stream io.ReadWriteCloser) {
	client.stream = stream
}

func (client *b294) WritePacket(packetId uint16, data []byte) error {
	// Convert packetId back for the client
	packetId = client.convertOutputPacketId(packetId)
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
	packet.Id = client.convertInputPacketId(packet.Id)

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

	packet.Data, err = client.readPacketType(packet.Id, bytes.NewReader(data))
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

func (client *b294) WriteLoginReply(reply int32) error {
	writer := bytes.NewBuffer([]byte{})
	writeInt32(writer, reply)
	return client.WritePacket(BanchoLoginReply, writer.Bytes())
}

func (client *b294) WriteMessage(message Message) error {
	if message.Sender != "#osu" {
		// Private messages & channels have not been implemented yet
		return nil
	}

	writer := bytes.NewBuffer([]byte{})
	writeString(writer, message.Sender)
	writeString(writer, message.Content)
	return client.WritePacket(BanchoSendMessage, writer.Bytes())
}

func (client *b294) WritePing() error {
	return client.WritePacket(BanchoPing, []byte{})
}

func (client *b294) WriteIrcChangeUsername(oldName string, newName string) error {
	writer := bytes.NewBuffer([]byte{})
	writeString(writer, fmt.Sprintf("%s>>>>%s", oldName, newName))
	return client.WritePacket(BanchoHandleIrcChangeUsername, writer.Bytes())
}

func (client *b294) WriteUserStats(info UserInfo) error {
	writer := bytes.NewBuffer([]byte{})

	if info.Presence.IsIrc {
		// Write "IrcJoin" packet
		writeString(writer, info.Name)
		return client.WritePacket(0xFFFF, writer.Bytes())
	}

	client.writeStats(writer, info)
	return client.WritePacket(BanchoHandleOsuUpdate, writer.Bytes())
}

func (client *b294) WriteUserQuit(quit UserQuit) error {
	writer := bytes.NewBuffer([]byte{})

	if quit.Info.Presence.IsIrc && quit.QuitState != QuitStateIrcRemaining {
		writeString(writer, quit.Info.Name)
		return client.WritePacket(BanchoHandleIrcQuit, writer.Bytes())
	}

	if quit.QuitState == QuitStateOsuRemaining {
		return nil
	}

	client.writeStats(writer, *quit.Info)
	return client.WritePacket(BanchoHandleOsuQuit, writer.Bytes())
}

func (client *b294) WriteSpectatorJoined(userId uint32) error {
	writer := bytes.NewBuffer([]byte{})
	writeUint32(writer, userId)
	return client.WritePacket(BanchoSpectatorJoined, writer.Bytes())
}

func (client *b294) WriteSpectatorLeft(userId uint32) error {
	writer := bytes.NewBuffer([]byte{})
	writeUint32(writer, userId)
	return client.WritePacket(BanchoSpectatorLeft, writer.Bytes())
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
	return client.WritePacket(BanchoSpectateFrames, writer.Bytes())
}

func (client *b294) WriteVersionUpdate() error {
	return client.WritePacket(BanchoVersionUpdate, []byte{})
}

func (client *b294) WriteSpectatorCantSpectate(userId uint32) error {
	writer := bytes.NewBuffer([]byte{})
	writeUint32(writer, userId)
	return client.WritePacket(BanchoSpectatorCantSpectate, writer.Bytes())
}

func (client *b294) convertInputPacketId(packetId uint16) uint16 {
	if packetId == 11 {
		// "IrcJoin" packet
		return 0xFFFF
	}
	if packetId > 11 {
		return packetId - 1
	}
	return packetId
}

func (client *b294) convertOutputPacketId(packetId uint16) uint16 {
	if packetId == 0xFFFF {
		// "IrcJoin" packet
		return 11
	}
	if packetId >= 11 {
		return packetId + 1
	}
	return packetId
}

func (client *b294) readPacketType(packetId uint16, reader io.Reader) (any, error) {
	switch packetId {
	case OsuSendUserStatus:
		return client.readStatus(reader)
	case OsuSendIrcMessage:
		return client.readMessage(reader)
	case OsuStartSpectating:
		return readUint32(reader)
	case OsuSpectateFrames:
		return client.readFrameBundle(reader)
	case OsuErrorReport:
		return readString(reader)
	case OsuSendIrcMessagePrivate:
		return client.readMessagePrivate(reader)
	default:
		return nil, nil
	}
}

func (client *b294) readStatus(reader io.Reader) (any, error) {
	var err error
	errors := NewErrorCollection()
	status := UserStatus{}
	status.Action, err = readUint8(reader)
	errors.Add(err)

	if status.Action != StatusUnknown {
		status.Text, err = readString(reader)
		errors.Add(err)
		status.BeatmapChecksum, err = readString(reader)
		errors.Add(err)
		mods, err := readUint16(reader)
		errors.Add(err)
		status.Mods = uint32(mods)
	}

	return status, errors.Next()
}

func (client *b294) readMessage(reader io.Reader) (*Message, error) {
	content, err := readString(reader)
	if err != nil {
		return nil, err
	}
	return &Message{Content: content, Target: "#osu"}, nil
}

func (client *b294) readMessagePrivate(reader io.Reader) (*Message, error) {
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

func (client *b294) readFrameBundle(reader io.Reader) (*ReplayFrameBundle, error) {
	count, err := readUint16(reader)
	if err != nil {
		return nil, err
	}

	frames := make([]*ReplayFrame, count)
	for i := 0; i < int(count); i++ {
		frame, err := client.readReplayFrame(reader)
		if err != nil {
			return nil, err
		}
		frames[i] = frame
	}

	action, err := readUint8(reader)
	if err != nil {
		return nil, err
	}

	return &ReplayFrameBundle{Frames: frames, Action: action}, nil
}

func (client *b294) readReplayFrame(reader io.Reader) (*ReplayFrame, error) {
	var err error
	errors := NewErrorCollection()
	frame := &ReplayFrame{}
	mouseLeft, err := readBoolean(reader)
	errors.Add(err)
	mouseRight, err := readBoolean(reader)
	errors.Add(err)
	frame.MouseX, err = readFloat32(reader)
	errors.Add(err)
	frame.MouseY, err = readFloat32(reader)
	errors.Add(err)
	frame.Time, err = readInt32(reader)
	errors.Add(err)

	frame.ButtonState = 0
	frame.LegacyByte = 0

	if mouseLeft {
		frame.ButtonState |= ButtonStateLeft1
	}
	if mouseRight {
		frame.ButtonState |= ButtonStateRight1
	}

	return frame, errors.Next()
}

func (client *b294) writeStatus(writer io.Writer, status *UserStatus) error {
	// Convert action enum
	action := status.Action

	if action > StatusSubmitting {
		// Actions after "StatusSubmitting" are not supported
		action = StatusUnknown
	}

	if status.UpdateStats {
		// This will make the client update the user's stats
		// It will not be present in later versions
		action = StatusStatsUpdate
	}

	writeUint8(writer, action)
	writeString(writer, status.Text)
	writeString(writer, status.BeatmapChecksum)
	writeUint32(writer, status.Mods)
	return nil
}

func (client *b294) writeStats(writer io.Writer, info UserInfo) error {
	writeInt32(writer, info.Id)
	writeString(writer, info.Name)
	writeUint64(writer, info.Stats.Rscore)
	writeFloat64(writer, info.Stats.Accuracy)
	writeInt32(writer, info.Stats.Playcount)
	writeUint64(writer, info.Stats.Tscore)
	writeInt32(writer, info.Stats.Rank)
	writeString(writer, fmt.Sprintf("%d", info.Id))
	client.writeStatus(writer, info.Status)
	writeUint8(writer, uint8(info.Presence.Timezone+24))
	writeString(writer, info.Presence.City)
	return nil
}

// Redirect UserPresence packets to UserStats
func (client *b294) WriteUserPresence(info UserInfo) error {
	return client.WriteUserStats(info)
}

func (client *b294) WriteUserPresenceSingle(info UserInfo) error {
	return client.WriteUserPresence(info)
}

func (client *b294) WriteUserPresenceBundle(infos []UserInfo) error {
	for _, info := range infos {
		err := client.WriteUserPresence(info)
		if err != nil {
			return err
		}
	}
	return nil
}

func (client *b294) WriteGetAttention() error {
	return client.WritePacket(BanchoGetAttention, []byte{})
}

func (client *b294) WriteAnnouncement(message string) error {
	writer := bytes.NewBuffer([]byte{})
	writeString(writer, message)
	return client.WritePacket(BanchoAnnounce, writer.Bytes())
}

// Unsupported Packets
func (client *b294) WriteMatchUpdate(match Match) error                  { return nil }
func (client *b294) WriteMatchNew(match Match) error                     { return nil }
func (client *b294) WriteMatchDisband(matchId uint32) error              { return nil }
func (client *b294) WriteLobbyJoin(userId uint32) error                  { return nil }
func (client *b294) WriteLobbyPart(userId uint32) error                  { return nil }
func (client *b294) WriteMatchJoinSuccess(match Match) error             { return nil }
func (client *b294) WriteMatchJoinFail() error                           { return nil }
func (client *b294) WriteFellowSpectatorJoined(userId uint32) error      { return nil }
func (client *b294) WriteFellowSpectatorLeft(userId uint32) error        { return nil }
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
