package chio

import "io"

// BanchoPacket is a struct that represents a packet that
// is sent or received
type BanchoPacket struct {
	Id   uint16
	Data interface{}
}

// BanchoIO is an interface that wraps the basic methods for
// reading and writing packets to a Bancho client
type BanchoIO interface {
	// Write writes len(p) bytes from p to the underlying data stream
	Write(p []byte) (n int, err error)

	// Read reads up to len(p) bytes into p. It returns the number of bytes
	Read(p []byte) (n int, err error)

	// Close closes the underlying data stream
	Close() error

	// Clone returns a copy of the BanchoIO interface
	Clone() BanchoIO

	// GetStream returns the underlying data stream
	GetStream() io.ReadWriteCloser

	// SetWriter sets the underlying data stream
	SetStream(stream io.ReadWriteCloser)

	// WritePacket writes a packet to the underlying data stream
	WritePacket(packetId uint16, data []byte) error

	// ReadPacket reads a packet from the underlying data stream
	ReadPacket() (packet *BanchoPacket, err error)

	// SupportedPackets returns a list of packetIds that are supported by the client
	SupportedPackets() []uint16

	// ImplementsPacket checks if the packetId is implemented in the client
	ImplementsPacket(packetId uint16) bool

	// Packet writers
	BanchoWriters
}

// BanchoWriters is an interface that wraps the methods for writing
// to a Bancho client
type BanchoWriters interface {
	WriteLoginReply(reply int32) error
	WriteMessage(message Message) error
	WritePing() error
	WriteIrcChangeUsername(oldName string, newName string) error
	WriteUserStats(info UserInfo) error
	WriteUserQuit(quit UserQuit) error
	WriteSpectatorJoined(userId int32) error
	WriteSpectatorLeft(userId int32) error
	WriteSpectateFrames(bundle ReplayFrameBundle) error
	WriteVersionUpdate() error
	WriteSpectatorCantSpectate(userId int32) error
	WriteGetAttention() error
	WriteAnnouncement(message string) error
	WriteMatchUpdate(match Match) error
	WriteMatchNew(match Match) error
	WriteMatchDisband(matchId uint32) error
	WriteLobbyJoin(userId int32) error
	WriteLobbyPart(userId int32) error
	WriteMatchJoinSuccess(match Match) error
	WriteMatchJoinFail() error
	WriteFellowSpectatorJoined(userId int32) error
	WriteFellowSpectatorLeft(userId int32) error
	WriteMatchStart(match Match) error
	WriteMatchScoreUpdate(frame ScoreFrame) error
	WriteMatchTransferHost() error
	WriteMatchAllPlayersLoaded() error
	WriteMatchPlayerFailed(slotId uint32) error
	WriteMatchComplete() error
	WriteMatchSkip() error
	WriteUnauthorized() error
	WriteChannelJoinSuccess(channel string) error
	WriteChannelRevoked(channel string) error
	WriteChannelAvailable(channel Channel) error
	WriteChannelAvailableAutojoin(channel Channel) error
	WriteBeatmapInfoReply(reply BeatmapInfoReply) error
	WriteLoginPermissions(permissions uint32) error
	WriteFriendsList(userIds []uint32) error
	WriteProtocolNegotiation(version int32) error
	WriteTitleUpdate(update TitleUpdate) error
	WriteMonitor() error
	WriteMatchPlayerSkipped(slotId int32) error
	WriteUserPresence(info UserInfo) error
	WriteRestart(retryMs int32) error
	WriteInvite(message Message) error
	WriteChannelInfoComplete() error
	WriteMatchChangePassword(password string) error
	WriteSilenceInfo(timeRemaining int32) error
	WriteUserSilenced(userId uint32) error
	WriteUserPresenceSingle(info UserInfo) error
	WriteUserPresenceBundle(infos []UserInfo) error
	WriteUserDMsBlocked(targetName string) error
	WriteTargetIsSilenced(targetName string) error
	WriteVersionUpdateForced() error
	WriteSwitchServer(target int32) error
	WriteAccountRestricted() error
	WriteRTX(message string) error
	WriteMatchAbort() error
	WriteSwitchTournamentServer(ip string) error
}

var clients map[int]BanchoIO = make(map[int]BanchoIO)

const lowestVersion int = 282
const highestVersion int = 294

// GetClientInterface returns a BanchoIO interface for the given client version
func GetClientInterface(stream io.ReadWriteCloser, clientVersion int) BanchoIO {
	if clientVersion < lowestVersion {
		client := clients[lowestVersion]
		return initializeClient(stream, client)
	}

	if clientVersion > highestVersion {
		client := clients[highestVersion]
		return initializeClient(stream, client)
	}

	if client, ok := clients[clientVersion]; ok {
		return initializeClient(stream, client)
	}

	// Find the next compatible version
	for version, client := range clients {
		if version <= clientVersion {
			return initializeClient(stream, client)
		}
	}

	return nil
}

func initializeClient(rw io.ReadWriteCloser, io BanchoIO) BanchoIO {
	newIO := io.Clone()
	newIO.SetStream(rw)
	return newIO
}

func init() {
	clients[282] = &b282{}
	clients[290] = &b282{}
	clients[291] = &b291{}
	clients[293] = &b291{}
	clients[294] = &b294{}
}
