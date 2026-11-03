package main

import "unicode/utf8"

// SplitSlackMessage splits a Slack message into multiple messages if it exceeds the maximum number of runes.
func SplitSlackMessage(text string, maxRunes int) []string {
	limit := maxRunes - 250 // Leave room for msg formatting, etc.
	count := utf8.RuneCountInString(text)

	if count < limit {
		return []string{text}
	}

	// TODO: Split the message into multiple parts.
	// 1. Make sure to break after a paragraph or on last line break.
	// 2. If the break is found to be make inside a block of code, then close the block and reopen it in the new part
	return nil // TODO
}
