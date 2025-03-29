package chio

import (
	"bytes"
	"io"
)

// b425 adds support for user permissions
type b425 struct {
	*b402
}

func (client b425) WriteLoginPermissions(stream io.Writer, permissions uint32) error {
	writer := bytes.NewBuffer([]byte{})
	writeUint32(writer, permissions)
	return client.WritePacket(stream, BanchoLoginPermissions, writer.Bytes())
}
