package chio

import "io"

// BanchoPackcet is a struct that represents a packet that
// is sent or received
type BanchoPacket struct {
	Id   uint16
	Data interface{}
}

// BanchoIO is an interface that wraps the basic methods for
// reading and writing packets to a Bancho client
type BanchoIO interface {
	// WritePacket writes a packet to the provided stream
	WritePacket(stream io.Writer, packetId uint16, data []byte) error

	// ReadPacket reads a packet from the provided stream
	ReadPacket(stream io.Reader) (packet *BanchoPacket, err error)

	// SupportedPackets returns a list of packetIds that are supported by the client
	SupportedPackets() []uint16

	// ImplementsPacket checks if the packetId is implemented in the client
	ImplementsPacket(packetId uint16) bool

	// ProtocolVersion returns the bancho protocol version used by the client
	ProtocolVersion() int

	// OverrideProtocolVersion lets you specify a custom bancho protocol version
	OverrideProtocolVersion(version int)

	// MatchSlotSize returns the number of slots that are used in the match
	MatchSlotSize() int

	// OverrideMatchSlotSize lets you specify a custom amount of slots to read & write to the client
	OverrideMatchSlotSize(amount int)

	// Packet writers
	BanchoWriters
}

// BanchoWriters is an interface that wraps the methods for writing
// to a Bancho client
type BanchoWriters interface {
	WriteLoginReply(stream io.Writer, reply int32) error
	WriteMessage(stream io.Writer, message Message) error
	WritePing(stream io.Writer) error
	WriteIrcChangeUsername(stream io.Writer, oldName, newName string) error
	WriteUserStats(stream io.Writer, info UserInfo) error
	WriteUserQuit(stream io.Writer, quit UserQuit) error
	WriteSpectatorJoined(stream io.Writer, userId int32) error
	WriteSpectatorLeft(stream io.Writer, userId int32) error
	WriteSpectateFrames(stream io.Writer, bundle ReplayFrameBundle) error
	WriteVersionUpdate(stream io.Writer) error
	WriteSpectatorCantSpectate(stream io.Writer, userId int32) error
	WriteGetAttention(stream io.Writer) error
	WriteAnnouncement(stream io.Writer, message string) error
	WriteMatchUpdate(stream io.Writer, match Match) error
	WriteMatchNew(stream io.Writer, match Match) error
	WriteMatchDisband(stream io.Writer, matchId int32) error
	WriteLobbyJoin(stream io.Writer, userId int32) error
	WriteLobbyPart(stream io.Writer, userId int32) error
	WriteMatchJoinSuccess(stream io.Writer, match Match) error
	WriteMatchJoinFail(stream io.Writer) error
	WriteFellowSpectatorJoined(stream io.Writer, userId int32) error
	WriteFellowSpectatorLeft(stream io.Writer, userId int32) error
	WriteMatchStart(stream io.Writer, match Match) error
	WriteMatchScoreUpdate(stream io.Writer, frame ScoreFrame) error
	WriteMatchTransferHost(stream io.Writer) error
	WriteMatchAllPlayersLoaded(stream io.Writer) error
	WriteMatchPlayerFailed(stream io.Writer, slotId uint32) error
	WriteMatchComplete(stream io.Writer) error
	WriteMatchSkip(stream io.Writer) error
	WriteUnauthorized(stream io.Writer) error
	WriteChannelJoinSuccess(stream io.Writer, channel string) error
	WriteChannelRevoked(stream io.Writer, channel string) error
	WriteChannelAvailable(stream io.Writer, channel Channel) error
	WriteChannelAvailableAutojoin(stream io.Writer, channel Channel) error
	WriteBeatmapInfoReply(stream io.Writer, reply BeatmapInfoReply) error
	WriteLoginPermissions(stream io.Writer, permissions uint32) error
	WriteFriendsList(stream io.Writer, userIds []int32) error
	WriteProtocolNegotiation(stream io.Writer, version int32) error
	WriteTitleUpdate(stream io.Writer, update TitleUpdate) error
	WriteMonitor(stream io.Writer) error
	WriteMatchPlayerSkipped(stream io.Writer, slotId int32) error
	WriteUserPresence(stream io.Writer, info UserInfo) error
	WriteRestart(stream io.Writer, retryMs int32) error
	WriteInvite(stream io.Writer, message Message) error
	WriteChannelInfoComplete(stream io.Writer) error
	WriteMatchChangePassword(stream io.Writer, password string) error
	WriteSilenceInfo(stream io.Writer, timeRemaining int32) error
	WriteUserSilenced(stream io.Writer, userId uint32) error
	WriteUserPresenceSingle(stream io.Writer, info UserInfo) error
	WriteUserPresenceBundle(stream io.Writer, infos []UserInfo) error
	WriteUserDMsBlocked(stream io.Writer, targetName string) error
	WriteTargetIsSilenced(stream io.Writer, targetName string) error
	WriteVersionUpdateForced(stream io.Writer) error
	WriteSwitchServer(stream io.Writer, target int32) error
	WriteAccountRestricted(stream io.Writer) error
	WriteRTX(stream io.Writer, message string) error
	WriteMatchAbort(stream io.Writer) error
	WriteSwitchTournamentServer(stream io.Writer, ip string) error
}

