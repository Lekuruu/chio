package chio

import (
	"bytes"
	"fmt"
	"io"
)

// b282 is the initial implementation of the bancho protocol.
// Every following version will be based on it.
type b282 struct {
	BanchoIO
	stream           io.ReadWriteCloser
	supportedPackets []uint16
}

func (client *b282) Write(p []byte) (n int, err error) {
	return client.stream.Write(p)
}

func (client *b282) Read(p []byte) (n int, err error) {
	return client.stream.Read(p)
}

func (client *b282) Close() error {
	return client.stream.Close()
}

func (client *b282) Clone() BanchoIO {
	return &b282{}
}

func (client *b282) GetStream() io.ReadWriteCloser {
	return client.stream
}

func (client *b282) SetStream(stream io.ReadWriteCloser) {
	client.stream = stream
}

func (client *b282) WritePacket(packetId uint16, data []byte) error {
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

func (client *b282) ReadPacket() (packet *BanchoPacket, err error) {
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

func (client *b282) SupportedPackets() []uint16 {
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
	}
	return client.supportedPackets
}

func (client *b282) ImplementsPacket(packetId uint16) bool {
	for _, id := range client.SupportedPackets() {
		if id == packetId {
			return true
		}
	}
	return false
}

func (client *b282) OverrideMatchSlotSize(amount int) {
	// Multiplayer is not supported in this version
}

func (client *b282) ConvertInputPacketId(packetId uint16) uint16 {
	if packetId == 11 {
		// "IrcJoin" packet
		return BanchoHandleIrcJoin
	}
	if packetId > 11 && packetId <= 45 {
		packetId -= 1
	}
	if packetId > 50 {
		packetId -= 1
	}
	return packetId
}

func (client *b282) ConvertOutputPacketId(packetId uint16) uint16 {
	if packetId == BanchoHandleIrcJoin {
		// "IrcJoin" packet
		return 11
	}
	if packetId >= 11 && packetId < 45 {
		return packetId + 1
	}
	if packetId > 50 {
		packetId += 1
	}
	return packetId
}

func (client *b282) ReadPacketType(packetId uint16, reader io.Reader) (any, error) {
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

func (client *b282) WriteLoginReply(reply int32) error {
	writer := bytes.NewBuffer([]byte{})
	writeInt32(writer, reply)
	return client.WritePacket(BanchoLoginReply, writer.Bytes())
}

func (client *b282) WriteMessage(message Message) error {
	if message.Target != "#osu" {
		// Private messages & channels have not been implemented yet
		return nil
	}

	writer := bytes.NewBuffer([]byte{})
	writeString(writer, message.Sender)
	writeString(writer, message.Content)
	return client.WritePacket(BanchoSendMessage, writer.Bytes())
}

func (client *b282) WritePing() error {
	return client.WritePacket(BanchoPing, []byte{})
}

func (client *b282) WriteIrcChangeUsername(oldName string, newName string) error {
	writer := bytes.NewBuffer([]byte{})
	writeString(writer, fmt.Sprintf("%s>>>>%s", oldName, newName))
	return client.WritePacket(BanchoHandleIrcChangeUsername, writer.Bytes())
}

func (client *b282) WriteUserStats(info UserInfo) error {
	writer := bytes.NewBuffer([]byte{})

	if info.Presence.IsIrc {
		writeString(writer, info.Name)
		return client.WritePacket(BanchoHandleIrcJoin, writer.Bytes())
	}

	client.WriteStats(writer, info)
	return client.WritePacket(BanchoHandleOsuUpdate, writer.Bytes())
}

func (client *b282) WriteUserQuit(quit UserQuit) error {
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

func (client *b282) WriteSpectatorJoined(userId int32) error {
	writer := bytes.NewBuffer([]byte{})
	writeInt32(writer, userId)
	return client.WritePacket(BanchoSpectatorJoined, writer.Bytes())
}

func (client *b282) WriteSpectatorLeft(userId int32) error {
	writer := bytes.NewBuffer([]byte{})
	writeInt32(writer, userId)
	return client.WritePacket(BanchoSpectatorLeft, writer.Bytes())
}

func (client *b282) WriteSpectateFrames(bundle ReplayFrameBundle) error {
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

func (client *b282) WriteVersionUpdate() error {
	return client.WritePacket(BanchoVersionUpdate, []byte{})
}

func (client *b282) WriteSpectatorCantSpectate(userId int32) error {
	writer := bytes.NewBuffer([]byte{})
	writeInt32(writer, userId)
	return client.WritePacket(BanchoSpectatorCantSpectate, writer.Bytes())
}

func (client *b282) WriteStatus(writer io.Writer, status *UserStatus) error {
	// Convert action enum
	action := status.Action

	if status.UpdateStats {
		// This will make the client update the user's stats
		// It will not be present in later versions
		action = StatusStatsUpdate
	}

	writeUint8(writer, action)

	if action != StatusUnknown {
		writeString(writer, status.Text)
		writeString(writer, status.BeatmapChecksum)
		writeUint16(writer, uint16(status.Mods))
	}

	return nil
}

func (client *b282) WriteStats(writer io.Writer, info UserInfo) error {
	writeInt32(writer, info.Id)
	writeString(writer, info.Name)
	writeUint64(writer, info.Stats.Rscore)
	writeFloat64(writer, info.Stats.Accuracy)
	writeInt32(writer, info.Stats.Playcount)
	writeUint64(writer, info.Stats.Tscore)
	writeInt32(writer, info.Stats.Rank)
	writeString(writer, info.AvatarFilename())
	client.WriteStatus(writer, info.Status)
	writeUint8(writer, uint8(info.Presence.Timezone+24))
	writeString(writer, info.Presence.City)
	return nil
}

// Redirect UserPresence packets to UserStats
func (client *b282) WriteUserPresence(info UserInfo) error {
	return client.WriteUserStats(info)
}

func (client *b282) WriteUserPresenceSingle(info UserInfo) error {
	return client.WriteUserPresence(info)
}

func (client *b282) WriteUserPresenceBundle(infos []UserInfo) error {
	for _, info := range infos {
		err := client.WriteUserPresence(info)
		if err != nil {
			return err
		}
	}
	return nil
}

func (client *b282) ReadStatus(reader io.Reader) (*UserStatus, error) {
	var err error
	errors := NewErrorCollection()
	status := &UserStatus{}
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

func (client *b282) ReadMessage(reader io.Reader) (*Message, error) {
	var err error
	message := &Message{}
	message.Content, err = readString(reader)
	if err != nil {
		return nil, err
	}

	// Private messages & channels have not been implemented yet
	message.Target = "#osu"
	message.Sender = ""

	return message, nil
}

func (client *b282) ReadFrameBundle(reader io.Reader) (*ReplayFrameBundle, error) {
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

	return &ReplayFrameBundle{Frames: frames, Action: action}, nil
}

func (client *b282) ReadReplayFrame(reader io.Reader) (*ReplayFrame, error) {
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

/* Unsupported Packets */

func (client *b282) WriteGetAttention() error                            { return nil }
func (client *b282) WriteAnnouncement(message string) error              { return nil }
func (client *b282) WriteMatchUpdate(match Match) error                  { return nil }
func (client *b282) WriteMatchNew(match Match) error                     { return nil }
func (client *b282) WriteMatchDisband(matchId int32) error               { return nil }
func (client *b282) WriteLobbyJoin(userId int32) error                   { return nil }
func (client *b282) WriteLobbyPart(userId int32) error                   { return nil }
func (client *b282) WriteMatchJoinSuccess(match Match) error             { return nil }
func (client *b282) WriteMatchJoinFail() error                           { return nil }
func (client *b282) WriteFellowSpectatorJoined(userId int32) error       { return nil }
func (client *b282) WriteFellowSpectatorLeft(userId int32) error         { return nil }
func (client *b282) WriteMatchStart(match Match) error                   { return nil }
func (client *b282) WriteMatchScoreUpdate(frame ScoreFrame) error        { return nil }
func (client *b282) WriteMatchTransferHost() error                       { return nil }
func (client *b282) WriteMatchAllPlayersLoaded() error                   { return nil }
func (client *b282) WriteMatchPlayerFailed(slotId uint32) error          { return nil }
func (client *b282) WriteMatchComplete() error                           { return nil }
func (client *b282) WriteMatchSkip() error                               { return nil }
func (client *b282) WriteUnauthorized() error                            { return nil }
func (client *b282) WriteChannelJoinSuccess(channel string) error        { return nil }
func (client *b282) WriteChannelRevoked(channel string) error            { return nil }
func (client *b282) WriteChannelAvailable(channel Channel) error         { return nil }
func (client *b282) WriteChannelAvailableAutojoin(channel Channel) error { return nil }
func (client *b282) WriteBeatmapInfoReply(reply BeatmapInfoReply) error  { return nil }
func (client *b282) WriteLoginPermissions(permissions uint32) error      { return nil }
func (client *b282) WriteFriendsList(userIds []uint32) error             { return nil }
func (client *b282) WriteProtocolNegotiation(version int32) error        { return nil }
func (client *b282) WriteTitleUpdate(update TitleUpdate) error           { return nil }
func (client *b282) WriteMonitor() error                                 { return nil }
func (client *b282) WriteMatchPlayerSkipped(slotId int32) error          { return nil }
func (client *b282) WriteRestart(retryMs int32) error                    { return nil }
func (client *b282) WriteInvite(message Message) error                   { return nil }
func (client *b282) WriteChannelInfoComplete() error                     { return nil }
func (client *b282) WriteMatchChangePassword(password string) error      { return nil }
func (client *b282) WriteSilenceInfo(timeRemaining int32) error          { return nil }
func (client *b282) WriteUserSilenced(userId uint32) error               { return nil }
func (client *b282) WriteUserDMsBlocked(targetName string) error         { return nil }
func (client *b282) WriteTargetIsSilenced(targetName string) error       { return nil }
func (client *b282) WriteVersionUpdateForced() error                     { return nil }
func (client *b282) WriteSwitchServer(target int32) error                { return nil }
func (client *b282) WriteAccountRestricted() error                       { return nil }
func (client *b282) WriteRTX(message string) error                       { return nil }
func (client *b282) WriteMatchAbort() error                              { return nil }
func (client *b282) WriteSwitchTournamentServer(ip string) error         { return nil }
