package handlers

func truncateString(s string, maxRunes int) string {
	if maxRunes <= 0 || s == "" {
		return ""
	}
	count := 0
	for i := range s {
		if count == maxRunes {
			return s[:i]
		}
		count++
	}
	return s
}
