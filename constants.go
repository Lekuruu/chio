package chio

const (
	OsuSendUserStatus              uint16 = 0
	OsuSendIrcMessage              uint16 = 1
	OsuExit                        uint16 = 2
	OsuRequestStatusUpdate         uint16 = 3
	OsuPong                        uint16 = 4
	BanchoLoginReply               uint16 = 5
	BanchoSendMessage              uint16 = 7
	BanchoPing                     uint16 = 8
	BanchoHandleIrcChangeUsername  uint16 = 9
	BanchoHandleIrcQuit            uint16 = 10
	BanchoHandleOsuUpdate          uint16 = 11
	BanchoHandleOsuQuit            uint16 = 12
	BanchoSpectatorJoined          uint16 = 13
	BanchoSpectatorLeft            uint16 = 14
	BanchoSpectateFrames           uint16 = 15
	OsuStartSpectating             uint16 = 16
	OsuStopSpectating              uint16 = 17
	OsuSpectateFrames              uint16 = 18
	BanchoVersionUpdate            uint16 = 19
	OsuErrorReport                 uint16 = 20
	OsuCantSpectate                uint16 = 21
	BanchoSpectatorCantSpectate    uint16 = 22
	BanchoGetAttention             uint16 = 23
	BanchoAnnounce                 uint16 = 24
	OsuSendIrcMessagePrivate       uint16 = 25
	BanchoMatchUpdate              uint16 = 26
	BanchoMatchNew                 uint16 = 27
	BanchoMatchDisband             uint16 = 28
	OsuLobbyPart                   uint16 = 29
	OsuLobbyJoin                   uint16 = 30
	OsuMatchCreate                 uint16 = 31
	OsuMatchJoin                   uint16 = 32
	OsuMatchPart                   uint16 = 33
	BanchoLobbyJoin                uint16 = 34
	BanchoLobbyPart                uint16 = 35
	BanchoMatchJoinSuccess         uint16 = 36
	BanchoMatchJoinFail            uint16 = 37
	OsuMatchChangeSlot             uint16 = 38
	OsuMatchReady                  uint16 = 39
	OsuMatchLock                   uint16 = 40
	OsuMatchChangeSettings         uint16 = 41
	BanchoFellowSpectatorJoined    uint16 = 42
	BanchoFellowSpectatorLeft      uint16 = 43
	OsuMatchStart                  uint16 = 44
	BanchoMatchStart               uint16 = 46
	OsuMatchScoreUpdate            uint16 = 47
	BanchoMatchScoreUpdate         uint16 = 48
	OsuMatchComplete               uint16 = 49
	BanchoMatchTransferHost        uint16 = 50
	OsuMatchChangeMods             uint16 = 51
	OsuMatchLoadComplete           uint16 = 52
	BanchoMatchAllPlayersLoaded    uint16 = 53
	OsuMatchNoBeatmap              uint16 = 54
	OsuMatchNotReady               uint16 = 55
	OsuMatchFailed                 uint16 = 56
	BanchoMatchPlayerFailed        uint16 = 57
	BanchoMatchComplete            uint16 = 58
	OsuMatchHasBeatmap             uint16 = 59
	OsuMatchSkipRequest            uint16 = 60
	BanchoMatchSkip                uint16 = 61
	BanchoUnauthorized             uint16 = 62
	OsuChannelJoin                 uint16 = 63
	BanchoChannelJoinSuccess       uint16 = 64
	BanchoChannelAvailable         uint16 = 65
	BanchoChannelRevoked           uint16 = 66
	BanchoChannelAvailableAutojoin uint16 = 67
	OsuBeatmapInfoRequest          uint16 = 68
	BanchoBeatmapInfoReply         uint16 = 69
	OsuMatchTransferHost           uint16 = 70
	BanchoLoginPermissions         uint16 = 71
	BanchoFriendsList              uint16 = 72
	OsuFriendsAdd                  uint16 = 73
	OsuFriendsRemove               uint16 = 74
	BanchoProtocolNegotiation      uint16 = 75
	BanchoTitleUpdate              uint16 = 76
	OsuMatchChangeTeam             uint16 = 77
	OsuChannelLeave                uint16 = 78
	OsuReceiveUpdates              uint16 = 79
	BanchoMonitor                  uint16 = 80
	BanchoMatchPlayerSkipped       uint16 = 81
	OsuSetIrcAwayMessage           uint16 = 82
	BanchoUserPresence             uint16 = 83
	OsuUserStatsRequest            uint16 = 85
	BanchoRestart                  uint16 = 86
	OsuInvite                      uint16 = 87
	BanchoInvite                   uint16 = 88
	BanchoChannelInfoComplete      uint16 = 89
	OsuMatchChangePassword         uint16 = 90
	BanchoMatchChangePassword      uint16 = 91
	BanchoSilenceInfo              uint16 = 92
	OsuTournamentMatchInfo         uint16 = 93
	BanchoUserSilenced             uint16 = 94
	BanchoUserPresenceSingle       uint16 = 95
	BanchoUserPresenceBundle       uint16 = 96
	OsuPresenceRequest             uint16 = 97
	OsuPresenceRequestAll          uint16 = 98
	OsuChangeFriendOnlyDMs         uint16 = 99
	BanchoUserDMsBlocked           uint16 = 100
	BanchoTargetIsSilenced         uint16 = 101
	BanchoVersionUpdateForced      uint16 = 102
	BanchoSwitchServer             uint16 = 103
	BanchoAccountRestricted        uint16 = 104
	BanchoRTX                      uint16 = 105
	BanchoMatchAbort               uint16 = 106
	BanchoSwitchTournamentServer   uint16 = 107
	OsuTournamentJoinMatchChannel  uint16 = 108
	OsuTournamentLeaveMatchChannel uint16 = 109
)

