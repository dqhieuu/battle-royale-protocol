package utils

import (
	"bytes"
	"encoding/binary"
)

func WriteAs1Byte(buffer *bytes.Buffer, val int) {
	binary.Write(buffer, binary.LittleEndian, int8(val))
}

func WriteAs2Byte(buffer *bytes.Buffer, val int) {
	binary.Write(buffer, binary.LittleEndian, int16(val))
}

func WriteAs4Byte(buffer *bytes.Buffer, val int) {
	binary.Write(buffer, binary.LittleEndian, int32(val))
}

func WriteAs8Byte(buffer *bytes.Buffer, val int) {
	binary.Write(buffer, binary.LittleEndian, int64(val))
}

func WriteBytes(buffer *bytes.Buffer, val []byte) {
	binary.Write(buffer, binary.LittleEndian, val)
}

func WriteString(buffer *bytes.Buffer, val string) {
	binary.Write(buffer, binary.LittleEndian, []byte(val))
}

func ReadAs1Byte(buffer *bytes.Buffer) int {
	var val int8
	binary.Read(buffer, binary.LittleEndian, &val)
	return int(val)
}

func ReadAs2Byte(buffer *bytes.Buffer) int {
	var val int16
	binary.Read(buffer, binary.LittleEndian, &val)
	return int(val)
}

func ReadAs4Byte(buffer *bytes.Buffer) int {
	var val int32
	binary.Read(buffer, binary.LittleEndian, &val)
	return int(val)
}

func ReadAs8Byte(buffer *bytes.Buffer) int {
	var val int64
	binary.Read(buffer, binary.LittleEndian, &val)
	return int(val)
}

func ReadBytes(buffer *bytes.Buffer) []byte {
	val := make([]byte, 1024)
	buffer.Read(val)
	return val
}

func ReadNBytes(buffer *bytes.Buffer, n int) []byte {
	val := make([]byte, n)
	binary.Read(buffer, binary.LittleEndian, &val)
	return val
}

func ReadString(buffer *bytes.Buffer) string {
	val := make([]byte, 1024)
	buffer.Read(val)
	return string(val[:bytes.IndexByte(val, 0)]) // trim all trailing 0x00
}

func ReadUntil(buffer *bytes.Buffer, delim byte) []byte {
	var val []byte
	for {
		b := buffer.Bytes()
		if len(b) == 0 {
			break
		}
		if b[0] == delim {
			break
		}
		val = append(val, b[0])
		buffer.Next(1)
	}
	return val
}
