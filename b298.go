package chio

import (
	"bytes"
	"fmt"
	"io"
)

// b298 adds a partial implementation of multiplayer, as well as fellow spectators
type b298 struct {
	*b296
}

func (client *b298) WritePacket(stream io.Writer, packetId uint16, data []byte) error {
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

func (client *b298) ReadPacket(stream io.Reader) (packet *BanchoPacket, err error) {
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
		OsuMatchReady,
		OsuMatchLock,
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
}

func (client *b298) ReadPacketType(packetId uint16, reader io.Reader) (any, error) {
	switch packetId {
	case OsuSendUserStatus:
		return client.ReadStatus(reader)
	case OsuSendIrcMessage:
		return client.ReadMessage(reader)
	case OsuSendIrcMessagePrivate:
		return client.ReadMessagePrivate(reader)
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
	default:
		return nil, nil
	}
}

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

func (client *b298) WriteMatchUpdate(stream io.Writer, match Match) error {
	writer := bytes.NewBuffer([]byte{})
	client.WriteMatch(writer, match)
	return client.WritePacket(stream, BanchoMatchUpdate, writer.Bytes())
}

func (client *b298) WriteMatchNew(stream io.Writer, match Match) error {
	writer := bytes.NewBuffer([]byte{})
	client.WriteMatch(writer, match)
	return client.WritePacket(stream, BanchoMatchNew, writer.Bytes())
}

func (client *b298) WriteMatchDisband(stream io.Writer, matchId int32) error {
	writer := bytes.NewBuffer([]byte{})
	writeInt32(writer, matchId)
	return client.WritePacket(stream, BanchoMatchDisband, writer.Bytes())
}

func (client *b298) WriteLobbyJoin(stream io.Writer, userId int32) error {
	writer := bytes.NewBuffer([]byte{})
	writeInt32(writer, userId)
	return client.WritePacket(stream, BanchoLobbyJoin, writer.Bytes())
}

func (client *b298) WriteLobbyPart(stream io.Writer, userId int32) error {
	writer := bytes.NewBuffer([]byte{})
	writeInt32(writer, userId)
	return client.WritePacket(stream, BanchoLobbyPart, writer.Bytes())
}

func (client *b298) WriteMatchJoinSuccess(stream io.Writer, match Match) error {
	writer := bytes.NewBuffer([]byte{})
	client.WriteMatch(writer, match)
	return client.WritePacket(stream, BanchoMatchJoinSuccess, writer.Bytes())
}

func (client *b298) WriteMatchJoinFail(stream io.Writer) error {
	return client.WritePacket(stream, BanchoMatchJoinFail, []byte{})
}

func (client *b298) WriteFellowSpectatorJoined(stream io.Writer, userId int32) error {
	writer := bytes.NewBuffer([]byte{})
	writeInt32(writer, userId)
	return client.WritePacket(stream, BanchoFellowSpectatorJoined, writer.Bytes())
}

func (client *b298) WriteFellowSpectatorLeft(stream io.Writer, userId int32) error {
	writer := bytes.NewBuffer([]byte{})
	writeInt32(writer, userId)
	return client.WritePacket(stream, BanchoFellowSpectatorLeft, writer.Bytes())
}

func (client *b298) ReadMatch(reader io.Reader) (*Match, error) {
	var err error
	errors := NewErrorCollection()
	match := &Match{}

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
		} else {
			slot.Status = SlotStatusLocked
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
