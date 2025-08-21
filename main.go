package main

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/anschmieg/gpt-cli/internal/app"
)

func main() {
	// Create the application
	model := app.NewModel()
	
	// Create the program with alt screen
	p := tea.NewProgram(model, tea.WithAltScreen())
	
	// Run the program
	if _, err := p.Run(); err != nil {
		log.Printf("Error running program: %v", err)
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}