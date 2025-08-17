package core

import (
	"context"
	"io"

	"github.com/anschmieg/gpt-cli/internal/providers"
)

// RunStreaming wires a provider adapter to the BufferManager-based StreamReader
// and returns channels for safe fragments and errors plus a cancel function.
// Defaults: fragment channel is unbuffered; errors channel is buffered (1).
func RunStreaming(ctx context.Context, adapter providers.StreamReader, prompt string) (<-chan string, <-chan error, context.CancelFunc) {
	ctx, cancel := context.WithCancel(ctx)
	frOut := make(chan string)
	errOut := make(chan error, 1)

	go func() {
		defer close(frOut)
		defer close(errOut)

		if adapter == nil {
			errOut <- io.ErrUnexpectedEOF
			return
		}

		rc, err := adapter.Stream(prompt)
		if err != nil {
			errOut <- err
			return
		}
		// ensure rc is closed when done or cancelled
		defer rc.Close()

		ch, err := StreamReader(rc)
		if err != nil {
			errOut <- err
			return
		}

		for {
			select {
			case <-ctx.Done():
				// honor cancellation
				return
			case s, ok := <-ch:
				if !ok {
					return
				}
				// forward fragment, respecting cancellation
				select {
				case <-ctx.Done():
					return
				case frOut <- s:
				}
			}
		}
	}()

	return frOut, errOut, cancel
}
