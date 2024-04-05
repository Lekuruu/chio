package chio

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"io"

	"github.com/bnch/uleb128"
)

func writeNothing(any interface{}, writer io.Writer) {}

func writeInt64(any interface{}, writer io.Writer) {
	value := any.(int64)
	binary.Write(writer, binary.LittleEndian, value)
}

func writeUint64(any interface{}, writer io.Writer) {
	value := any.(uint64)
	binary.Write(writer, binary.LittleEndian, value)
}

func writeInt32(any interface{}, writer io.Writer) {
	value := any.(int32)
	binary.Write(writer, binary.LittleEndian, value)
}

func writeUint32(any interface{}, writer io.Writer) {
	value := any.(uint32)
	binary.Write(writer, binary.LittleEndian, value)
}

func writeInt16(any interface{}, writer io.Writer) {
	value := any.(int16)
	binary.Write(writer, binary.LittleEndian, value)
}

func writeUint16(any interface{}, writer io.Writer) {
	value := any.(uint16)
	binary.Write(writer, binary.LittleEndian, value)
}

func writeInt8(any interface{}, writer io.Writer) {
	value := any.(int8)
	binary.Write(writer, binary.LittleEndian, value)
}

func writeUint8(any interface{}, writer io.Writer) {
	value := any.(uint8)
	binary.Write(writer, binary.LittleEndian, value)
}

func writeBool(any interface{}, writer io.Writer) {
	value := any.(bool)
	binary.Write(writer, binary.LittleEndian, value)
}

func writeFloat32(any interface{}, writer io.Writer) {
	value := any.(float32)
	binary.Write(writer, binary.LittleEndian, value)
}

func writeFloat64(any interface{}, writer io.Writer) {
	value := any.(float64)
	binary.Write(writer, binary.LittleEndian, value)
}

func writeString(any interface{}, writer io.Writer) {
	str := any.(string)

	if str == "" {
		binary.Write(writer, binary.LittleEndian, uint8(0x00))
		return
	}

	binary.Write(writer, binary.LittleEndian, uint8(0x0b))
	length := uleb128.Marshal(len(str))

	writer.Write(length)
	writer.Write([]byte(str))
}

func writeIntList16(any interface{}, writer io.Writer) {
	list := any.([]int32)
	binary.Write(writer, binary.LittleEndian, uint16(len(list)))

	for _, value := range list {
		binary.Write(writer, binary.LittleEndian, value)
	}
}

func writeIntList32(any interface{}, writer io.Writer) {
	list := any.([]int32)
	binary.Write(writer, binary.LittleEndian, uint32(len(list)))

	for _, value := range list {
		binary.Write(writer, binary.LittleEndian, value)
	}
}

func compressBuffer(buffer bytes.Buffer) {
	zb := new(bytes.Buffer)
	zw := gzip.NewWriter(zb)
	zw.Write(buffer.Bytes())
	zw.Close()
	buffer.Reset()
	buffer.Write(zb.Bytes())
}