var clients map[int]BanchoIO = make(map[int]BanchoIO)

const lowestVersion int = 282
const highestVersion int = 558

// GetClientInterface returns a BanchoIO interface for the given client version
func GetClientInterface(clientVersion int) BanchoIO {
	if clientVersion < lowestVersion {
		client := clients[lowestVersion]
		return client
	}

	if clientVersion > highestVersion {
		client := clients[highestVersion]
		return client
	}

	if client, ok := clients[clientVersion]; ok {
		return client
	}

	// Find the next compatible version
	for version, client := range clients {
		if version <= clientVersion {
			return client
		}
	}

	return nil
}

func init() {
	clients[282] = &b282{slotSize: 8, protocolVersion: 0}
	clients[290] = clients[282]
	clients[291] = &b291{clients[282].(*b282)}
	clients[293] = clients[291]
	clients[294] = &b294{clients[291].(*b291)}
	clients[295] = clients[294]
	clients[296] = &b296{clients[294].(*b294)}
	clients[297] = clients[296]
	clients[298] = &b298{clients[296].(*b296)}
	clients[311] = clients[298]
	clients[312] = &b312{clients[298].(*b298)}
	clients[319] = clients[312]
	clients[320] = &b320{clients[312].(*b312)}
	clients[322] = clients[320]
	clients[323] = &b323{clients[320].(*b320)}
	clients[333] = clients[323]
	clients[334] = &b334{clients[323].(*b323)}
	clients[337] = clients[334]
	clients[338] = &b338{clients[334].(*b334)}
	clients[339] = clients[338]
	clients[340] = &b340{clients[338].(*b338)}
	clients[341] = clients[340]
	clients[342] = &b342{clients[340].(*b340)}
	clients[348] = clients[342]
	clients[349] = &b349{clients[342].(*b342)}
	clients[353] = clients[349]
	clients[354] = &b354{clients[349].(*b349)}
	clients[387] = clients[354]
	clients[388] = &b388{clients[354].(*b354)}
	clients[401] = clients[388]
	clients[402] = &b402{clients[388].(*b388)}
	clients[424] = clients[402]
	clients[425] = &b425{clients[402].(*b402)}
	clients[451] = clients[425]
	clients[452] = &b452{clients[425].(*b425)}
	clients[469] = clients[452]
	clients[470] = &b470{clients[452].(*b452)}
	clients[486] = clients[470]
	clients[487] = &b487{clients[470].(*b470)}
	clients[488] = clients[487]
	clients[489] = &b489{clients[487].(*b487), 1}
	clients[490] = &b490{clients[489].(*b489)}
	clients[503] = clients[490]
	clients[504] = &b504{clients[490].(*b490), 2}
	clients[534] = clients[504]
	clients[535] = &b535{clients[504].(*b504), 3}
	clients[557] = clients[535]
	clients[558] = &b558{clients[535].(*b535), 4}
	clients[564] = clients[558]
	clients[365] = &b365{clients[354].(*b354)}
	clients[373] = clients[365]
	clients[374] = &b374{clients[365].(*b365)}
	clients[590] = clients[558]
	clients[591] = &b591{clients[558].(*b558)}
	clients[612] = clients[591]
	clients[613] = &b613{clients[591].(*b591)}
}
