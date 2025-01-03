package chio

import (
	"bytes"
	"fmt"
	"io"
)

// b296 adds the "Time" value to score frames
type b296 struct {
	*b294
}

func (client *b296) Clone() BanchoIO {
	previous := &b294{}
	return &b296{previous.Clone().(*b294)}
}

func (client *b296) WritePacket(packetId uint16, data []byte) error {
	// Convert packetId back for the client
	packetId = client.ConvertOutputPacketId(packetId)
	writer := bytes.NewBuffer([]byte{})

	err := writeUint16(writer, packetId)
	if err != nil {
		return err
	}

	compressed := compressData(data)
	err = writeUint32(writer, uint32(len(compressed)))
	if err != nil {
		return err
	}

	_, err = writer.Write(compressed)
	if err != nil {
		return err
	}

	_, err = client.Write(writer.Bytes())
	return err
}

func (client *b296) ReadPacket() (packet *BanchoPacket, err error) {
	packet = &BanchoPacket{}
	packet.Id, err = readUint16(client.stream)
	if err != nil {
		return nil, err
	}

	// Convert packet ID to a usable value
	packet.Id = client.ConvertInputPacketId(packet.Id)

	if !client.ImplementsPacket(packet.Id) {
		return nil, fmt.Errorf("packet '%d' not implemented", packet.Id)
	}

	length, err := readInt32(client.stream)
	if err != nil {
		return nil, err
	}

	compressedData := make([]byte, length)
	n, err := client.stream.Read(compressedData)
	if err != nil {
		return nil, err
	}

	if n != int(length) {
		return nil, fmt.Errorf("expected %d bytes, got %d", length, n)
	}

	data, err := decompressData(compressedData)
	if err != nil {
		return nil, err
	}

	packet.Data, err = client.ReadPacketType(packet.Id, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	return packet, nil
}

func (client *b296) ReadPacketType(packetId uint16, reader io.Reader) (any, error) {
	switch packetId {
	case OsuSendUserStatus:
		return client.ReadStatus(reader)
	case OsuSendIrcMessage:
		return client.ReadMessage(reader)
	case OsuSendIrcMessagePrivate:
		return client.ReadMessagePrivate(reader)
	case OsuStartSpectating:
		return readUint32(reader)
	case OsuSpectateFrames:
		return client.ReadFrameBundle(reader)
	case OsuErrorReport:
		return readString(reader)
	default:
		return nil, nil
	}
}

func (client *b296) WriteSpectateFrames(bundle ReplayFrameBundle) error {
	writer := bytes.NewBuffer([]byte{})
	writeUint16(writer, uint16(len(bundle.Frames)))

	for _, frame := range bundle.Frames {
		// Convert button state
		leftMouse := ButtonStateLeft1&frame.ButtonState > 0 || ButtonStateLeft2&frame.ButtonState > 0
		rightMouse := ButtonStateRight1&frame.ButtonState > 0 || ButtonStateRight2&frame.ButtonState > 0

		writeBoolean(writer, leftMouse)
		writeBoolean(writer, rightMouse)
		writeFloat32(writer, frame.MouseX)
		writeFloat32(writer, frame.MouseY)
		writeInt32(writer, frame.Time)
	}

	writeUint8(writer, bundle.Action)

	if bundle.Frame != nil {
		client.WriteScoreFrame(writer, *bundle.Frame)
	}

	return client.WritePacket(BanchoSpectateFrames, writer.Bytes())
}

func (client *b296) WriteScoreFrame(writer io.Writer, frame ScoreFrame) error {
	writeString(writer, frame.Checksum())
	writeInt32(writer, frame.Time)
	writeUint8(writer, frame.Id)
	writeUint16(writer, frame.Total300)
	writeUint16(writer, frame.Total100)
	writeUint16(writer, frame.Total50)
	writeUint16(writer, frame.TotalGeki)
	writeUint16(writer, frame.TotalKatu)
	writeUint16(writer, frame.TotalMiss)
	writeUint32(writer, frame.TotalScore)
	writeUint16(writer, frame.MaxCombo)
	writeUint16(writer, frame.CurrentCombo)
	writeBoolean(writer, frame.Perfect)
	writeUint8(writer, frame.Hp)
	return nil
}

func (client *b296) ReadFrameBundle(reader io.Reader) (*ReplayFrameBundle, error) {
	count, err := readUint16(reader)
	if err != nil {
		return nil, err
	}

	frames := make([]*ReplayFrame, count)
	for i := 0; i < int(count); i++ {
		frame, err := client.ReadReplayFrame(reader)
		if err != nil {
			return nil, err
		}
		frames[i] = frame
	}

	action, err := readUint8(reader)
	if err != nil {
		return nil, err
	}

	frame, err := client.ReadScoreFrame(reader)
	if err != nil && err.Error() != "EOF" {
		return nil, err
	}

	return &ReplayFrameBundle{Frames: frames, Action: action, Frame: frame}, nil
}

func (client b296) ReadScoreFrame(reader io.Reader) (*ScoreFrame, error) {
	var err error
	errors := NewErrorCollection()
	frame := &ScoreFrame{}
	_, err = readString(reader) // Checksum
	errors.Add(err)
	frame.Time, err = readInt32(reader)
	errors.Add(err)
	frame.Id, err = readUint8(reader)
	errors.Add(err)
	frame.Total300, err = readUint16(reader)
	errors.Add(err)
	frame.Total100, err = readUint16(reader)
	errors.Add(err)
	frame.Total50, err = readUint16(reader)
	errors.Add(err)
	frame.TotalGeki, err = readUint16(reader)
	errors.Add(err)
	frame.TotalKatu, err = readUint16(reader)
	errors.Add(err)
	frame.TotalMiss, err = readUint16(reader)
	errors.Add(err)
	frame.TotalScore, err = readUint32(reader)
	errors.Add(err)
	frame.MaxCombo, err = readUint16(reader)
	errors.Add(err)
	frame.CurrentCombo, err = readUint16(reader)
	errors.Add(err)
	frame.Perfect, err = readBoolean(reader)
	errors.Add(err)
	frame.Hp, err = readUint8(reader)
	errors.Add(err)
	return frame, errors.Next()
}
