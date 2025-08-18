package main

import (
    "fmt"
    "log"
    "os"

    "github.com/anschmieg/gpt-cli/internal/app"
    tea "github.com/charmbracelet/bubbletea"
)

// runner abstracts the tea.Program Run for testing.
type runner interface {
    Run() (tea.Model, error)
}

// newRunner constructs a default BubbleTea program runner. Tests may override.
var newRunner = func(m tea.Model) runner { return tea.NewProgram(m, tea.WithAltScreen()) }
// exitMain allows tests to intercept process exit.
var exitMain = func(code int) { os.Exit(code) }

func main() {
	// Create the application
	model := app.NewModel()

    // Create the program runner (overridable for tests)
    p := newRunner(model)

    // Run the program
    if _, err := p.Run(); err != nil {
        log.Printf("Error running program: %v", err)
        fmt.Printf("Error: %v\n", err)
        exitMain(1)
    }
}
