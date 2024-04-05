package chio

import (
	"crypto/md5"
	"encoding/hex"
)

type UserInfo struct {
	Id           int32
	IsIrc        bool
	Name         string
	Timezone     int8
	CountryIndex int8
	Permissions  uint8
	Mode         uint8
	Longitude    float32
	Latitude     float32
	Rank         uint32
	City         string
	Status       UserStatus
	Rscore       uint64
	Tscore       uint64
	Accuracy     float32
	Playcount    uint32
	Pp           uint16
}

type UserStatus struct {
	Action          uint8
	Text            string
	Mods            uint32
	Mode            uint8
	BeatmapChecksum string
	BeatmapId       int32
}

type UserQuit struct {
	Info      UserInfo
	QuitState uint8
}

type Message struct {
	Sender   string
	Content  string
	Target   string
	SenderId int32
}

type Channel struct {
	Name      string
	Topic     string
	Owner     string
	UserCount int32
}

type BeatmapInfo struct {
	Index        int16
	BeatmapId    int32
	BeatmapSetId int32
	ThreadId     int32
	RankedStatus int8
	OsuRank      int8
	TaikoRank    int8
	FruitsRank   int8
	ManiaRank    int8
	Checksum     string
}

type BeatmapInfoReply struct {
	Beatmaps []BeatmapInfo
}

type BeatmapInfoRequest struct {
	Filenames []string
	Ids       []int32
}

type ReplayFrame struct {
	ButtonState uint8
	LegacyByte  uint8
	MouseX      float32
	MouseY      float32
	Time        int32
}

type ReplayFrameBundle struct {
	Action uint8
	Extra  int32
	Frames []ReplayFrame
	Frame  ScoreFrame
}

type ScoreFrame struct {
	Time         int32
	Id           uint8
	Total300     uint16
	Total100     uint16
	Total50      uint16
	TotalGeki    uint16
	TotalKatu    uint16
	TotalMiss    uint16
	TotalScore   uint32
	MaxCombo     uint16
	CurrentCombo uint16
	Perfect      bool
	Hp           uint8
	TagByte      uint8
}

// ScoreFrame checksum calculation used in version b323
func (sf *ScoreFrame) GetChecksum() string {
	hash := md5.New().Sum([]byte(
		string(sf.Time) +
			"false" + // "pass" ?
			string(sf.Total300) +
			string(sf.Total50) +
			string(sf.TotalGeki) +
			string(sf.TotalKatu) +
			string(sf.TotalMiss) +
			string(sf.CurrentCombo) +
			string(sf.MaxCombo) +
			string(sf.Hp),
	))

	return hex.EncodeToString(hash[:])
}

type Match struct {
	Id              int32
	InProgress      bool
	MatchType       uint8
	Mods            uint32
	Name            string
	Password        string
	BeatmapText     string
	BeatmapId       int32
	BeatmapChecksum string
	Slots           []MatchSlot
	HostId          int32
	Mode            uint8
	ScoringType     uint8
	TeamType        uint8
	Freemode        bool
	Seed            int32
}

type MatchSlot struct {
	UserId int32
	Status uint16
	Team   uint8
	Mods   uint32
}

type MatchJoin struct {
	MatchId  int32
	Password string
}

type TitleUpdate struct {
	ImageUrl    string
	RedirectUrl string
}
