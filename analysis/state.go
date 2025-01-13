package analysis

import (
	"exoconfiglsp/lsp"
	"fmt"
	"strings"
	"regexp"
	"os"
	"log"
	"encoding/json"
	"path/filepath"
)

type State struct {
	// Map of file names to contents
	Documents map[string]string
}

func NewState() State {
	return State{Documents: map[string]string{}}
}

func getLogger(filename string) *log.Logger {
	logfile, err := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	if err != nil {
		panic("hey, you didnt give me a good file")
	}

	return log.New(logfile, "[ExoConfigLSP]", log.Ldate|log.Ltime|log.Lshortfile)
}

func isValidDataType(data_type string) bool {
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	logger := getLogger(dir + "/log.txt")
	file_dir := dir + "/analysis/ExoTypesUnits.json"
	fileData, err := os.ReadFile(file_dir)
	if err != nil {
		logger.Printf("Failed to open ExoTypesUnits.json")
		return false
	}

	var data_types map[string]any
	err = json.Unmarshal(fileData, &data_types)
	if err != nil {
		logger.Printf("Failed to parse ExoTypesUnits.json")
		return false
	}

	for k,_ := range data_types {
		if k == data_type { 
			return true
		}
	}
	logger.Println("No Match")
	return false
}

func isValidDataUnit(data_type string, data_unit string) bool {
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	logger := getLogger(dir + "/log.txt")
	file_dir := dir + "/analysis/ExoTypesUnits.json"
	fileData, err := os.ReadFile(file_dir)
	if err != nil {
		logger.Printf("Failed to open ExoTypesUnits.json")
		return false
	}

	var data_types map[string][]string
	err = json.Unmarshal(fileData, &data_types)
	if err != nil {
		logger.Printf("Failed to parse ExoTypesUnits.json")
		return false
	}
	value, _ := data_types[data_type]
	for v := range value {
		if value[v] == data_unit { 
			return true
		}
	}
	logger.Println("No Match")
	return false
}

func getDiagnosticsForFile(text string) []lsp.Diagnostic {
	diagnostics := []lsp.Diagnostic{}
	prev_line := ""
	for row, line := range strings.Split(text, "\n") {
		if (strings.Contains(line, "data_type")) {
			idx := strings.Index(line, "data_type")
			re := regexp.MustCompile(`"([^"]+)"`)
			_ = re
			matches := re.FindAllStringSubmatch(line, -1)

			// Check if there are at least two matches and print the second one
			if len(matches) < 2 {
				break
			}
			unit := matches[1][1] // matches[1][1] contains the second value inside quotes
			if !isValidDataType(unit) {
				_ = row
				_ = idx
			       diagnostics = append(diagnostics, lsp.Diagnostic{
					Range:    LineRange(row, idx+1, idx+len("data_type")),
					Severity: 1,
					Source:   "Exo",
					Message:  unit + " not a valid data_type",
				})
			}
			
		}

		if (strings.Contains(line, "data_unit")) {
			idx := strings.Index(line, "data_unit")
			re := regexp.MustCompile(`"([^"]+)"`)
			_ = re
			matches := re.FindAllStringSubmatch(line, -1)

			// Check if there are at least two matches and print the second one
			if len(matches) < 2 {
				break
			}
			///////////////////////////////////////////
			data_unit := matches[1][1] // matches[1][1] contains the second value inside quotes

			idx = strings.Index(prev_line, "data_type")
			re = regexp.MustCompile(`"([^"]+)"`)
			_ = re
			matches = re.FindAllStringSubmatch(prev_line, -1)

			// Check if there are at least two matches and print the second one
			if len(matches) < 2 {
				break
			}
			data_type := matches[1][1] // matches[1][1] contains the second value inside quotes

			if !isValidDataUnit(data_type, data_unit) {
				_ = row
				_ = idx
			       diagnostics = append(diagnostics, lsp.Diagnostic{
					Range:    LineRange(row, idx+1, idx+len("data_type")),
					Severity: 1,
					Source:   "Exo",
					Message:  data_unit + " not a valid data_unit for " + data_type,
				})
			}
		}

		if strings.Contains(line, "Neovim") {
			idx := strings.Index(line, "Neovim")
			diagnostics = append(diagnostics, lsp.Diagnostic{
				Range:    LineRange(row, idx, idx+len("Neovim")),
				Severity: 2,
				Source:   "Common Sense",
				Message:  "Great choice :)",
			})

		}
		prev_line = line
	}

	return diagnostics
}

