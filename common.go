package chio

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"errors"
	"io"

	"github.com/bnch/uleb128"
)

func writeUint64(w io.Writer, v uint64) error {
	return binary.Write(w, binary.LittleEndian, v)
}

func writeInt64(w io.Writer, v int64) error {
	return writeUint64(w, uint64(v))
}

func writeUint32(w io.Writer, v uint32) error {
	return binary.Write(w, binary.LittleEndian, v)
}

func writeInt32(w io.Writer, v int32) error {
	return writeUint32(w, uint32(v))
}

func writeUint16(w io.Writer, v uint16) error {
	return binary.Write(w, binary.LittleEndian, v)
}

func writeInt16(w io.Writer, v int16) error {
	return writeUint16(w, uint16(v))
}

func writeUint8(w io.Writer, v uint8) error {
	return binary.Write(w, binary.LittleEndian, v)
}

func writeInt8(w io.Writer, v int8) error {
	return writeUint8(w, uint8(v))
}

func writeBoolean(w io.Writer, v bool) error {
	return binary.Write(w, binary.LittleEndian, v)
}

func writeFloat32(w io.Writer, v float32) error {
	return binary.Write(w, binary.LittleEndian, v)
}

func writeFloat64(w io.Writer, v float64) error {
	return binary.Write(w, binary.LittleEndian, v)
}

func writeIntList16(w io.Writer, v []int32) error {
	if err := writeUint16(w, uint16(len(v))); err != nil {
		return err
	}
	for _, i := range v {
		if err := writeInt32(w, i); err != nil {
			return err
		}
	}
	return nil
}

func writeIntList32(w io.Writer, v []int32) error {
	if err := writeUint32(w, uint32(len(v))); err != nil {
		return err
	}
	for _, i := range v {
		if err := writeInt32(w, i); err != nil {
			return err
		}
	}
	return nil
}

func writeString(w io.Writer, v string) error {
	if v == "" {
		binary.Write(w, binary.LittleEndian, uint8(0x00))
		return nil
	}

	if err := binary.Write(w, binary.LittleEndian, uint8(0x0b)); err != nil {
		return err
	}

	w.Write(uleb128.Marshal(len(v)))
	w.Write([]byte(v))
	return nil
}

func compressData(data []byte) []byte {
	zb := new(bytes.Buffer)
	zw := gzip.NewWriter(zb)
	zw.Write(data)
	zw.Close()
	return zb.Bytes()
}

func readUint64(r io.Reader) (v uint64, err error) {
	err = binary.Read(r, binary.LittleEndian, &v)
	return v, err
}

func readInt64(r io.Reader) (v int64, err error) {
	uv, err := readUint64(r)
	return int64(uv), err
}

func readUint32(r io.Reader) (v uint32, err error) {
	err = binary.Read(r, binary.LittleEndian, &v)
	return v, err
}

func readInt32(r io.Reader) (v int32, err error) {
	uv, err := readUint32(r)
	return int32(uv), err
}

func readUint16(r io.Reader) (v uint16, err error) {
	err = binary.Read(r, binary.LittleEndian, &v)
	return v, err
}

func readInt16(r io.Reader) (v int16, err error) {
	uv, err := readUint16(r)
	return int16(uv), err
}

func readUint8(r io.Reader) (v uint8, err error) {
	err = binary.Read(r, binary.LittleEndian, &v)
	return v, err
}

func readInt8(r io.Reader) (v int8, err error) {
	uv, err := readUint8(r)
	return int8(uv), err
}

func readBoolean(r io.Reader) (v bool, err error) {
	err = binary.Read(r, binary.LittleEndian, &v)
	return v, err
}

func readFloat32(r io.Reader) (v float32, err error) {
	err = binary.Read(r, binary.LittleEndian, &v)
	return v, err
}

func readFloat64(r io.Reader) (v float64, err error) {
	err = binary.Read(r, binary.LittleEndian, &v)
	return v, err
}

func readIntList16(r io.Reader) (v []int32, err error) {
	l, err := readUint16(r)
	if err != nil {
		return nil, err
	}

	v = make([]int32, l)
	for i := uint16(0); i < l; i++ {
		v[i], err = readInt32(r)
		if err != nil {
			return nil, err
		}
	}

	return v, nil
}

func readIntList32(r io.Reader) (v []int32, err error) {
	l, err := readUint32(r)
	if err != nil {
		return nil, err
	}

	v = make([]int32, l)
	for i := uint32(0); i < l; i++ {
		v[i], err = readInt32(r)
		if err != nil {
			return nil, err
		}
	}

	return v, nil
}

func readString(r io.Reader) (v string, err error) {
	var b uint8
	err = binary.Read(r, binary.LittleEndian, &b)
	if err != nil {
		return "", err
	}

	if b == 0x00 {
		return "", nil
	}

	if b != 0x0b {
		return "", errors.New("invalid string type")
	}

	l := uleb128.UnmarshalReader(r)
	if err != nil {
		return "", err
	}

	buf := make([]byte, l)
	_, err = r.Read(buf)
	if err != nil {
		return "", err
	}

	return string(buf), nil
}

func decompressData(data []byte) []byte {
	dst := bytes.NewBuffer([]byte{})
	zr, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return []byte{}
	}
	io.Copy(dst, zr)
	zr.Close()
	return dst.Bytes()
}
