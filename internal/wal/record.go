package wal

import (
	"encoding/binary"
	"hash/crc32"
)

// RecordType distinguishes the kind of mutation a record represents.
type RecordType uint8

const (
	RecordPut    RecordType = 1
	RecordDelete RecordType = 2
)

// Record is the in-memory form of one WAL entry.
type Record struct {
	Type  RecordType // 1 for PUT, 2 for DELETE
	Key   []byte
	Value []byte
}

// Fixed header field sizes, in bytes. Layout:
//   [ CRC32 (4) ][ length (4) ][ type (1) ][ keyLen (4) ][ key ][ value ]
//   CRC covers everything to its right (length + payload).
const (
	crcSize    = 4
	lengthSize = 4
	typeSize   = 1
	keyLenSize = 4

	headerSize = crcSize + lengthSize // bytes to read before knowing the record's length
)

var crc32Table = crc32.MakeTable(crc32.Castagnoli) // using polynomial crc32Table for all checksums.

func encodeRecord(r Record) []byte {
	keyLen := make([]byte, 4)
	binary.BigEndian.PutUint32(keyLen, uint32(len(r.Key)))

	payload := make([]byte, 0)
	payload = append(payload, byte(r.Type))
	payload = append(payload, keyLen...)
	payload = append(payload, r.Key...)
	payload = append(payload, r.Value...)

	length := make([]byte, 4)
	binary.BigEndian.PutUint32(length, uint32(len(payload)))

	// CRC covers length + payload (everything to the right of the CRC)
	crcInput := make([]byte, 0)
	crcInput = append(crcInput, length...)
	crcInput = append(crcInput, payload...)
	sum := crc32.Checksum(crcInput, crc32Table)

	encsum := make([]byte, 4)
	binary.BigEndian.PutUint32(encsum, sum)

	final := make([]byte, 0)
	final = append(final, encsum...)
	final = append(final, length...)
	final = append(final, payload...)
	return final
}
