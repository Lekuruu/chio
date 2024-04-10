package chio

import (
	"io"
)

func writeChannel(any interface{}, writer io.Writer) {
	channel := any.(Channel)
	writeString(channel.Name, writer)
	writeString(channel.Topic, writer)
	writeInt16(channel.UserCount, writer)
}

func writeMessage(any interface{}, writer io.Writer) {
	message := any.(Message)
	writeString(message.Sender, writer)
	writeString(message.Content, writer)
	writeString(message.Target, writer)
	writeInt32(message.SenderId, writer)
}

func writePresence(any interface{}, writer io.Writer) {
	info := any.(UserInfo)

	userId := info.Id
	if info.IsIrc {
		userId = -userId
	}

	writeInt32(userId, writer)
	writeString(info.Name, writer)
	writeUint8(info.Timezone+24, writer)
	writeUint8(info.CountryIndex, writer)
	writeUint8(info.Permissions|(info.Mode<<5), writer)
	writeFloat32(info.Longitude, writer)
	writeFloat32(info.Latitude, writer)
	writeInt32(info.Rank, writer)
}

func writeStats(any interface{}, writer io.Writer) {
	info := any.(UserInfo)
	writeInt32(info.Id, writer)
	writeStatus(info.Status, writer)
	writeUint64(info.Rscore, writer)
	writeFloat32(info.Accuracy, writer)
	writeInt32(info.Playcount, writer)
	writeUint64(info.Tscore, writer)
	writeInt32(info.Rank, writer)
}

func writeStatus(any interface{}, writer io.Writer) {
	status := any.(UserStatus)
	writeUint8(status.Action, writer)
	writeString(status.Text, writer)
	writeString(status.BeatmapChecksum, writer)
	writeUint32(status.Mods, writer)
	writeUint8(status.Mode, writer)
	writeInt32(status.BeatmapId, writer)
}

func writeQuit(any interface{}, writer io.Writer) {
	quit := any.(UserQuit)
	writeInt32(quit.Info.Id, writer)
	writeUint8(quit.QuitState, writer)
}

func writeBeatmapInfo(any interface{}, writer io.Writer) {
	info := any.(BeatmapInfo)
	writeInt16(info.Index, writer)
	writeInt32(info.BeatmapId, writer)
	writeInt32(info.BeatmapSetId, writer)
	writeInt32(info.ThreadId, writer)
	writeInt8(info.RankedStatus, writer)
	writeInt8(info.OsuRank, writer)
	writeInt8(info.TaikoRank, writer)
	writeInt8(info.FruitsRank, writer)
	writeInt8(info.ManiaRank, writer)
	writeString(info.Checksum, writer)
}

func writeBeatmapInfoReply(any interface{}, writer io.Writer) {
	reply := any.(BeatmapInfoReply)
	writeInt32(int32(len(reply.Beatmaps)), writer)
	for _, info := range reply.Beatmaps {
		writeBeatmapInfo(info, writer)
	}
}

func writeReplayFrame(any interface{}, writer io.Writer) {
	frame := any.(ReplayFrame)
	writeUint8(frame.ButtonState, writer)
	writeUint8(frame.LegacyByte, writer)
	writeFloat32(frame.MouseX, writer)
	writeFloat32(frame.MouseY, writer)
	writeInt32(frame.Time, writer)
}

func writeScoreFrame(any interface{}, writer io.Writer) {
	frame := any.(ScoreFrame)
	writeInt32(frame.Time, writer)
	writeUint8(frame.Id, writer)
	writeUint16(frame.Total300, writer)
	writeUint16(frame.Total100, writer)
	writeUint16(frame.Total50, writer)
	writeUint16(frame.TotalGeki, writer)
	writeUint16(frame.TotalKatu, writer)
	writeUint16(frame.TotalMiss, writer)
	writeUint32(frame.TotalScore, writer)
	writeUint16(frame.MaxCombo, writer)
	writeUint16(frame.CurrentCombo, writer)
	writeBool(frame.Perfect, writer)
	writeUint8(frame.Hp, writer)
	writeUint8(frame.TagByte, writer)
}

