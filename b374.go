package chio

import (
	"io"
)

// b374 adds the ButtonState constant, which deprecates
// the old button left/right booleans.
type b374 struct {
	*b365
}

func (client *b374) WriteReplayFrame(writer io.Writer, frame ReplayFrame) error {
	writeUint8(writer, frame.ButtonState)
	writeBoolean(writer, false)
	writeFloat32(writer, frame.MouseX)
	writeFloat32(writer, frame.MouseY)
	writeInt32(writer, frame.Time)
	return nil
}

func (client *b374) ReadReplayFrame(reader io.Reader) (*ReplayFrame, error) {
	var err error
	errors := NewErrorCollection()
	frame := &ReplayFrame{}

	frame.ButtonState, err = readUint8(reader)
	errors.Add(err)

	legacyMouseRight, err := readBoolean(reader)
	errors.Add(err)

	frame.MouseX, err = readFloat32(reader)
	errors.Add(err)

	frame.MouseY, err = readFloat32(reader)
	errors.Add(err)

	frame.Time, err = readInt32(reader)
	errors.Add(err)

	if legacyMouseRight {
		frame.ButtonState |= ButtonStateRight1
	}

	return frame, errors.Next()
}
