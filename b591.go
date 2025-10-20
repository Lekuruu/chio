package chio

import (
	"bytes"
	"fmt"
	"io"
)

// b591 implements password protected multiplayer matches.
type b591 struct {
	*b558
}

func (client *b591) ReadPacket(stream io.Reader) (packet *BanchoPacket, err error) {
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

	compressionEnabled, err := readBoolean(stream)
	if err != nil {
		return nil, err
	}

	length, err := readInt32(stream)
	if err != nil {
		return nil, err
	}

	data := make([]byte, length)
	n, err := stream.Read(data)
	if err != nil {
		return nil, err
	}

	if n != int(length) {
		return nil, fmt.Errorf("expected %d bytes, got %d", length, n)
	}

	if compressionEnabled {
		data, err = decompressData(data)
		if err != nil {
			return nil, err
		}
	}

	packet.Data, err = client.ReadPacketType(packet.Id, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	return packet, nil
}

func (client *b591) ReadPacketType(packetId uint16, reader io.Reader) (any, error) {
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
		return client.ReadMatchJoin(reader)
	case OsuMatchChangeSettings:
		return client.ReadMatch(reader)
	case OsuMatchChangeSlot:
		return readUint32(reader)
	case OsuMatchLock:
		return readUint32(reader)
	case OsuMatchScoreUpdate:
		return client.ReadScoreFrame(reader)
	case OsuMatchChangeMods:
		return readUint32(reader)
	case OsuChannelJoin:
		return readString(reader)
	case OsuChannelLeave:
		return readString(reader)
	case OsuBeatmapInfoRequest:
		return client.ReadBeatmapInfoRequest(reader)
	case OsuMatchTransferHost:
		return readInt32(reader)
	case OsuFriendsAdd:
		return readInt32(reader)
	case OsuFriendsRemove:
		return readInt32(reader)
	default:
		return nil, nil
	}
}

func (client *b591) ReadMatch(reader io.Reader) (*Match, error) {
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
	mods, err := readUint16(reader)
	match.Mods = uint32(mods)
	errors.Add(err)
	match.Name, err = readString(reader)
	errors.Add(err)
	match.Password, err = readString(reader)
	errors.Add(err)
	match.BeatmapText, err = readString(reader)
	errors.Add(err)
	match.BeatmapId, err = readInt32(reader)
	errors.Add(err)
	match.BeatmapChecksum, err = readString(reader)
	errors.Add(err)

	match.Slots = make([]*MatchSlot, client.slotSize)

	for i := 0; i < client.slotSize; i++ {
		slot := &MatchSlot{}
		slot.Status, err = readUint8(reader)
		slot.Team = SlotTeamNeutral
		slot.Mods = NoMod
		slot.UserId = -1
		match.Slots[i] = slot
		errors.Add(err)
	}

	if client.ProtocolVersion() >= 4 {
		for i := 0; i < client.slotSize; i++ {
			match.Slots[i].Team, err = readUint8(reader)
			errors.Add(err)
		}
	}

	for i := 0; i < client.slotSize; i++ {
		if !match.Slots[i].HasPlayer() {
			continue
		}

		match.Slots[i].UserId, err = readInt32(reader)
		errors.Add(err)
	}

	match.HostId, err = readInt32(reader)
	errors.Add(err)
	match.Mode, err = readUint8(reader)
	errors.Add(err)

	if client.ProtocolVersion() < 3 {
		return match, errors.Next()
	}

	match.ScoringType, err = readUint8(reader)
	errors.Add(err)
	match.TeamType, err = readUint8(reader)
	errors.Add(err)

	return match, errors.Next()
}

func (client *b591) ReadMatchJoin(reader io.Reader) (*MatchJoin, error) {
	var err error
	errors := NewErrorCollection()
	join := &MatchJoin{}

	join.MatchId, err = readInt32(reader)
	errors.Add(err)
	join.Password, err = readString(reader)
	errors.Add(err)

	return join, errors.Next()
}

func (client *b591) WriteMatch(writer io.Writer, match Match) error {
	writeUint8(writer, uint8(match.Id))
	writeBoolean(writer, match.InProgress)
	writeUint8(writer, match.Type)
	writeUint16(writer, uint16(match.Mods))
	writeString(writer, match.Name)
	writeString(writer, match.Password)
	writeString(writer, match.BeatmapText)
	writeInt32(writer, match.BeatmapId)
	writeString(writer, match.BeatmapChecksum)

	for _, slot := range match.Slots {
		writeUint8(writer, slot.Status)
	}

	if client.ProtocolVersion() >= 4 {
		for _, slot := range match.Slots {
			writeUint8(writer, slot.Team)
		}
	}

	for _, slot := range match.Slots {
		if slot.HasPlayer() {
			writeInt32(writer, slot.UserId)
		}
	}

	writeInt32(writer, match.HostId)
	writeUint8(writer, match.Mode)

	if client.ProtocolVersion() >= 3 {
		writeUint8(writer, match.ScoringType)
		writeUint8(writer, match.TeamType)
	}

	return nil
}

func (client *b591) WriteMatchUpdate(stream io.Writer, match Match) error {
	writer := bytes.NewBuffer([]byte{})
	client.WriteMatch(writer, match)
	return client.WritePacket(stream, BanchoMatchUpdate, writer.Bytes())
}

func (client *b591) WriteMatchNew(stream io.Writer, match Match) error {
	writer := bytes.NewBuffer([]byte{})
	client.WriteMatch(writer, match)
	return client.WritePacket(stream, BanchoMatchNew, writer.Bytes())
}

func (client *b591) WriteMatchJoinSuccess(stream io.Writer, match Match) error {
	writer := bytes.NewBuffer([]byte{})
	client.WriteMatch(writer, match)
	return client.WritePacket(stream, BanchoMatchJoinSuccess, writer.Bytes())
}

func (client *b591) WriteMatchStart(stream io.Writer, match Match) error {
	writer := bytes.NewBuffer([]byte{})
	client.WriteMatch(writer, match)
	return client.WritePacket(stream, BanchoMatchStart, writer.Bytes())
}
