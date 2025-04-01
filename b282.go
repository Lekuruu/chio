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
	supportedPackets []uint16
	protocolVersion  int
	slotSize         int
}

func (client *b282) WritePacket(stream io.Writer, packetId uint16, data []byte) error {
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

	_, err = stream.Write(writer.Bytes())
	return err
}

func (client *b282) ReadPacket(stream io.Reader) (packet *BanchoPacket, err error) {
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

	length, err := readInt32(stream)
	if err != nil {
		return nil, err
	}

	compressedData := make([]byte, length)
	n, err := stream.Read(compressedData)
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

func (client *b282) ProtocolVersion() int {
	return client.protocolVersion
}

func (client *b282) OverrideProtocolVersion(version int) {
	client.protocolVersion = version
}

func (client *b282) MatchSlotSize() int {
	return client.slotSize
}

func (client *b282) OverrideMatchSlotSize(amount int) {
	client.slotSize = amount
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

func (client *b282) WriteLoginReply(stream io.Writer, reply int32) error {
	writer := bytes.NewBuffer([]byte{})
	writeInt32(writer, reply)
	return client.WritePacket(stream, BanchoLoginReply, writer.Bytes())
}

func (client *b282) WriteMessage(stream io.Writer, message Message) error {
	if message.Target != "#osu" {
		// Private messages & channels have not been implemented yet
		return nil
	}

	writer := bytes.NewBuffer([]byte{})
	writeString(writer, message.Sender)
	writeString(writer, message.Content)
	return client.WritePacket(stream, BanchoSendMessage, writer.Bytes())
}

func (client *b282) WritePing(stream io.Writer) error {
	return client.WritePacket(stream, BanchoPing, []byte{})
}

func (client *b282) WriteIrcChangeUsername(stream io.Writer, oldName string, newName string) error {
	writer := bytes.NewBuffer([]byte{})
	writeString(writer, fmt.Sprintf("%s>>>>%s", oldName, newName))
	return client.WritePacket(stream, BanchoHandleIrcChangeUsername, writer.Bytes())
}

func (client *b282) WriteUserStats(stream io.Writer, info UserInfo) error {
	writer := bytes.NewBuffer([]byte{})

	if info.Presence.IsIrc {
		writeString(writer, info.Name)
		return client.WritePacket(stream, BanchoHandleIrcJoin, writer.Bytes())
	}

	client.WriteStats(writer, info)
	return client.WritePacket(stream, BanchoHandleOsuUpdate, writer.Bytes())
}

func (client *b282) WriteUserQuit(stream io.Writer, quit UserQuit) error {
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

func (client *b282) WriteSpectatorJoined(stream io.Writer, userId int32) error {
	writer := bytes.NewBuffer([]byte{})
	writeInt32(writer, userId)
	return client.WritePacket(stream, BanchoSpectatorJoined, writer.Bytes())
}

func (client *b282) WriteSpectatorLeft(stream io.Writer, userId int32) error {
	writer := bytes.NewBuffer([]byte{})
	writeInt32(writer, userId)
	return client.WritePacket(stream, BanchoSpectatorLeft, writer.Bytes())
}

func (client *b282) WriteSpectateFrames(stream io.Writer, bundle ReplayFrameBundle) error {
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
	return client.WritePacket(stream, BanchoSpectateFrames, writer.Bytes())
}

func (client *b282) WriteVersionUpdate(stream io.Writer) error {
	return client.WritePacket(stream, BanchoVersionUpdate, []byte{})
}

func (client *b282) WriteSpectatorCantSpectate(stream io.Writer, userId int32) error {
	writer := bytes.NewBuffer([]byte{})
	writeInt32(writer, userId)
	return client.WritePacket(stream, BanchoSpectatorCantSpectate, writer.Bytes())
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
	writeString(writer, info.Presence.Location())
	return nil
}

// Redirect UserPresence packets to UserStats
func (client *b282) WriteUserPresence(stream io.Writer, info UserInfo) error {
	return client.WriteUserStats(stream, info)
}

func (client *b282) WriteUserPresenceSingle(stream io.Writer, info UserInfo) error {
	return client.WriteUserPresence(stream, info)
}

func (client *b282) WriteUserPresenceBundle(stream io.Writer, infos []UserInfo) error {
	for _, info := range infos {
		err := client.WriteUserPresence(stream, info)
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

func (client *b282) WriteGetAttention(stream io.Writer) error                        { return nil }
func (client *b282) WriteAnnouncement(stream io.Writer, message string) error        { return nil }
func (client *b282) WriteMatchUpdate(stream io.Writer, match Match) error            { return nil }
func (client *b282) WriteMatchNew(stream io.Writer, match Match) error               { return nil }
func (client *b282) WriteMatchDisband(stream io.Writer, matchId int32) error         { return nil }
func (client *b282) WriteLobbyJoin(stream io.Writer, userId int32) error             { return nil }
func (client *b282) WriteLobbyPart(stream io.Writer, userId int32) error             { return nil }
func (client *b282) WriteMatchJoinSuccess(stream io.Writer, match Match) error       { return nil }
func (client *b282) WriteMatchJoinFail(stream io.Writer) error                       { return nil }
func (client *b282) WriteFellowSpectatorJoined(stream io.Writer, userId int32) error { return nil }
func (client *b282) WriteFellowSpectatorLeft(stream io.Writer, userId int32) error   { return nil }
func (client *b282) WriteMatchStart(stream io.Writer, match Match) error             { return nil }
func (client *b282) WriteMatchScoreUpdate(stream io.Writer, frame ScoreFrame) error  { return nil }
func (client *b282) WriteMatchTransferHost(stream io.Writer) error                   { return nil }
func (client *b282) WriteMatchAllPlayersLoaded(stream io.Writer) error               { return nil }
func (client *b282) WriteMatchPlayerFailed(stream io.Writer, slotId uint32) error    { return nil }
func (client *b282) WriteMatchComplete(stream io.Writer) error                       { return nil }
func (client *b282) WriteMatchSkip(stream io.Writer) error                           { return nil }
func (client *b282) WriteUnauthorized(stream io.Writer) error                        { return nil }
func (client *b282) WriteChannelJoinSuccess(stream io.Writer, channel string) error  { return nil }
func (client *b282) WriteChannelRevoked(stream io.Writer, channel string) error      { return nil }
func (client *b282) WriteChannelAvailable(stream io.Writer, channel Channel) error   { return nil }
func (client *b282) WriteChannelAvailableAutojoin(stream io.Writer, channel Channel) error {
	return nil // what the fuck golang
}
func (client *b282) WriteBeatmapInfoReply(stream io.Writer, reply BeatmapInfoReply) error { return nil }
func (client *b282) WriteLoginPermissions(stream io.Writer, permissions uint32) error     { return nil }
func (client *b282) WriteFriendsList(stream io.Writer, userIds []int32) error             { return nil }
func (client *b282) WriteProtocolNegotiation(stream io.Writer, version int32) error       { return nil }
func (client *b282) WriteTitleUpdate(stream io.Writer, update TitleUpdate) error          { return nil }
func (client *b282) WriteMonitor(stream io.Writer) error                                  { return nil }
func (client *b282) WriteMatchPlayerSkipped(stream io.Writer, slotId int32) error         { return nil }
func (client *b282) WriteRestart(stream io.Writer, retryMs int32) error                   { return nil }
func (client *b282) WriteInvite(stream io.Writer, message Message) error                  { return nil }
func (client *b282) WriteChannelInfoComplete(stream io.Writer) error                      { return nil }
func (client *b282) WriteMatchChangePassword(stream io.Writer, password string) error     { return nil }
func (client *b282) WriteSilenceInfo(stream io.Writer, timeRemaining int32) error         { return nil }
func (client *b282) WriteUserSilenced(stream io.Writer, userId uint32) error              { return nil }
func (client *b282) WriteUserDMsBlocked(stream io.Writer, targetName string) error        { return nil }
func (client *b282) WriteTargetIsSilenced(stream io.Writer, targetName string) error      { return nil }
func (client *b282) WriteVersionUpdateForced(stream io.Writer) error                      { return nil }
func (client *b282) WriteSwitchServer(stream io.Writer, target int32) error               { return nil }
func (client *b282) WriteAccountRestricted(stream io.Writer) error                        { return nil }
func (client *b282) WriteRTX(stream io.Writer, message string) error                      { return nil }
func (client *b282) WriteMatchAbort(stream io.Writer) error                               { return nil }
func (client *b282) WriteSwitchTournamentServer(stream io.Writer, ip string) error        { return nil }
