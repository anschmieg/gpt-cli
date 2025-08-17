package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/anschmieg/gpt-cli/internal/core"
	"github.com/anschmieg/gpt-cli/internal/providers"
	"github.com/anschmieg/gpt-cli/internal/ui"
)

func isTTY() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}

func readPromptFromStdin() (string, error) {
	b, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func main() {
	var (
		provider = flag.String("provider", "", "provider adapter to use: sdk|http")
		apiKey   = flag.String("api-key", "", "API key for the provider (or set via env)")
		baseURL  = flag.String("base-url", "", "Provider base URL (overrides defaults)")
		model    = flag.String("model", "", "Model to use (passed to provider)")
		stream   = flag.Bool("stream", false, "display streaming output if provider supports it")
		color    = flag.String("color", "auto", "color output mode: auto|always|never")
	)
	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	var prompt string
	if flag.NArg() > 0 {
		prompt = flag.Arg(0)
	} else {
		fi, _ := os.Stdin.Stat()
		if (fi.Mode() & os.ModeCharDevice) == 0 {
			p, err := readPromptFromStdin()
			if err != nil {
				fmt.Fprintf(os.Stderr, "error reading stdin: %v\n", err)
				os.Exit(1)
			}
			prompt = p
		}
	}

	if prompt == "" {
		fmt.Fprintln(os.Stderr, "no prompt provided; pass a prompt argument or pipe into stdin")
		os.Exit(2)
	}

	if *apiKey == "" {
		if v := os.Getenv("OPENAI_API_KEY"); v != "" {
			apiKey = &v
		}
	}

	adapter := providers.NewProviderAdapter(ctx, providers.AdapterType(*provider), *apiKey, *baseURL, *model)
	if adapter == nil {
		fmt.Fprintln(os.Stderr, "failed to create provider adapter")
		os.Exit(1)
	}

	tty := isTTY()
	useGlamour := false
	switch *color {
	case "always":
		useGlamour = true
	case "never":
		useGlamour = false
	default:
		useGlamour = tty
	}
	renderer := ui.NewRenderer(useGlamour)

	if !*stream {
		if c, ok := adapter.(providers.SyncCompleter); ok {
			txt, err := c.Complete(prompt)
			if err != nil {
				fmt.Fprintf(os.Stderr, "completion error: %v\n", err)
				os.Exit(1)
			}
			fmt.Print(renderer.RenderFragment(txt))
			os.Exit(0)
		}
	}

	frCh, errCh, cancelRun := core.RunStreaming(ctx, adapter, prompt)
	defer cancelRun()

	done := make(chan struct{})

	go func() {
		for f := range frCh {
			out := renderer.RenderFragment(f)
			fmt.Print(out)
			time.Sleep(10 * time.Millisecond)
		}
		close(done)
	}()

	select {
	case err := <-errCh:
		if err != nil {
			fmt.Fprintf(os.Stderr, "stream error: %v\n", err)
			os.Exit(1)
		}
	case <-done:
	case <-ctx.Done():
	}
}
