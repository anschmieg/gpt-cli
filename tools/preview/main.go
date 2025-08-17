package main

import (
	"fmt"

	"github.com/anschmieg/gpt-cli/internal/ui"
)

func main() {
	sample := "\n\n    Hello\n\n\n\n      world\n\n    ```\n    code\n    ```\n\n"
	fmt.Printf("---RAW---\n%s\n---NORMALIZED---\n%s", sample, ui.RenderPlain(sample))
}