func writeReplayFrameBundle(any interface{}, writer io.Writer) {
	bundle := any.(ReplayFrameBundle)
	writeInt32(bundle.Extra, writer)
	writeUint16(uint16(len(bundle.Frames)), writer)
	for _, frame := range bundle.Frames {
		writeReplayFrame(frame, writer)
	}
	writeUint8(bundle.Action, writer)
	if bundle.Frame != nil {
		writeScoreFrame(*bundle.Frame, writer)
	}
}

func writeMatch(any interface{}, writer io.Writer) {
	match := any.(Match)

	// Adjust slot size
	if len(match.Slots) != 8 && !IgnoreMatchSlotSize {
		match.Slots = match.Slots[:8]

		// Fill up with empty slots
		for i := 0; i < (8 - len(match.Slots)); i++ {
			match.Slots = append(
				match.Slots,
				MatchSlot{
					UserId: -1,
					Status: SlotStatusLocked,
					Team:   SlotTeamNeutral,
					Mods:   NoMod,
				},
			)
		}

		// TODO: Figure out when slot size was changed to 16
		//		 and add a handler for that.
	}

	writeInt32(match.Id, writer)
	writeBool(match.InProgress, writer)
	writeUint8(match.Type, writer)
	writeUint32(match.Mods, writer)
	writeString(match.Name, writer)
	writeString(match.Password, writer)
	writeString(match.BeatmapText, writer)
	writeInt32(match.BeatmapId, writer)
	writeString(match.BeatmapChecksum, writer)

	for _, slot := range match.Slots {
		writeUint8(slot.Status, writer)
	}
	for _, slot := range match.Slots {
		writeUint8(slot.Team, writer)
	}
	for _, slot := range match.Slots {
		if slot.HasPlayer() {
			writeInt32(slot.UserId, writer)
		}
	}

	writeInt32(match.HostId, writer)
	writeUint8(match.Mode, writer)
	writeUint8(match.ScoringType, writer)
	writeUint8(match.TeamType, writer)
	writeBool(match.Freemod, writer)
	if match.Freemod {
		for _, slot := range match.Slots {
			writeUint32(slot.Mods, writer)
		}
	}
	writeInt32(match.Seed, writer)
}

func writeTitleUpdate(any interface{}, writer io.Writer) {
	update := any.(TitleUpdate)
	writeString(
		(update.ImageUrl + "|" + update.RedirectUrl),
		writer,
	)
}

func readStatus(reader io.Reader) interface{} {
	return UserStatus{
		Action:          readUint8(reader),
		Text:            readString(reader),
		BeatmapChecksum: readString(reader),
		Mods:            readUint32(reader),
		Mode:            readUint8(reader),
		BeatmapId:       readInt32(reader),
	}
}

