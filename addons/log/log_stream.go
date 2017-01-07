package log

import "io"

type LogStream interface {
	io.Writer
	AfterLog()
}

func NewWriterStream(w io.Writer) LogStream {
	return writerStream{w}
}

type writerStream struct {
	io.Writer
}

func (writerStream) AfterLog() {}

// TODO: implement a log rotating file stream
