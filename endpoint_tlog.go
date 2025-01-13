package main

import (
	"io"
	"os"
	"time"
	"encoding/binary"
)

type TlogWriter struct {
	file *os.File
	io.Closer
}

// can't read from the log, but we don't want to error out either
func (t *TlogWriter) Read(_ []byte) (int, error) {
	// FIXME: this runs in a gorountine so it's somewhat safe to lock up
	// TODO: but we need a way to get out of there too
	for {
	}
	return 0, nil
}

func (t *TlogWriter) Write(b []byte) (n int, e error) {
	// 64 bit timestamp in microseconds (big endian)
	var usec uint64 = uint64(time.Now().UnixNano() / 1000)
	e = binary.Write(t.file, binary.BigEndian, usec)
	if e != nil {
		return 0, e
	}

	n, e  = t.file.Write(b)
	return n + 8, e
}
