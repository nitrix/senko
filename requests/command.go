package requests

import "strings"

type EventCommand struct {
	Author string
	Content string
}

func (e EventCommand) Match(pattern string) (map[string]string, bool) {
	matches := make(map[string]string)
	contentParts := strings.Split(e.Content, " ")
	patternParts := strings.Split(pattern, " ")

	for _, patternPart := range patternParts {
		// When we're still processing patternParts and there are no contentParts left,
		// then it's impossible to get a match.
		if len(contentParts) == 0 {
			return nil, false
		}

		// Placeholder.
		if strings.HasPrefix(patternPart, "<") && strings.HasSuffix(patternPart, ">") {
			name := strings.TrimPrefix(patternPart, "<")
			name = strings.TrimSuffix(name, ">")
			matches[name] = contentParts[0]
			contentParts = contentParts[1:]
			continue
		}

		// Choice.
		if strings.HasPrefix(patternPart, "{") && strings.HasSuffix(patternPart, "}") {
			// TODO
		}

		// Exact match.
		if patternPart == contentParts[0] {
			contentParts = contentParts[1:]
			continue
		}

		return nil, false
	}

	return matches, true
}
