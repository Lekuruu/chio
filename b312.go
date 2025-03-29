package chio

import (
	"bytes"
	"fmt"
	"io"
)

// b312 adds the match start & update packets, as well
// as the "InProgress" boolean inside the match struct
type b312 struct {
	*b298
}

func (client *b312) WritePacket(stream io.Writer, packetId uint16, data []byte) error {
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

func (client *b312) ReadPacket(stream io.Reader) (packet *BanchoPacket, err error) {
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
		OsuMatchReady,
		OsuMatchLock,
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

func (client *b312) ReadPacketType(packetId uint16, reader io.Reader) (any, error) {
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
	case OsuMatchScoreUpdate:
		return client.ReadScoreFrame(reader)
	default:
		return nil, nil
	}
}

/* New Packets */

func (client *b312) WriteMatchStart(stream io.Writer, match Match) error {
	return client.WritePacket(stream, BanchoMatchStart, []byte{})
}

func (client *b312) WriteMatchScoreUpdate(stream io.Writer, frame ScoreFrame) error {
	writer := bytes.NewBuffer([]byte{})
	client.WriteScoreFrame(writer, frame)
	return client.WritePacket(stream, BanchoMatchScoreUpdate, writer.Bytes())
}

func (client *b312) WriteMatchUpdate(stream io.Writer, match Match) error {
	writer := bytes.NewBuffer([]byte{})
	client.WriteMatch(writer, match)
	return client.WritePacket(stream, BanchoMatchUpdate, writer.Bytes())
}

func (client *b312) WriteMatchNew(stream io.Writer, match Match) error {
	writer := bytes.NewBuffer([]byte{})
	client.WriteMatch(writer, match)
	return client.WritePacket(stream, BanchoMatchNew, writer.Bytes())
}

func (client *b312) WriteMatchJoinSuccess(stream io.Writer, match Match) error {
	writer := bytes.NewBuffer([]byte{})
	client.WriteMatch(writer, match)
	return client.WritePacket(stream, BanchoMatchJoinSuccess, writer.Bytes())
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

func (client *b312) ReadMatch(reader io.Reader) (*Match, error) {
	var err error
	errors := NewErrorCollection()
	match := &Match{}

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
