package chio

import (
	"bytes"
	"io"
)

// b425 adds support for user permissions
type b425 struct {
	*b402
}

func (client *b425) SupportedPackets() []uint16 {
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
		OsuMatchTransferHost,
		BanchoLoginPermissions,
	}
	return client.supportedPackets
}

func (client *b425) ImplementsPacket(packetId uint16) bool {
	for _, id := range client.SupportedPackets() {
		if id == packetId {
			return true
		}
	}
	return false
}

func (client b425) WriteLoginPermissions(stream io.Writer, permissions uint32) error {
	writer := bytes.NewBuffer([]byte{})
	writeUint32(writer, permissions)
	return client.WritePacket(stream, BanchoLoginPermissions, writer.Bytes())
}
