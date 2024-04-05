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

func compressBuffer(buffer *bytes.Buffer) {
	zb := new(bytes.Buffer)
	zw := gzip.NewWriter(zb)
	zw.Write(buffer.Bytes())
	zw.Close()
	buffer.Reset()
	buffer.Write(zb.Bytes())
}

func compressData(data []byte) []byte {
	zb := new(bytes.Buffer)
	zw := gzip.NewWriter(zb)
	zw.Write(data)
	zw.Close()
	return zb.Bytes()
}

func readInt8(reader io.Reader) int8 {
	var value int8
	binary.Read(reader, binary.LittleEndian, &value)
	return value
}

func readUint8(reader io.Reader) uint8 {
	var value uint8
	binary.Read(reader, binary.LittleEndian, &value)
	return value
}

func readInt16(reader io.Reader) int16 {
	var value int16
	binary.Read(reader, binary.LittleEndian, &value)
	return value
}

func readUint16(reader io.Reader) uint16 {
	var value uint16
	binary.Read(reader, binary.LittleEndian, &value)
	return value
}

func readInt32(reader io.Reader) int32 {
	var value int32
	binary.Read(reader, binary.LittleEndian, &value)
	return value
}

func readUint32(reader io.Reader) uint32 {
	var value uint32
	binary.Read(reader, binary.LittleEndian, &value)
	return value
}

func readInt64(reader io.Reader) int64 {
	var value int64
	binary.Read(reader, binary.LittleEndian, &value)
	return value
}

func readUint64(reader io.Reader) uint64 {
	var value uint64
	binary.Read(reader, binary.LittleEndian, &value)
	return value
}

func readBool(reader io.Reader) bool {
	var value bool
	binary.Read(reader, binary.LittleEndian, &value)
	return value
}

func readFloat32(reader io.Reader) float32 {
	var value float32
	binary.Read(reader, binary.LittleEndian, &value)
	return value
}

func readFloat64(reader io.Reader) float64 {
	var value float64
	binary.Read(reader, binary.LittleEndian, &value)
	return value
}

func readString(reader io.Reader) string {
	prefix := readUint8(reader)

	if prefix == 0x00 {
		return ""
	}

	length := uleb128.UnmarshalReader(reader)

	if length == 0 {
		return ""
	}

	data := make([]byte, length)
	reader.Read(data)
	return string(data)
}

func readIntList16(reader io.Reader) []int32 {
	length := readUint16(reader)
	list := make([]int32, length)

	for i := 0; i < int(length); i++ {
		list[i] = readInt32(reader)
	}

	return list
}

func readIntList32(reader io.Reader) []int32 {
	length := readUint32(reader)
	list := make([]int32, length)

	for i := 0; i < int(length); i++ {
		list[i] = readInt32(reader)
	}

	return list
}

func decompressBuffer(buffer *bytes.Buffer) {
	dst := new(bytes.Buffer)
	zr, _ := gzip.NewReader(buffer)
	io.Copy(dst, zr)
	zr.Close()
	buffer.Reset()
	buffer.Write(dst.Bytes())
}

func decompressData(data []byte) []byte {
	dst := new(bytes.Buffer)
	zr, _ := gzip.NewReader(bytes.NewReader(data))
	io.Copy(dst, zr)
	zr.Close()
	return dst.Bytes()
}
