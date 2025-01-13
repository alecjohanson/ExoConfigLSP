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
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	logger := getLogger(dir + "/log.txt")
	logger.Println("Checking File Diagnostics")
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
			if idx < 0 {
				logger.Printf("No Data_type:%d:", idx)
				diagnostics = append(diagnostics, lsp.Diagnostic{
					Range:    LineRange(row, idx, idx+len("data_unit")),
					Severity: 1,
					Source:   "Exo",
					Message:  "No Associated data_type found for " + data_unit,
				})
				break
			}
			_ = regexp.MustCompile(`"([^"]+)"`)
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

func data_types_completion() []lsp.CompletionItem {
	var items []lsp.CompletionItem
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	file_dir := dir + "/analysis/ExoTypesUnits.json"
	fileData, _ := os.ReadFile(file_dir)

	var data_types map[string]any
	_ = json.Unmarshal(fileData, &data_types)

	for k,_ := range data_types {
		items = append(items, lsp.CompletionItem{ Label: k })
	}
	return items
}

func data_units_completion(data_type string) []lsp.CompletionItem {
	var items []lsp.CompletionItem
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	file_dir := dir + "/analysis/ExoTypesUnits.json"
	fileData, _ := os.ReadFile(file_dir)

	var data_types map[string][]string
	_ = json.Unmarshal(fileData, &data_types)
	value, _ := data_types[data_type]
	for v := range value {
		items = append(items, lsp.CompletionItem{ Label: value[v]})
	}
	return items
}

func (s *State) TextDocumentCompletion(id int, uri string, line_num int) lsp.CompletionResponse {
	var text string = s.Documents[uri]
	var items []lsp.CompletionItem
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	logger := getLogger(dir + "/log.txt")
	logger.Printf("Completion Request %d on %s", id, uri)

	line := strings.Split(text, "\n")[line_num] 
	if (strings.Contains(line, "data_type")) {
		items = data_types_completion()
	} else if (strings.Contains(line, "data_unit")) {
		type_line := strings.Split(text, "\n")[line_num-1] 
		var re = regexp.MustCompile(`"([^"]+)"`)
		var matches = re.FindAllStringSubmatch(type_line, -1)

		// Check if there are at least two matches and print the second one
		if len(matches) >= 2 { 
			data_type := matches[1][1] // matches[1][1] contains the second value inside quotes
			items = data_units_completion(data_type)
		}
	} else {
		items = append(items, lsp.CompletionItem{ Label: "\"<channel_id>\": { \n \"display_name\": \"<display_name>\", \n \"properties\": { \n \"data_type\": \"<dt>\", \n \"data_unit\": \"<du>\" \n } \n },", Detail: "Full Channel"})
		items = append(items, lsp.CompletionItem{ Label: "\"data_type\": \"\","})
		items = append(items, lsp.CompletionItem{ Label: "\"data_unit\": \"\","})

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