const (
	StatusIdle         uint8 = 0
	StatusAfk          uint8 = 1
	StatusPlaying      uint8 = 2
	StatusEditing      uint8 = 3
	StatusModding      uint8 = 4
	StatusMultiplayer  uint8 = 5
	StatusWatching     uint8 = 6
	StatusUnknown      uint8 = 7
	StatusTesting      uint8 = 8
	StatusSubmitting   uint8 = 9
	StatusStatsUpdate  uint8 = 10
	StatusPaused       uint8 = 10
	StatusLobby        uint8 = 11
	StatusMultiplaying uint8 = 12
	StatusOsuDirect    uint8 = 13
)

const (
	ModeOsu   uint8 = 0
	ModeTaiko uint8 = 1
	ModeCatch uint8 = 2
	ModeMania uint8 = 3
)

const (
	InvalidLogin          int32 = -1
	InvalidVersion        int32 = -2
	UserBanned            int32 = -3
	UserInactive          int32 = -4
	ServerError           int32 = -5
	UnauthorizedTestBuild int32 = -6
)

const (
	PermissionsNone       = 0
	PermissionsRegular    = 1 << 0
	PermissionsBAT        = 1 << 1
	PermissionsSupporter  = 1 << 2
	PermissionsFriend     = 1 << 3
	PermissionsPeppy      = 1 << 4
	PermissionsTournament = 1 << 5
)

const (
	QuitStateGone         uint8 = 0
	QuitStateOsuRemaining uint8 = 1
	QuitStateIrcRemaining uint8 = 2
)

const (
	AvatarExtensionNone = 0
	AvatarExtensionPng  = 1
	AvatarExtensionJpg  = 2
)

const (
	PresenceFilterNone    uint8 = 0
	PresenceFilterAll     uint8 = 1
	PresenceFilterFriends uint8 = 2
)

const (
	ReplayActionStandard      uint8 = 0
	ReplayActionNewSong       uint8 = 1
	ReplayActionSkip          uint8 = 2
	ReplayActionCompletion    uint8 = 3
	ReplayActionFail          uint8 = 4
	ReplayActionPause         uint8 = 5
	ReplayActionUnpause       uint8 = 6
	ReplayActionSongSelect    uint8 = 7
	ReplayActionWatchingOther uint8 = 8
)

