package chio

import (
	"bytes"
	"io"
)

// b504 adds the taiko & fruits ranks to the beatmap info packet
type b504 struct {
	*b490
	protocolVersion int
}

func (client *b504) ProtocolVersion() int {
	return client.protocolVersion
}

func (client *b504) OverrideProtocolVersion(version int) {
	client.protocolVersion = version
}

func (client *b504) WriteBeatmapInfoReply(stream io.Writer, reply BeatmapInfoReply) error {
	buffer := bytes.NewBuffer([]byte{})
	writeInt32(buffer, int32(len(reply.Beatmaps)))

	for _, info := range reply.Beatmaps {
		writeInt16(buffer, info.Index)
		writeInt32(buffer, info.BeatmapId)
		writeInt32(buffer, info.BeatmapSetId)
		writeInt32(buffer, info.ThreadId)
		writeInt8(buffer, info.RankedStatus)
		writeInt8(buffer, info.OsuRank)

		if client.ProtocolVersion() >= 2 {
			writeInt8(buffer, info.TaikoRank)
			writeInt8(buffer, info.FruitsRank)
		}

		writeString(buffer, info.Checksum)
	}

	return nil
}
