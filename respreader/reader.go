package respreader

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strconv"
)

const (
	//SimpleString const
	SimpleString = '+'
	//BulkString const
	BulkString = '$'
	//Integer const
	Integer = ':'
	//Array const
	Array = '*'
	//Error const
	Error = '-'
)

var (
	//ErrInvalidSyntax error
	ErrInvalidSyntax = errors.New("resp: invalid syntax")
)

//RESPReader reader
type RESPReader struct {
	*bufio.Reader
}

//NewReader create new reader
func NewReader(reader io.Reader) *RESPReader {
	return &RESPReader{
		Reader: bufio.NewReaderSize(reader, 4*1024),
	}
}

//ReadObject read buffer with resp structure
func (r *RESPReader) ReadObject() ([]byte, error) {
	line, err := r.readLine()
	if err != nil {
		return nil, err
	}

	switch line[0] {
	case SimpleString, Integer, Error:
		return line, nil
	case BulkString:
		return r.readBulkString(line)
	case Array:
		return r.readArray(line)
	default:
		return nil, ErrInvalidSyntax
	}
}

func (r *RESPReader) readLine() (line []byte, err error) {
	line, err = r.ReadBytes('\n')
	if err != nil {
		return nil, err
	}

	if len(line) > 1 && line[len(line)-2] == '\r' {
		return line, nil
	}

	// Line was too short or \n wasn't preceded by \r.
	return nil, ErrInvalidSyntax
}

func (r *RESPReader) readBulkString(line []byte) ([]byte, error) {
	count, err := r.getCount(line)
	if err != nil {
		return nil, err
	}
	if count == -1 {
		return line, nil
	}

	buf := make([]byte, len(line)+count+2)
	copy(buf, line)
	_, err = io.ReadFull(r, buf[len(line):])
	if err != nil {
		return nil, err
	}

	return buf, nil
}

func (r *RESPReader) getCount(line []byte) (int, error) {
	end := bytes.IndexByte(line, '\r')
	return strconv.Atoi(string(line[1:end]))
}

func (r *RESPReader) readArray(line []byte) ([]byte, error) {
	// Get number of array elements.
	count, err := r.getCount(line)
	if err != nil {
		return nil, err
	}

	// Read `count` number of RESP objects in the array.
	for i := 0; i < count; i++ {
		buf, err := r.ReadObject()
		if err != nil {
			return nil, err
		}
		line = append(line, buf...)
	}

	return line, nil
}
