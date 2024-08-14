package apigen

import (
	"fmt"
	"io"
)

type printer struct {
	w   io.Writer
	err error
}

func newPrinter(w io.Writer) *printer {
	return &printer{w: w}
}

func (p *printer) printf(f string, a ...interface{}) error {
	if p.err == nil {
		_, p.err = fmt.Fprintf(p.w, f, a...)
	}
	if p.err == nil && (len(f) == 0 || f[len(f)-1] != '\n') {
		_, p.err = p.w.Write([]byte{'\n'})
	}
	return p.err
}
