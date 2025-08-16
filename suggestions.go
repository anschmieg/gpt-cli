package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

// SuggestionResponse represents the JSON structure for shell suggestions
type SuggestionResponse struct {
	Suggestions []Suggestion `json:"suggestions"`
	Context     string       `json:"context,omitempty"`
	Safe        bool         `json:"safe"`
}

// Suggestion represents a single shell command suggestion
type Suggestion struct {
	Command     string   `json:"command"`
	Description string   `json:"description"`
	Category    string   `json:"category,omitempty"`
	Risk        string   `json:"risk"` // "low", "medium", "high"
	Args        []string `json:"args,omitempty"`
}

// generateSuggestionPrompt creates a specialized system prompt for shell suggestions
func generateSuggestionPrompt() string {
	return `You are a shell command assistant. Your task is to suggest shell commands based on user requests.

CRITICAL RULES:
1. Output ONLY valid JSON in the exact format specified below
2. Never suggest commands that could be dangerous (rm -rf, dd, formatting commands, etc.)
3. Mark risk levels appropriately: "low" for safe commands, "medium" for commands that modify files, "high" for potentially dangerous commands
4. Provide 1-3 relevant suggestions maximum
5. Include clear descriptions of what each command does

Output format (JSON only, no markdown, no extra text):
{
  "suggestions": [
    {
      "command": "actual shell command",
      "description": "clear explanation of what this does", 
      "category": "file_management|system_info|network|development|etc",
      "risk": "low|medium|high",
      "args": ["arg1", "arg2"]
    }
  ],
  "context": "brief context about the task",
  "safe": true/false
}

Example request: "list files in current directory"
Example response:
{
  "suggestions": [
    {
      "command": "ls -la",
      "description": "List all files and directories with detailed information including hidden files",
      "category": "file_management", 
      "risk": "low",
      "args": ["-la"]
    }
  ],
  "context": "Listing directory contents",
  "safe": true
}`
}

// parseSuggestionResponse parses the AI response and validates it as suggestion JSON
func parseSuggestionResponse(response string) (*SuggestionResponse, error) {
	// Clean up the response - remove any markdown formatting
	cleaned := strings.TrimSpace(response)
	
	// Remove markdown code blocks if present
	if strings.HasPrefix(cleaned, "```json") {
		cleaned = strings.TrimPrefix(cleaned, "```json")
		cleaned = strings.TrimSuffix(cleaned, "```")
		cleaned = strings.TrimSpace(cleaned)
	} else if strings.HasPrefix(cleaned, "```") {
		cleaned = strings.TrimPrefix(cleaned, "```")
		cleaned = strings.TrimSuffix(cleaned, "```")
		cleaned = strings.TrimSpace(cleaned)
	}
	
	var suggestions SuggestionResponse
	if err := json.Unmarshal([]byte(cleaned), &suggestions); err != nil {
		return nil, fmt.Errorf("failed to parse suggestion response as JSON: %w", err)
	}
	
	// Validate the response
	if len(suggestions.Suggestions) == 0 {
		return nil, fmt.Errorf("no suggestions found in response")
	}
	
	// Validate each suggestion
	for i, suggestion := range suggestions.Suggestions {
		if suggestion.Command == "" {
			return nil, fmt.Errorf("suggestion %d has empty command", i)
		}
		if suggestion.Description == "" {
			return nil, fmt.Errorf("suggestion %d has empty description", i)
		}
		if suggestion.Risk == "" {
			suggestions.Suggestions[i].Risk = "medium" // Default to medium risk
		}
	}
	
	return &suggestions, nil
}

// formatSuggestionOutput formats the suggestions for console output
func formatSuggestionOutput(suggestions *SuggestionResponse) string {
	output, err := json.MarshalIndent(suggestions, "", "  ")
	if err != nil {
		return fmt.Sprintf(`{"error": "failed to format suggestions: %v"}`, err)
	}
	return string(output)
}

// runSuggestionMode runs the CLI in shell suggestion mode
func runSuggestionMode(config *CoreConfig, providerOpts *ProviderOptions) error {
	// Override system prompt for suggestions
	originalSystem := config.System
	config.System = generateSuggestionPrompt()
	
	// Disable streaming for suggestions to get complete JSON response
	originalStream := config.Stream
	config.Stream = false
	
	// Get the response from the provider
	response, err := callProvider(config, providerOpts)
	if err != nil {
		// Restore original config
		config.System = originalSystem
		config.Stream = originalStream
		return err
	}
	
	// Restore original config
	config.System = originalSystem
	config.Stream = originalStream
	
	// Choose the response text (prefer markdown for richer content)
	responseText := response.Markdown
	if responseText == "" {
		responseText = response.Text
	}
	
	// Parse and validate the suggestion response
	suggestions, err := parseSuggestionResponse(responseText)
	if err != nil {
		// If parsing fails, return a safe error response
		errorResponse := &SuggestionResponse{
			Suggestions: []Suggestion{
				{
					Command:     "echo 'Unable to generate safe suggestions'",
					Description: "The AI was unable to generate safe command suggestions for your request",
					Category:    "error",
					Risk:        "low",
					Args:        []string{"'Unable to generate safe suggestions'"},
				},
			},
			Context: "Error parsing AI response",
			Safe:    true,
		}
		fmt.Print(formatSuggestionOutput(errorResponse))
		return nil
	}
	
	// Output the formatted suggestions
	fmt.Print(formatSuggestionOutput(suggestions))
	return nil
}