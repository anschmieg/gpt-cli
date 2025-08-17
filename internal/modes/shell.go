package modes

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/anschmieg/gpt-cli/internal/config"
	"github.com/anschmieg/gpt-cli/internal/providers"
	"github.com/anschmieg/gpt-cli/internal/ui"
)

// ShellSuggestion represents a shell command suggestion with safety information
type ShellSuggestion struct {
	Command     string `json:"command"`
	SafetyLevel string `json:"safety_level"` // safe, moderate, dangerous
	Explanation string `json:"explanation"`
	Reasoning   string `json:"reasoning"`
}

// ShellMode handles shell command suggestion functionality
type ShellMode struct {
	config   *config.Config
	provider providers.Provider
	ui       *ui.UI
}

// NewShellMode creates a new shell mode instance
func NewShellMode(config *config.Config, provider providers.Provider, ui *ui.UI) *ShellMode {
	return &ShellMode{
		config:   config,
		provider: provider,
		ui:       ui,
	}
}

// SuggestCommand generates a shell command suggestion for the given prompt
func (s *ShellMode) SuggestCommand(prompt string) (*ShellSuggestion, error) {
	systemPrompt := `You are a shell command assistant. Given a user request, suggest a bash/shell command that accomplishes their goal.

IMPORTANT: You must respond with ONLY a valid JSON object in this exact format:
{
  "command": "the actual shell command",
  "safety_level": "safe|moderate|dangerous",
  "explanation": "brief explanation of what the command does",
  "reasoning": "explanation of why this safety level was assigned"
}

Safety level guidelines:
- "safe": Commands that only read data, display information, or perform non-destructive operations
- "moderate": Commands that modify files/directories in controlled ways, install packages, or change configuration
- "dangerous": Commands that can delete data, modify system files, change permissions, or affect system security

Examples:
- "ls -la" = safe (only lists files)
- "mkdir project" = moderate (creates directory)
- "rm -rf /" = dangerous (deletes everything)

User request: ` + prompt

	// Create a modified provider call with shell-specific system prompt
	originalSystem := s.config.System
	s.config.System = systemPrompt
	s.config.Temperature = 0.1 // Lower temperature for more consistent output
	
	// Restore original config after call
	defer func() {
		s.config.System = originalSystem
	}()

	response, err := s.provider.CallProvider(prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to get command suggestion: %w", err)
	}

	suggestion, err := s.parseShellSuggestion(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse command suggestion: %w", err)
	}

	return suggestion, nil
}

// parseShellSuggestion parses the LLM response into a ShellSuggestion
func (s *ShellMode) parseShellSuggestion(response string) (*ShellSuggestion, error) {
	// Clean the response - remove markdown code blocks if present
	cleaned := strings.TrimSpace(response)
	
	// Remove markdown code blocks
	if strings.HasPrefix(cleaned, "```json") {
		cleaned = strings.TrimPrefix(cleaned, "```json")
	}
	if strings.HasPrefix(cleaned, "```") {
		cleaned = strings.TrimPrefix(cleaned, "```")
	}
	if strings.HasSuffix(cleaned, "```") {
		cleaned = strings.TrimSuffix(cleaned, "```")
	}
	
	// Extract JSON from the response if it's embedded in other text
	jsonRegex := regexp.MustCompile(`\{[^{}]*(?:\{[^{}]*\}[^{}]*)*\}`)
	jsonMatch := jsonRegex.Find([]byte(cleaned))
	if jsonMatch != nil {
		cleaned = string(jsonMatch)
	}

	var suggestion ShellSuggestion
	err := json.Unmarshal([]byte(cleaned), &suggestion)
	if err != nil {
		return nil, fmt.Errorf("invalid JSON response: %w\nResponse: %s", err, cleaned)
	}

	// Validate required fields
	if suggestion.Command == "" {
		return nil, fmt.Errorf("command field is required")
	}
	if suggestion.SafetyLevel == "" {
		return nil, fmt.Errorf("safety_level field is required")
	}
	if suggestion.Explanation == "" {
		return nil, fmt.Errorf("explanation field is required")
	}

	// Validate safety level
	validSafetyLevels := map[string]bool{
		"safe":      true,
		"moderate":  true,
		"dangerous": true,
	}
	if !validSafetyLevels[suggestion.SafetyLevel] {
		return nil, fmt.Errorf("invalid safety_level: %s", suggestion.SafetyLevel)
	}

	return &suggestion, nil
}

// InteractiveMode runs the shell suggestion mode interactively
func (s *ShellMode) InteractiveMode(prompt string) error {
	// Get command suggestion
	fmt.Println("🤖 Generating shell command suggestion...")
	suggestion, err := s.SuggestCommand(prompt)
	if err != nil {
		return err
	}

	// Display the suggestion
	s.displaySuggestion(suggestion)

	// Prompt user for action
	return s.promptUserAction(suggestion)
}

