package chio

import (
	"testing"
)

func TestEncodeChannel(t *testing.T) {
	data, err := Encode(
		BanchoChannelAvailable,
		Channel{
			Name:      "#osu",
			Topic:     "Welcome to #osu!",
			UserCount: 12069,
		},
		20130815,
	)

	if err != nil {
		t.Error(err)
	}

	t.Log(data)
}

func TestEncodeMessage(t *testing.T) {
	data, err := Encode(
		BanchoSendMessage,
		Message{
			Sender:   "peppy",
			Content:  "Hello, world!",
			Target:   "#osu",
			SenderId: 2,
		},
		20130815,
	)

	if err != nil {
		t.Error(err)
	}

	t.Log(data)
}

func TestEncodeUserPresence(t *testing.T) {
	data, err := Encode(
		BanchoHandleOsuUpdate,
		UserInfo{
			Id:           2,
			IsIrc:        false,
			Name:         "peppy",
			Timezone:     10,
			CountryIndex: 16,
			Permissions:  UserPermissionsRegular | UserPermissionsPeppy,
			Mode:         OsuGamemodeOsu,
			Longitude:    135.5,
			Latitude:     -25.5,
			Rank:         764188,
		},
		20130815,
	)

	if err != nil {
		t.Error(err)
	}

	t.Log(data)
}

func TestEncodeUserStats(t *testing.T) {
	data, err := Encode(
		BanchoHandleOsuUpdate,
		UserInfo{
			Id:           2,
			IsIrc:        false,
			Name:         "peppy",
			Timezone:     10,
			CountryIndex: 16,
			Permissions:  UserPermissionsRegular | UserPermissionsPeppy,
			Mode:         OsuGamemodeOsu,
			Longitude:    135.5,
			Latitude:     -25.5,
			Rank:         764188,
			City:         "Unknown",
			Status: UserStatus{
				Action:          OsuStatusPlaying,
				Text:            "Peter Lambert - osu! tutorial (peppy) [Gameplay basics]",
				Mods:            Hidden | HardRock | DoubleTime | Flashlight | Perfect,
				Mode:            OsuGamemodeOsu,
				BeatmapChecksum: "3c8b50ebd781978beb39160c6aaf148c",
				BeatmapId:       22538,
			},
			Rscore:    445065750,
			Tscore:    1988967885,
			Accuracy:  0.9713,
			Playcount: 7695,
			Pp:        1144,
		},
		20130815,
	)

	if err != nil {
		t.Error(err)
	}

	t.Log(data)
}

func TestEncodeQuit(t *testing.T) {
	data, err := Encode(
		BanchoHandleOsuQuit,
		UserQuit{
			UserInfo{
				Id:    2,
				IsIrc: false,
				Name:  "peppy",
			},
			QuitStateGone,
		},
		20130815,
	)

	if err != nil {
		t.Error(err)
	}

	t.Log(data)
}

func TestEncodeBeatmapInfoReply(t *testing.T) {
	data, err := Encode(
		BanchoBeatmapInfoReply,
		BeatmapInfoReply{
			Beatmaps: []BeatmapInfo{
				{
					Index:        0,
					BeatmapId:    22538,
					BeatmapSetId: 22538,
					ThreadId:     0,
					RankedStatus: RankedStatusPending,
					OsuRank:      OsuRankN,
					TaikoRank:    OsuRankN,
					FruitsRank:   OsuRankN,
					ManiaRank:    OsuRankN,
					Checksum:     "3c8b50ebd781978beb39160c6aaf148c",
				},
			},
		},
		20130815,
	)

	if err != nil {
		t.Error(err)
	}

	t.Log(data)
}

func TestEncodeReplyFrameBundle(t *testing.T) {
	data, err := Encode(
		BanchoSpectateFrames,
		ReplayFrameBundle{
			Action: ReplayActionStandard,
			Extra:  0,
			Frames: []ReplayFrame{
				{
					ButtonState: ButtonStateNoButton,
					LegacyByte:  0,
					MouseX:      1.25346,
					MouseY:      2.55346,
					Time:        3468,
				},
			},
			Frame: &ScoreFrame{
				Time:         3468,
				Id:           2,
				Total300:     5,
				Total100:     1,
				Total50:      0,
				TotalGeki:    0,
				TotalKatu:    0,
				TotalMiss:    0,
				TotalScore:   13461,
				MaxCombo:     6,
				CurrentCombo: 6,
				Perfect:      true,
				Hp:           10,
				TagByte:      0,
			},
		},
		20130815,
	)

	if err != nil {
		t.Error(err)
	}

	t.Log(data)
}

func TestEncodeMatch(t *testing.T) {
	data, err := Encode(
		BanchoMatchUpdate,
		Match{
			Id:              2,
			InProgress:      false,
			Type:            MatchTypeStandard,
			Name:            "peppy's game",
			Password:        "",
			Mods:            Hidden,
			Mode:            OsuGamemodeOsu,
			BeatmapId:       22538,
			BeatmapText:     "Peter Lambert - osu! tutorial (peppy) [Gameplay basics]",
			BeatmapChecksum: "3c8b50ebd781978beb39160c6aaf148c",
			Slots: []MatchSlot{
				{
					UserId: 2,
					Status: SlotStatusNotReady,
					Team:   SlotTeamNeutral,
					Mods:   Hidden | HardRock,
				},
			},
			HostId:      2,
			ScoringType: ScoringTypeCombo,
			TeamType:    TeamTypeHeadToHead,
			Freemod:     true,
			Seed:        0,
		},
		20130815,
	)

	if err != nil {
		t.Error(err)
	}

	t.Log(data)
}

func TestEncodeInt(t *testing.T) {
	data, err := Encode(
		BanchoLoginReply,
		int32(-1), 20130815,
	)

	if err != nil {
		t.Error(err)
	}

	t.Log(data)
}

func TestEncodeString(t *testing.T) {
	data, err := Encode(
		BanchoAnnounce,
		"Hello, World!",
		20130815,
	)

	if err != nil {
		t.Error(err)
	}

	t.Log(data)
}

func TestEncodeList(t *testing.T) {
	data, err := Encode(
		BanchoFriendsList,
		[]int32{1, 2, 3, 4, 5},
		20130815,
	)

	if err != nil {
		t.Error(err)
	}

	t.Log(data)
}