func init() {
	RegisterEncoder(BanchoLoginReply, 20130815, 232, writeInt32)
	RegisterEncoder(BanchoProtocolNegotiation, 20130815, 232, writeInt32)
	RegisterEncoder(BanchoLoginPermissions, 20130815, 232, writeInt32)
	RegisterEncoder(BanchoFriendsList, 20130815, 232, writeIntList16)
	RegisterEncoder(BanchoPing, 20130815, 232, writeNothing)
	RegisterEncoder(BanchoAnnounce, 20130815, 232, writeString)
	RegisterEncoder(BanchoGetAttention, 20130815, 232, writeNothing)
	RegisterEncoder(BanchoTitleUpdate, 20130815, 232, writeTitleUpdate)
	RegisterEncoder(BanchoMonitor, 20130815, 232, writeNothing)
	RegisterEncoder(BanchoHandleOsuUpdate, 20130815, 20130812, writeStats)
	RegisterEncoder(BanchoUserPresence, 20130815, 20121119, writePresence)
	RegisterEncoder(BanchoUserPresenceSingle, 20130815, 232, writeInt32)
	RegisterEncoder(BanchoUserPresenceBundle, 20130815, 1700, writeIntList16)
	RegisterEncoder(BanchoHandleOsuQuit, 20130815, 1700, writeQuit)
	RegisterEncoder(BanchoHandleIrcQuit, 20130815, 232, writeString)
	RegisterEncoder(BanchoHandleIrcChangeUsername, 20130815, 232, writeString)
	RegisterEncoder(BanchoChannelAvailable, 20130815, 20120725, writeChannel)
	RegisterEncoder(BanchoChannelAvailableAutojoin, 20130815, 20120725, writeChannel)
	RegisterEncoder(BanchoChannelInfoComplete, 20130815, 232, writeNothing)
	RegisterEncoder(BanchoChannelJoinSuccess, 20130815, 232, writeString)
	RegisterEncoder(BanchoChannelRevoked, 20130815, 232, writeString)
	RegisterEncoder(BanchoSendMessage, 20130815, 20121223, writeMessage)
	RegisterEncoder(BanchoSpectatorJoined, 20130815, 232, writeInt32)
	RegisterEncoder(BanchoSpectatorLeft, 20130815, 232, writeInt32)
	RegisterEncoder(BanchoSpectateFrames, 20130815, 20130329, writeReplayFrameBundle)
	RegisterEncoder(BanchoSpectatorCantSpectate, 20130815, 232, writeInt32)
	RegisterEncoder(BanchoFellowSpectatorJoined, 20130815, 232, writeInt32)
	RegisterEncoder(BanchoFellowSpectatorLeft, 20130815, 232, writeInt32)
	RegisterEncoder(BanchoLobbyJoin, 20130815, 232, writeInt32)
	RegisterEncoder(BanchoLobbyPart, 20130815, 232, writeInt32)
	RegisterEncoder(BanchoMatchNew, 20130815, 20120812, writeMatch)
	RegisterEncoder(BanchoMatchUpdate, 20130815, 20120812, writeMatch)
	RegisterEncoder(BanchoMatchDisband, 20130815, 232, writeInt32)
	RegisterEncoder(BanchoMatchJoinSuccess, 20130815, 20120812, writeMatch)
	RegisterEncoder(BanchoMatchJoinFail, 20130815, 232, writeNothing)
	RegisterEncoder(BanchoMatchChangePassword, 20130815, 232, writeString)
	RegisterEncoder(BanchoMatchStart, 20130815, 20120812, writeMatch)
	RegisterEncoder(BanchoMatchScoreUpdate, 20130815, 483, writeScoreFrame)
	RegisterEncoder(BanchoMatchTransferHost, 20130815, 232, writeNothing)
	RegisterEncoder(BanchoMatchAllPlayersLoaded, 20130815, 232, writeNothing)
	RegisterEncoder(BanchoMatchPlayerFailed, 20130815, 232, writeInt32)
	RegisterEncoder(BanchoMatchPlayerSkipped, 20130815, 232, writeInt32)
	RegisterEncoder(BanchoMatchSkip, 20130815, 232, writeNothing)
	RegisterEncoder(BanchoMatchComplete, 20130815, 232, writeNothing)
	RegisterEncoder(BanchoInvite, 20130815, 20121223, writeMessage)
	RegisterEncoder(BanchoBeatmapInfoReply, 20130815, 20121008, writeBeatmapInfoReply)
	RegisterEncoder(BanchoSilenceInfo, 20130815, 232, writeInt32)
	RegisterEncoder(BanchoUserSilenced, 20130815, 232, writeInt32)
	RegisterEncoder(BanchoUserDMsBlocked, 20130815, 232, writeMessage)
	RegisterEncoder(BanchoTargetIsSilenced, 20130815, 232, writeMessage)
	RegisterEncoder(BanchoVersionUpdate, 20130815, 232, writeNothing)
	RegisterEncoder(BanchoVersionUpdateForced, 20130815, 232, writeNothing)
	RegisterEncoder(BanchoSwitchServer, 20130815, 232, writeString)
	RegisterEncoder(BanchoRestart, 20130815, 232, writeInt32)
	RegisterEncoder(BanchoAccountRestricted, 20130815, 232, writeNothing)
	RegisterEncoder(BanchoRTX, 20130815, 232, writeString)
	RegisterEncoder(BanchoMatchAbort, 20130815, 232, writeNothing)
	RegisterEncoder(BanchoSwitchTournamentServer, 20130815, 232, writeString)

	RegisterDecoder(OsuSendUserStatus, 20130815, 232, readStatus)
}