// displaySuggestion displays the command suggestion in a formatted way
func (s *ShellMode) displaySuggestion(suggestion *ShellSuggestion) {
	fmt.Println("\n" + s.ui.TitleStyle.Render("💡 Command Suggestion"))
	
	// Command
	fmt.Printf("\n%s\n", s.ui.PromptStyle.Render("Command:"))
	fmt.Printf("  %s\n", s.ui.InputStyle.Render(suggestion.Command))
	
	// Safety level with color coding
	fmt.Printf("\n%s\n", s.ui.PromptStyle.Render("Safety Level:"))
	safetyColor := s.getSafetyColor(suggestion.SafetyLevel)
	fmt.Printf("  %s\n", safetyColor.Render(strings.ToUpper(suggestion.SafetyLevel)))
	
	// Explanation
	fmt.Printf("\n%s\n", s.ui.PromptStyle.Render("What it does:"))
	fmt.Printf("  %s\n", suggestion.Explanation)
	
	// Reasoning if provided
	if suggestion.Reasoning != "" {
		fmt.Printf("\n%s\n", s.ui.PromptStyle.Render("Safety reasoning:"))
		fmt.Printf("  %s\n", suggestion.Reasoning)
	}
}

// getSafetyColor returns appropriate color styling for safety level
func (s *ShellMode) getSafetyColor(level string) lipgloss.Style {
	switch level {
	case "safe":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#059669")) // Green
	case "moderate":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#EAB308")) // Yellow
	case "dangerous":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#DC2626")) // Red
	default:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280")) // Gray
	}
}

// promptUserAction prompts the user for what to do with the suggestion
func (s *ShellMode) promptUserAction(suggestion *ShellSuggestion) error {
	fmt.Printf("\n%s\n", s.ui.PromptStyle.Render("What would you like to do?"))
	fmt.Println("  [e] Execute immediately")
	fmt.Println("  [m] Manually edit then execute") 
	fmt.Println("  [r] Refine the suggestion")
	fmt.Println("  [a] Abort")
	fmt.Print("\nChoice [e/m/r/a]: ")

	var choice string
	fmt.Scanln(&choice)
	choice = strings.ToLower(strings.TrimSpace(choice))

	switch choice {
	case "e", "execute":
		return s.executeCommand(suggestion.Command)
	case "m", "edit":
		return s.editAndExecute(suggestion.Command)
	case "r", "refine":
		return s.refineCommand(suggestion)
	case "a", "abort":
		fmt.Println("Command suggestion aborted.")
		return nil
	default:
		fmt.Printf("Invalid choice: %s. Please choose e, m, r, or a.\n", choice)
		return s.promptUserAction(suggestion)
	}
}

// executeCommand executes the command directly
func (s *ShellMode) executeCommand(command string) error {
	fmt.Printf("\n%s %s\n", s.ui.LoadingStyle.Render("🚀 Executing:"), command)
	
	cmd := exec.Command("sh", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	
	return cmd.Run()
}

// editAndExecute allows the user to edit the command before execution
func (s *ShellMode) editAndExecute(command string) error {
	fmt.Printf("\n%s\n", s.ui.PromptStyle.Render("Edit the command (or press Enter to use as-is):"))
	fmt.Printf("Current: %s\n", command)
	fmt.Print("Edited:  ")
	
	var editedCommand string
	fmt.Scanln(&editedCommand)
	
	if strings.TrimSpace(editedCommand) == "" {
		editedCommand = command
	}
	
	return s.executeCommand(editedCommand)
}

// refineCommand allows the user to provide additional context to refine the suggestion
func (s *ShellMode) refineCommand(suggestion *ShellSuggestion) error {
	fmt.Printf("\n%s\n", s.ui.PromptStyle.Render("Please provide additional context or corrections:"))
	fmt.Print("Refinement: ")
	
	var refinement string
	fmt.Scanln(&refinement)
	
	if strings.TrimSpace(refinement) == "" {
		fmt.Println("No refinement provided. Returning to original suggestion.")
		return s.promptUserAction(suggestion)
	}
	
	// Create a new prompt with the refinement
	newPrompt := fmt.Sprintf("Original command suggestion: %s\nUser feedback: %s\nPlease provide an improved command suggestion.", suggestion.Command, refinement)
	
	fmt.Println("\n🤖 Generating refined suggestion...")
	newSuggestion, err := s.SuggestCommand(newPrompt)
	if err != nil {
		fmt.Printf("Error refining command: %v\n", err)
		return s.promptUserAction(suggestion) // Fallback to original
	}
	
	s.displaySuggestion(newSuggestion)
	return s.promptUserAction(newSuggestion)
}