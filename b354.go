package chio

import (
	"bytes"
	"fmt"
	"io"
)

// b354 adds the beatmap info packets
type b354 struct {
	*b349
}

func (client *b354) ReadPacket(stream io.Reader) (packet *BanchoPacket, err error) {
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

func (client *b354) ReadPacketType(packetId uint16, reader io.Reader) (any, error) {
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
		return readUint32(reader)
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
	default:
		return nil, nil
	}
}

func (client *b354) SupportedPackets() []uint16 {
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
		BanchoFellowSpectatorJoined,
		BanchoFellowSpectatorLeft,
		OsuMatchStart,
		BanchoMatchStart,
		OsuMatchScoreUpdate,
		BanchoMatchScoreUpdate,
		OsuMatchComplete,
		OsuMatchChangeSettings,
		BanchoMatchTransferHost,
		OsuMatchChangeMods,
		OsuMatchLoadComplete,
		BanchoMatchAllPlayersLoaded,
		OsuMatchNoBeatmap,
		OsuMatchNotReady,
		OsuMatchFailed,
		BanchoMatchPlayerFailed,
		BanchoMatchComplete,
		OsuMatchSkipRequest,
		BanchoMatchSkip,
		OsuChannelJoin,
		BanchoChannelJoinSuccess,
		BanchoChannelAvailable,
		BanchoChannelRevoked,
		BanchoChannelAvailableAutojoin,
		OsuBeatmapInfoRequest,
		BanchoBeatmapInfoReply,
	}
	return client.supportedPackets
}

func (client *b354) ImplementsPacket(packetId uint16) bool {
	for _, id := range client.SupportedPackets() {
		if id == packetId {
			return true
		}
	}
	return false
}

func (client *b354) ReadBeatmapInfoRequest(reader io.Reader) (*BeatmapInfoRequest, error) {
	var err error
	errors := NewErrorCollection()
	request := &BeatmapInfoRequest{}
	filenameCount, err := readInt32(reader)
	errors.Add(err)

	for range filenameCount {
		filename, err := readString(reader)
		errors.Add(err)
		request.Filenames = append(request.Filenames, filename)
	}

	return request, errors.Next()
}

func (client *b354) WriteBeatmapInfoReply(stream io.Writer, reply BeatmapInfoReply) error {
	buffer := bytes.NewBuffer([]byte{})
	writeInt32(buffer, int32(len(reply.Beatmaps)))

	for _, info := range reply.Beatmaps {
		writeInt16(buffer, info.Index)
		writeInt32(buffer, info.BeatmapId)
		writeInt32(buffer, info.BeatmapSetId)
		writeInt32(buffer, info.ThreadId)
		writeBoolean(buffer, info.IsRanked())
		writeInt8(buffer, info.OsuRank)
		writeString(buffer, info.Checksum)
	}

	return nil
}
