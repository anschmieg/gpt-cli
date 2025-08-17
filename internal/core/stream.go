package core

import (
	"bufio"
	"io"
)

// StreamReader consumes an io.Reader that yields arbitrary chunks (as an
// underlying provider would stream) and uses BufferManager to produce safe
// fragments which are sent on the returned channel. The channel is closed
// when the reader reaches EOF or an error occurs (the error is returned).
func StreamReader(r io.Reader) (<-chan string, error) {
	out := make(chan string)
	br := bufio.NewReader(r)

	go func() {
		defer close(out)
		bm := NewBufferManager()
		buf := make([]byte, 1024)
		for {
			n, err := br.Read(buf)
			if n > 0 {
				chunk := string(buf[:n])
				frags := bm.AddChunk(chunk)
				for _, f := range frags {
					out <- f
				}
			}
			if err != nil {
				if err == io.EOF {
					// emit remaining buffer as best-effort
					rem := bm.ForceFlush()
					if rem != "" {
						out <- rem
					}
				}
				return
			}
		}
	}()

	return out, nil
}