const (
	ButtonStateNoButton uint8 = 0
	ButtonStateLeft1    uint8 = 1 << 0
	ButtonStateRight1   uint8 = 1 << 1
	ButtonStateLeft2    uint8 = 1 << 2
	ButtonStateRight2   uint8 = 1 << 3
	ButtonStateSmoke    uint8 = 1 << 4
)

const (
	OsuRankXH int8 = 0
	OsuRankSH int8 = 1
	OsuRankX  int8 = 2
	OsuRankS  int8 = 3
	OsuRankA  int8 = 4
	OsuRankB  int8 = 5
	OsuRankC  int8 = 6
	OsuRankD  int8 = 7
	OsuRankF  int8 = 8
	OsuRankN  int8 = 9
)

const (
	NoMod       uint32 = 0
	NoFail      uint32 = 1 << 0
	Easy        uint32 = 1 << 1
	NoVideo     uint32 = 1 << 2
	Hidden      uint32 = 1 << 3
	HardRock    uint32 = 1 << 4
	SuddenDeath uint32 = 1 << 5
	DoubleTime  uint32 = 1 << 6
	Relax       uint32 = 1 << 7
	HalfTime    uint32 = 1 << 8
	Nightcore   uint32 = 1 << 9
	Flashlight  uint32 = 1 << 10
	Autoplay    uint32 = 1 << 11
	SpunOut     uint32 = 1 << 12
	Autopilot   uint32 = 1 << 13
	Perfect     uint32 = 1 << 14
	Key4        uint32 = 1 << 15
	Key5        uint32 = 1 << 16
	Key6        uint32 = 1 << 17
	Key7        uint32 = 1 << 18
	Key8        uint32 = 1 << 19
	FadeIn      uint32 = 1 << 20
	Random      uint32 = 1 << 21
	Cinema      uint32 = 1 << 22
	Target      uint32 = 1 << 23
	Key9        uint32 = 1 << 24
	KeyCoop     uint32 = 1 << 25
	Key1        uint32 = 1 << 26
	Key3        uint32 = 1 << 27
	Key2        uint32 = 1 << 28
	ScoreV2     uint32 = 1 << 29
	Mirror      uint32 = 1 << 30
)

const (
	MatchTypeStandard  uint8 = 0
	MatchTypePowerplay uint8 = 1
)

const (
	ScoringTypeScore    uint8 = 0
	ScoringTypeAccuracy uint8 = 1
	ScoringTypeCombo    uint8 = 2
	ScoringTypeScoreV2  uint8 = 3
)

const (
	TeamTypeHeadToHead uint8 = 0
	TeamTypeTagCoop    uint8 = 1
	TeamTypeTeamVs     uint8 = 2
	TeamTypeTagTeam    uint8 = 3
)

const (
	SlotStatusOpen      uint8 = 1 << 0
	SlotStatusLocked    uint8 = 1 << 1
	SlotStatusNotReady  uint8 = 1 << 2
	SlotStatusReady     uint8 = 1 << 3
	SlotStatusNoMap     uint8 = 1 << 4
	SlotStatusPlaying   uint8 = 1 << 5
	SlotStatusComplete  uint8 = 1 << 6
	SlotStatusQuit      uint8 = 1 << 7
	SlotStatusHasPlayer uint8 = SlotStatusNotReady | SlotStatusReady | SlotStatusNoMap | SlotStatusPlaying | SlotStatusComplete
)

const (
	SlotTeamNeutral uint8 = 0
	SlotTeamBlue    uint8 = 1
	SlotTeamRed     uint8 = 2
)

const (
	RankedStatusPending   int8 = 0
	RankedStatusRanked    int8 = 1
	RankedStatusApproved  int8 = 2
	RankedStatusQualified int8 = 3
)
