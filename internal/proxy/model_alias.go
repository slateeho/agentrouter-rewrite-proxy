package proxy

import (
	"bytes"
	"encoding/json"
	"strings"
)

// RewriteModel replaces an exact model name in a JSON body.
// Only replaces when model exactly matches oldModel.
// Returns original bytes if JSON is malformed or no match.
func RewriteModel(raw []byte, oldModel, newModel string) []byte {
	if len(raw) == 0 {
		return raw
	}
	var body map[string]any
	if err := json.Unmarshal(raw, &body); err != nil {
		return raw
	}
	if m, ok := body["model"].(string); ok && m == oldModel {
		body["model"] = newModel
		out, err := json.Marshal(body)
		if err != nil {
			return raw
		}
		return out
	}
	return raw
}

// ResponseRewriteModel replaces model in a response JSON body.
// Used to hide upstream model name from client.
// Handles both exact match and AgentRouter's canonicalized Opus-5 model IDs.
func ResponseRewriteModel(raw []byte, oldModel, newModel string) []byte {
	if len(raw) == 0 {
		return raw
	}
	var body map[string]any
	if err := json.Unmarshal(raw, &body); err != nil {
		return raw
	}

	// Helper to check if model should be rewritten
	shouldRewrite := func(m string) bool {
		if m == oldModel {
			return true
		}
		// AgentRouter returns canonicalized model IDs like "anthropic/claude-opus-5-ps-aws-dst"
		// Rewrite if it starts with "anthropic/" + oldModel + "-"
		if strings.HasPrefix(m, "anthropic/"+oldModel+"-") {
			return true
		}
		return false
	}

	// For non-streaming response, replace at top level
	if m, ok := body["model"].(string); ok && shouldRewrite(m) {
		body["model"] = newModel
		out, err := json.Marshal(body)
		if err != nil {
			return raw
		}
		return out
	}

	// For Anthropic SSE streams, check message.model
	if m, ok := body["message"].(map[string]any); ok {
		if mm, ok := m["model"].(string); ok && shouldRewrite(mm) {
			m["model"] = newModel
			out, err := json.Marshal(body)
			if err != nil {
				return raw
			}
			return out
		}
	}

	return raw
}

// ResponseRewriteModelInSSE rewrites model in SSE event data.
// Handles both standard SSE "data: {...}" format and Anthropic SSE format
// with "event:" and "data:" lines.
func ResponseRewriteModelInSSE(raw []byte, oldModel, newModel string) []byte {
	if len(raw) == 0 {
		return raw
	}

	// Split into lines to find "data: " lines
	lines := bytes.Split(raw, []byte("\n"))

	// Find all data: lines and their positions
	var dataLines []struct {
		lineNum int
		start   int
		end     int
	}

	pos := 0
	for i, line := range lines {
		if bytes.HasPrefix(line, []byte("data: ")) {
			lineStart := pos
			lineEnd := pos + len(line)
			dataLines = append(dataLines, struct {
				lineNum int
				start   int
				end     int
			}{lineNum: i, start: lineStart, end: lineEnd})
		}
		pos += len(line) + 1 // +1 for the newline
	}

	// If no data lines found, return unchanged
	if len(dataLines) == 0 {
		return raw
	}

	// Process each data: line
	result := make([]byte, 0, len(raw))
	lastEnd := 0

	for _, dl := range dataLines {
		// Add everything before this data line
		result = append(result, raw[lastEnd:dl.start]...)

		// Extract the JSON data after "data: "
		jsonDataStart := dl.start + len("data: ")
		jsonDataEnd := dl.end
		if jsonDataEnd > jsonDataStart && raw[jsonDataEnd-1] == '\n' {
			jsonDataEnd--
		}
		jsonData := raw[jsonDataStart:jsonDataEnd]

		// Rewrite the model in the JSON
		rewritten := ResponseRewriteModel(jsonData, oldModel, newModel)

		// Add the "data: " prefix + rewritten JSON
		result = append(result, "data: "...)
		result = append(result, rewritten...)

		lastEnd = dl.end
	}

	// Add remaining content after the last data line
	result = append(result, raw[lastEnd:]...)

	return result
}
