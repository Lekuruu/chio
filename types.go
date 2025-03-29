package chio

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strconv"
)

type UserInfo struct {
	Id       int32
	Name     string
	Presence *UserPresence
	Status   *UserStatus
	Stats    *UserStats
}

func (u *UserInfo) AvatarFilename() string {
	// Hopefully this is the right way to do it
	return fmt.Sprintf("%d_000.png", u.Id)
}

type UserPresence struct {
	IsIrc        bool
	Timezone     int8
	CountryIndex int8
	Permissions  uint8
	Longitude    float32
	Latitude     float32
	City         string
}

func (presence *UserPresence) CountryName() string {
	return CountryNames[presence.CountryIndex]
}

func (presence *UserPresence) CountryCode() string {
	return CountryCodes[presence.CountryIndex]
}

func (presence *UserPresence) Location() string {
	if presence.City != "" {
		return fmt.Sprintf("%s / %s", presence.CountryName(), presence.City)
	}
	return presence.CountryName()
}

type UserStats struct {
	Rank      int32
	Rscore    uint64
	Tscore    uint64
	Accuracy  float64
	Playcount int32
	PP        uint16
}

type UserStatus struct {
	Action          uint8
	Text            string
	Mods            uint32
	Mode            uint8
	BeatmapChecksum string
	BeatmapId       int32
	UpdateStats     bool
}

type UserQuit struct {
	Info      *UserInfo
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
	UserCount int16
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

// IsRanked is used to check whether or not the map is ranked/approved
func (info *BeatmapInfo) IsRanked() bool {
	return info.RankedStatus == RankedStatusRanked ||
		info.RankedStatus == RankedStatusApproved
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
	Frames []*ReplayFrame
	Frame  *ScoreFrame
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
func (sf *ScoreFrame) Checksum() string {
	hash := md5.Sum([]byte(
		strconv.FormatUint(uint64(sf.Time), 10) +
			"false" + // "pass" ?
			strconv.Itoa(int(sf.Total300)) +
			strconv.Itoa(int(sf.Total50)) +
			strconv.Itoa(int(sf.TotalGeki)) +
			strconv.Itoa(int(sf.TotalKatu)) +
			strconv.Itoa(int(sf.TotalMiss)) +
			strconv.Itoa(int(sf.CurrentCombo)) +
			strconv.Itoa(int(sf.MaxCombo)) +
			strconv.Itoa(int(sf.Hp)),
	))

	return hex.EncodeToString(hash[:])
}

type Match struct {
	Id              int32
	InProgress      bool
	Type            uint8
	Mods            uint32
	Name            string
	Password        string
	BeatmapText     string
	BeatmapId       int32
	BeatmapChecksum string
	Slots           []*MatchSlot
	HostId          int32
	Mode            uint8
	ScoringType     uint8
	TeamType        uint8
	Freemod         bool
	Seed            int32
}

type MatchSlot struct {
	UserId int32
	Status uint8
	Team   uint8
	Mods   uint32
}

func (s *MatchSlot) HasPlayer() bool {
	return SlotStatusHasPlayer&uint8(s.Status) > 0
}

type MatchJoin struct {
	MatchId  int32
	Password string
}

type TitleUpdate struct {
	ImageUrl    string
	RedirectUrl string
}
