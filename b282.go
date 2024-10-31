package chio

import (
	"bytes"
	"fmt"
	"io"
)

type b282 struct {
	BanchoIO
	Stream io.ReadWriteCloser
}

func (client *b282) Write(p []byte) (n int, err error) {
	return client.Stream.Write(p)
}

func (client *b282) Read(p []byte) (n int, err error) {
	return client.Stream.Read(p)
}

func (client *b282) Close() error {
	return client.Stream.Close()
}

func (client *b282) Clone() BanchoIO {
	return &b282{}
}

func (client *b282) GetStream() io.ReadWriteCloser {
	return client.Stream
}

func (client *b282) SetStream(stream io.ReadWriteCloser) {
	client.Stream = stream
}

func (client *b282) WritePacket(packetId uint16, data []byte) error {
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

func (client *b282) ReadPacket() (packet *BanchoPacket, err error) {
	packet = &BanchoPacket{}
	packet.PacketId, err = readUint16(client.Stream)
	if err != nil {
		return nil, err
	}

	// Convert packet ID to a usable value
	packet.PacketId = client.convertInputPacketId(packet.PacketId)

	if !client.ImplementsPacket(packet.PacketId) {
		return nil, nil
	}

	length, err := readInt32(client.Stream)
	if err != nil {
		return nil, err
	}

	compressedData := make([]byte, length)
	n, err := client.Stream.Read(compressedData)
	if err != nil {
		return nil, err
	}

	if n != int(length) {
		return nil, fmt.Errorf("expected %d bytes, got %d", length, n)
	}

	data := decompressData(compressedData)
	packet.Data, err = client.readPacketType(packet.PacketId, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	return packet, nil
}

var supportedPackets []uint16 = []uint16{
	OsuSendUserStatus,
	OsuSendIrcMessage,
	OsuExit,
	OsuRequestStatusUpdate,
	OsuPong,
	BanchoLoginReply,
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

func (client *b282) ImplementsPacket(packetId uint16) bool {
	for _, id := range supportedPackets {
		if id == packetId {
			return true
		}
	}
	return false
}

func (client *b282) WriteLoginReply(reply int32) error {
	writer := bytes.NewBuffer([]byte{})
	writeInt32(writer, reply)
	return client.WritePacket(BanchoLoginReply, writer.Bytes())
}

func (client *b282) WriteMessage(message Message) error {
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
	writeStats(writer, info)
	return client.WritePacket(BanchoHandleOsuUpdate, writer.Bytes())
}

func (client *b282) WriteUserQuit(quit UserQuit) error {
	writer := bytes.NewBuffer([]byte{})

	if quit.Info.Presence.IsIrc {
		writeString(writer, quit.Info.Name)
		return client.WritePacket(BanchoHandleIrcQuit, writer.Bytes())
	}

	writeStats(writer, *quit.Info)
	return client.WritePacket(BanchoHandleOsuQuit, writer.Bytes())
}

func (client *b282) WriteSpectatorJoined(userId uint32) error {
	writer := bytes.NewBuffer([]byte{})
	writeUint32(writer, userId)
	return client.WritePacket(BanchoSpectatorJoined, writer.Bytes())
}

func (client *b282) WriteSpectatorLeft(userId uint32) error {
	writer := bytes.NewBuffer([]byte{})
	writeUint32(writer, userId)
	return client.WritePacket(BanchoSpectatorLeft, writer.Bytes())
}

func (client *b282) WriteSpectateFrames(bundle ReplayFrameBundle) error {
	writer := bytes.NewBuffer([]byte{})
	writeUint16(writer, uint16(len(bundle.Frames)))

	for _, frame := range bundle.Frames {
		writeUint8(writer, frame.ButtonState)
		writeUint8(writer, frame.LegacyByte)
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

func (client *b282) WriteSpectatorCantSpectate(userId uint32) error {
	writer := bytes.NewBuffer([]byte{})
	writeUint32(writer, userId)
	return client.WritePacket(BanchoSpectatorCantSpectate, writer.Bytes())
}

func (client *b282) convertInputPacketId(packetId uint16) uint16 {
	if packetId == 6 {
		// "CommandError" packet
		return 0xFFFF
	}
	if packetId > 6 {
		return packetId - 1
	}
	return packetId
}

func (client *b282) convertOutputPacketId(packetId uint16) uint16 {
	if packetId == 0xFFFF {
		// "CommandError" packet
		return 6
	}
	if packetId >= 6 {
		return packetId + 1
	}
	return packetId
}

func (client *b282) readPacketType(packetId uint16, reader io.Reader) (any, error) {
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
	default:
		return nil, nil
	}
}

func (client *b282) readStatus(reader io.Reader) (any, error) {
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
		status.Mods, err = readUint32(reader)
		errors.Add(err)
	}

	if status.Action == StatusIdle && status.BeatmapChecksum != "" {
		// There is a bug where the client is playing but
		// didn't update the status correctly
		status.Action = StatusPlaying
	}

	return status, errors.Next()
}

func (client *b282) readMessage(reader io.Reader) (*Message, error) {
	var err error
	errors := NewErrorCollection()
	message := &Message{}
	message.Sender, err = readString(reader)
	errors.Add(err)
	message.Content, err = readString(reader)
	errors.Add(err)

	// Private messages & channels have not been implemented yet
	message.Target = "#osu"

	return message, errors.Next()
}

func (client *b282) readFrameBundle(reader io.Reader) (*ReplayFrameBundle, error) {
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

func (client *b282) readReplayFrame(reader io.Reader) (*ReplayFrame, error) {
	var err error
	errors := NewErrorCollection()
	frame := &ReplayFrame{}
	frame.ButtonState, err = readUint8(reader)
	errors.Add(err)
	frame.LegacyByte, err = readUint8(reader)
	errors.Add(err)
	frame.MouseX, err = readFloat32(reader)
	errors.Add(err)
	frame.MouseY, err = readFloat32(reader)
	errors.Add(err)
	frame.Time, err = readInt32(reader)
	errors.Add(err)
	return frame, errors.Next()
}

func writeStatus(writer io.Writer, status *UserStatus) error {
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

func writeStats(writer io.Writer, info UserInfo) error {
	writeInt32(writer, info.Id)
	writeString(writer, info.Name)
	writeUint64(writer, info.Stats.Rscore)
	writeFloat64(writer, info.Stats.Accuracy)
	writeInt32(writer, info.Stats.Playcount)
	writeUint64(writer, info.Stats.Tscore)
	writeInt32(writer, info.Stats.Rank)
	writeString(writer, fmt.Sprintf("%d_000.png", info.Id))
	writeStatus(writer, info.Status)
	writeUint8(writer, uint8(info.Presence.Timezone+24))
	writeString(writer, info.Presence.City)
	return nil
}

// Unsupported Packets
func (client *b282) WriteGetAttention() error                            { return nil }
func (client *b282) WriteAnnounce(message string) error                  { return nil }
func (client *b282) WriteMatchUpdate(match Match) error                  { return nil }
func (client *b282) WriteMatchNew(match Match) error                     { return nil }
func (client *b282) WriteMatchDisband(matchId uint32) error              { return nil }
func (client *b282) WriteLobbyJoin(userId uint32) error                  { return nil }
func (client *b282) WriteLobbyPart(userId uint32) error                  { return nil }
func (client *b282) WriteMatchJoinSuccess(match Match) error             { return nil }
func (client *b282) WriteMatchJoinFail() error                           { return nil }
func (client *b282) WriteFellowSpectatorJoined(userId uint32) error      { return nil }
func (client *b282) WriteFellowSpectatorLeft(userId uint32) error        { return nil }
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
func (client *b282) WriteUserPresence(presence UserInfo) error           { return nil }
func (client *b282) WriteRestart(retryMs int32) error                    { return nil }
func (client *b282) WriteInvite(message Message) error                   { return nil }
func (client *b282) WriteChannelInfoComplete() error                     { return nil }
func (client *b282) WriteMatchChangePassword(password string) error      { return nil }
func (client *b282) WriteSilenceInfo(timeRemaining int32) error          { return nil }
func (client *b282) WriteUserSilenced(userId uint32) error               { return nil }
func (client *b282) WriteUserPresenceSingle(info UserInfo) error         { return nil }
func (client *b282) WriteUserPresenceBundle(infos []UserInfo) error      { return nil }
func (client *b282) WriteUserDMsBlocked(targetName string) error         { return nil }
func (client *b282) WriteTargetIsSilenced(targetName string) error       { return nil }
func (client *b282) WriteVersionUpdateForced() error                     { return nil }
func (client *b282) WriteSwitchServer(t int32) error                     { return nil }
func (client *b282) WriteAccountRestricted() error                       { return nil }
func (client *b282) WriteRTX(message string) error                       { return nil }
func (client *b282) WriteMatchAbort() error                              { return nil }
func (client *b282) WriteSwitchTournamentServer(ip string) error         { return nil }
