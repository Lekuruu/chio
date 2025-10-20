package chio

import (
	"bytes"
	"io"
)

// b365 adds a level display on the user panel, which has a bug that causes
// the client to crash, when the user has a very high total score.
type b365 struct {
	*b354
}

func (client *b365) WriteUserStats(stream io.Writer, info UserInfo) error {
	writer := bytes.NewBuffer([]byte{})

	if info.Presence.IsIrc {
		writeString(writer, info.Name)
		return client.WritePacket(stream, BanchoHandleIrcJoin, writer.Bytes())
	}

	// Cap total score to prevent client crash
	originalTscore := info.Stats.Tscore
	if info.Stats.Tscore > 17705429348 {
		info.Stats.Tscore = 17705429348
	}

	client.WriteStats(writer, info)
	err := client.WritePacket(stream, BanchoHandleOsuUpdate, writer.Bytes())

	// Restore original value
	info.Stats.Tscore = originalTscore
	return err
}

func (client *b365) WriteUserPresence(stream io.Writer, info UserInfo) error {
	writer := bytes.NewBuffer([]byte{})

	// Cap total score to prevent client crash
	originalTscore := info.Stats.Tscore
	if info.Stats.Tscore > 17705429348 {
		info.Stats.Tscore = 17705429348
	}

	writeUint32(writer, uint32(info.Id))
	writeString(writer, info.Name)
	writeInt8(writer, info.Presence.Timezone)
	writeInt8(writer, info.Presence.CountryIndex)
	writeUint8(writer, info.Presence.Permissions)
	writeFloat32(writer, info.Presence.Longitude)
	writeFloat32(writer, info.Presence.Latitude)
	writeUint32(writer, uint32(info.Stats.Rank))

	err := client.WritePacket(stream, BanchoUserPresence, writer.Bytes())

	// Restore original value
	info.Stats.Tscore = originalTscore
	return err
}