func (s *State) OpenDocument(uri, text string) []lsp.Diagnostic {
	s.Documents[uri] = text

	return getDiagnosticsForFile(text)
}

func (s *State) UpdateDocument(uri, text string) []lsp.Diagnostic {
	s.Documents[uri] = text

	return getDiagnosticsForFile(text)
}

func (s *State) Hover(id int, uri string, position lsp.Position) lsp.HoverResponse {
	// In real life, this would look up the type in our type analysis code...

	document := s.Documents[uri]

	return lsp.HoverResponse{
		Response: lsp.Response{
			RPC: "2.0",
			ID:  &id,
		},
		Result: lsp.HoverResult{
			Contents: fmt.Sprintf("File: %s, Characters: %d", uri, len(document)),
		},
	}
}

func (s *State) Definition(id int, uri string, position lsp.Position) lsp.DefinitionResponse {
	// In real life, this would look up the definition

	return lsp.DefinitionResponse{
		Response: lsp.Response{
			RPC: "2.0",
			ID:  &id,
		},
		Result: lsp.Location{
			URI: uri,
			Range: lsp.Range{
				Start: lsp.Position{
					Line:      position.Line - 1,
					Character: 0,
				},
				End: lsp.Position{
					Line:      position.Line - 1,
					Character: 0,
				},
			},
		},
	}
}
func (s *State) TextDocumentCodeAction(id int, uri string) lsp.TextDocumentCodeActionResponse {
	text := s.Documents[uri]

	actions := []lsp.CodeAction{}
	for row, line := range strings.Split(text, "\n") {
		idx := strings.Index(line, "VS Code")
		if idx >= 0 {
			replaceChange := map[string][]lsp.TextEdit{}
			replaceChange[uri] = []lsp.TextEdit{
				{
					Range:   LineRange(row, idx, idx+len("VS Code")),
					NewText: "Neovim",
				},
			}

			actions = append(actions, lsp.CodeAction{
				Title: "Replace VS C*de with a superior editor",
				Edit:  &lsp.WorkspaceEdit{Changes: replaceChange},
			})

			censorChange := map[string][]lsp.TextEdit{}
			censorChange[uri] = []lsp.TextEdit{
				{
					Range:   LineRange(row, idx, idx+len("VS Code")),
					NewText: "VS C*de",
				},
			}

			actions = append(actions, lsp.CodeAction{
				Title: "Censor to VS C*de",
				Edit:  &lsp.WorkspaceEdit{Changes: censorChange},
			})
		}
	}

	response := lsp.TextDocumentCodeActionResponse{
		Response: lsp.Response{
			RPC: "2.0",
			ID:  &id,
		},
		Result: actions,
	}

	return response
}

func (s *State) TextDocumentCompletion(id int, uri string) lsp.CompletionResponse {

	// Ask your static analysis tools to figure out good completions
	items := []lsp.CompletionItem{
		{
			Label:         "Neovim (BTW)",
			Detail:        "Very cool editor",
			Documentation: "Fun to watch in videos. Don't forget to like & subscribe to streamers using it :)",
		},
	}

	response := lsp.CompletionResponse{
		Response: lsp.Response{
			RPC: "2.0",
			ID:  &id,
		},
		Result: items,
	}

	return response
}

func LineRange(line, start, end int) lsp.Range {
	return lsp.Range{
		Start: lsp.Position{
			Line:      line,
			Character: start,
		},
		End: lsp.Position{
			Line:      line,
			Character: end,
		},
	}
}
