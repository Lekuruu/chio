package chio

const (
	OsuSendUserStatus              uint16 = 0
	OsuSendIrcMessage              uint16 = 1
	OsuExit                        uint16 = 2
	OsuRequestStatusUpdate         uint16 = 3
	OsuPong                        uint16 = 4
	BanchoLoginReply               uint16 = 5
	BanchoCommandError             uint16 = 6
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

	/* Packets that are unused today, but used in legacy clients */

	BanchoHandleIrcJoin   uint16 = 0xFFFF
	OsuMatchChangeBeatmap uint16 = 0xFFFE
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
	StatusPaused       uint8 = 10
	StatusLobby        uint8 = 11
	StatusMultiplaying uint8 = 12
	StatusOsuDirect    uint8 = 13

	// Unused in later versions, but required for compatibility
	StatusStatsUpdate uint8 = 10
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
	CompletenessStatusOnly uint8 = 0
	CompletenessStatistics uint8 = 1
	CompletenessFull       uint8 = 2
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

var CountryNames []string = []string{
	"Unknown",
	"Oceania",
	"Europe",
	"Andorra",
	"United Arab Emirates",
	"Afghanistan",
	"Antigua and Barbuda",
	"Anguilla",
	"Albania",
	"Armenia",
	"Netherlands Antilles",
	"Angola",
	"Antarctica",
	"Argentina",
	"American Samoa",
	"Austria",
	"Australia",
	"Aruba",
	"Azerbaijan",
	"Bosnia and Herzegovina",
	"Barbados",
	"Bangladesh",
	"Belgium",
	"Burkina Faso",
	"Bulgaria",
	"Bahrain",
	"Burundi",
	"Benin",
	"Bermuda",
	"Brunei Darussalam",
	"Bolivia",
	"Brazil",
	"The Bahamas",
	"Bhutan",
	"Bouvet Island",
	"Botswana",
	"Belarus",
	"Belize",
	"Canada",
	"Cocos (Keeling) Islands",
	"Democratic Republic of the Congo",
	"Central African Republic",
	"Republic of the Congo",
	"Switzerland",
	"Côte d'Ivoire",
	"Cook Islands",
	"Chile",
	"Cameroon",
	"China",
	"Colombia",
	"Costa Rica",
	"Cuba",
	"Cape Verde",
	"Christmas Island",
	"Cyprus",
	"Czech Republic",
	"Germany",
	"Djibouti",
	"Denmark",
	"Dominica",
	"Dominican Republic",
	"Algeria",
	"Ecuador",
	"Estonia",
	"Egypt",
	"Western Sahara",
	"Eritrea",
	"Spain",
	"Ethiopia",
	"Finland",
	"Fiji",
	"Falkland Islands (Malvinas)",
	"Micronesia, Federated States of Micronesia",
	"Faroe Islands",
	"France",
	"France, Metropolitan",
	"Gabon",
	"United Kingdom",
	"Grenada",
	"Georgia",
	"French Guiana",
	"Ghana",
	"Gibraltar",
	"Greenland",
	"Gambia",
	"Guinea",
	"Guadeloupe",
	"Equatorial Guinea",
	"Greece",
	"South Georgia and the South Sandwich Islands",
	"Guatemala",
	"Guam",
	"Guinea-Bissau",
	"Guyana",
	"Hong Kong",
	"Heard Island and McDonald Islands",
	"Honduras",
	"Croatia",
	"Haiti",
	"Hungary",
	"Indonesia",
	"Ireland",
	"Israel",
	"India",
	"British Indian Ocean Territory",
	"Iraq",
	"Iran, Islamic Republic of Iran",
	"Iceland",
	"Italy",
	"Jamaica",
	"Jordan",
	"Japan",
	"Kenya",
	"Kyrgyzstan",
	"Cambodia",
	"Kiribati",
	"Comoros",
	"Saint Kitts and Nevis",
	"Korea, Democratic People's Republic of Korea",
	"Korea, Republic of Korea",
	"Kuwait",
	"Cayman Islands",
	"Kazakhstan",
	"Lao People's Democratic Republic",
	"Lebanon",
	"Saint Lucia",
	"Liechtenstein",
	"Sri Lanka",
	"Liberia",
	"Lesotho",
	"Lithuania",
	"Luxembourg",
	"Latvia",
	"Libyan Arab Jamahiriya",
	"Morocco",
	"Monaco",
	"Moldova, Republic of Moldova",
	"Madagascar",
	"Marshall Islands",
	"Macedonia, the Former Yugoslav Republic of Macedonia",
	"Mali",
	"Myanmar",
	"Mongolia",
	"Macau",
	"Northern Mariana Islands",
	"Martinique",
	"Mauritania",
	"Montserrat",
	"Malta",
	"Mauritius",
	"Maldives",
	"Malawi",
	"Mexico",
	"Malaysia",
	"Mozambique",
	"Namibia",
	"New Caledonia",
	"Niger",
	"Norfolk Island",
	"Nigeria",
	"Nicaragua",
	"Netherlands",
	"Norway",
	"Nepal",
	"Nauru",
	"Niue",
	"New Zealand",
	"Oman",
	"Panama",
	"Peru",
	"French Polynesia",
	"Papua New Guinea",
	"Philippines",
	"Pakistan",
	"Poland",
	"Saint Pierre and Miquelon",
	"Pitcairn",
	"Puerto Rico",
	"Palestinian Territory, Occupied",
	"Portugal",
	"Palau",
	"Paraguay",
	"Qatar",
	"Réunion",
	"Romania",
	"Russian Federation",
	"Rwanda",
	"Saudi Arabia",
	"Solomon Islands",
	"Seychelles",
	"Sudan",
	"Sweden",
	"Singapore",
	"Saint Helena, Ascension and Tristan da Cunha",
	"Slovenia",
	"Svalbard and Jan Mayen",
	"Slovakia",
	"Sierra Leone",
	"San Marino",
	"Senegal",
	"Somalia",
	"Suriname",
	"Sao Tome and Principe",
	"El Salvador",
	"Syrian Arab Republic",
	"Eswatini",
	"Turks and Caicos Islands",
	"Chad",
	"French Southern Territories",
	"Togo",
	"Thailand",
	"Tajikistan",
	"Tokelau",
	"Turkmenistan",
	"Tunisia",
	"Tonga",
	"Timor-Leste",
	"Turkey",
	"Trinidad and Tobago",
	"Tuvalu",
	"Taiwan, Province of China",
	"Tanzania, United Republic of Tanzania",
	"Ukraine",
	"Uganda",
	"United States Minor Outlying Islands",
	"United States",
	"Uruguay",
	"Uzbekistan",
	"Holy See (Vatican City State)",
	"Saint Vincent and the Grenadines",
	"Venezuela, Bolivarian Republic of Venezuela",
	"Virgin Islands, British",
	"Virgin Islands, U.S.",
	"Viet Nam",
	"Vanuatu",
	"Wallis and Futuna",
	"Samoa",
	"Yemen",
	"Mayotte",
	"Serbia",
	"South Africa",
	"Zambia",
	"Montenegro",
	"Zimbabwe",
	"Unknown",
	"Satellite Provider",
	"Other",
	"Åland Islands",
	"Guernsey",
	"Isle of Man",
	"Jersey",
	"Saint Barthélemy",
	"Saint Martin (French part)",
}

var CountryCodes []string = []string{
	"XX",
	"OC",
	"EU",
	"AD",
	"AE",
	"AF",
	"AG",
	"AI",
	"AL",
	"AM",
	"AN",
	"AO",
	"AQ",
	"AR",
	"AS",
	"AT",
	"AU",
	"AW",
	"AZ",
	"BA",
	"BB",
	"BD",
	"BE",
	"BF",
	"BG",
	"BH",
	"BI",
	"BJ",
	"BM",
	"BN",
	"BO",
	"BR",
	"BS",
	"BT",
	"BV",
	"BW",
	"BY",
	"BZ",
	"CA",
	"CC",
	"CD",
	"CF",
	"CG",
	"CH",
	"CI",
	"CK",
	"CL",
	"CM",
	"CN",
	"CO",
	"CR",
	"CU",
	"CV",
	"CX",
	"CY",
	"CZ",
	"DE",
	"DJ",
	"DK",
	"DM",
	"DO",
	"DZ",
	"EC",
	"EE",
	"EG",
	"EH",
	"ER",
	"ES",
	"ET",
	"FI",
	"FJ",
	"FK",
	"FM",
	"FO",
	"FR",
	"FX",
	"GA",
	"GB",
	"GD",
	"GE",
	"GF",
	"GH",
	"GI",
	"GL",
	"GM",
	"GN",
	"GP",
	"GQ",
	"GR",
	"GS",
	"GT",
	"GU",
	"GW",
	"GY",
	"HK",
	"HM",
	"HN",
	"HR",
	"HT",
	"HU",
	"ID",
	"IE",
	"IL",
	"IN",
	"IO",
	"IQ",
	"IR",
	"IS",
	"IT",
	"JM",
	"JO",
	"JP",
	"KE",
	"KG",
	"KH",
	"KI",
	"KM",
	"KN",
	"KP",
	"KR",
	"KW",
	"KY",
	"KZ",
	"LA",
	"LB",
	"LC",
	"LI",
	"LK",
	"LR",
	"LS",
	"LT",
	"LU",
	"LV",
	"LY",
	"MA",
	"MC",
	"MD",
	"MG",
	"MH",
	"MK",
	"ML",
	"MM",
	"MN",
	"MO",
	"MP",
	"MQ",
	"MR",
	"MS",
	"MT",
	"MU",
	"MV",
	"MW",
	"MX",
	"MY",
	"MZ",
	"NA",
	"NC",
	"NE",
	"NF",
	"NG",
	"NI",
	"NL",
	"NO",
	"NP",
	"NR",
	"NU",
	"NZ",
	"OM",
	"PA",
	"PE",
	"PF",
	"PG",
	"PH",
	"PK",
	"PL",
	"PM",
	"PN",
	"PR",
	"PS",
	"PT",
	"PW",
	"PY",
	"QA",
	"RE",
	"RO",
	"RU",
	"RW",
	"SA",
	"SB",
	"SC",
	"SD",
	"SE",
	"SG",
	"SH",
	"SI",
	"SJ",
	"SK",
	"SL",
	"SM",
	"SN",
	"SO",
	"SR",
	"ST",
	"SV",
	"SY",
	"SZ",
	"TC",
	"TD",
	"TF",
	"TG",
	"TH",
	"TJ",
	"TK",
	"TM",
	"TN",
	"TO",
	"TL",
	"TR",
	"TT",
	"TV",
	"TW",
	"TZ",
	"UA",
	"UG",
	"UM",
	"US",
	"UY",
	"UZ",
	"VA",
	"VC",
	"VE",
	"VG",
	"VI",
	"VN",
	"VU",
	"WF",
	"WS",
	"YE",
	"YT",
	"RS",
	"ZA",
	"ZM",
	"ME",
	"ZW",
	"XA",
	"SP",
	"XO",
	"AX",
	"GG",
	"IM",
	"JE",
	"BL",
	"MF",
}

func GetCountryIndexFromCode(code string) int8 {
	for i, c := range CountryCodes {
		if c == code {
			return int8(i)
		}
	}
	return 0
}

func GetCountryIndexFromName(name string) int8 {
	for i, c := range CountryNames {
		if c == name {
			return int8(i)
		}
	}
	return 0
}
