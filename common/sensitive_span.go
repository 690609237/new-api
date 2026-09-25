package common

import (
	"slices"
	"strings"
)

// SensitiveRuleMaxRunes is the largest span, including the matched terms,
// allowed for a combined sensitive-word rule.
const SensitiveRuleMaxRunes = 100

// FindSensitiveRuleSpan returns the shortest occurrence covering every term.
// Combined rules must fit within SensitiveRuleMaxRunes; single terms have no
// distance limit. Offsets are rune indices in text, and end is exclusive. A
// repeated term in a rule has the same meaning as a single occurrence.
func FindSensitiveRuleSpan(text string, words []string) (start, end int, found bool) {
	if text == "" || len(words) == 0 {
		return 0, 0, false
	}
	type occurrence struct {
		start int
		end   int
		word  int
	}
	runes := []rune(strings.ToLower(text))
	seen := make(map[string]int, len(words))
	occurrences := make([]occurrence, 0)
	for _, raw := range words {
		word := strings.ToLower(raw)
		if word == "" {
			return 0, 0, false
		}
		if _, exists := seen[word]; exists {
			continue
		}
		index := len(seen)
		seen[word] = index
		pattern := []rune(word)
		for pos := 0; pos+len(pattern) <= len(runes); pos++ {
			if slices.Equal(runes[pos:pos+len(pattern)], pattern) {
				occurrences = append(occurrences, occurrence{pos, pos + len(pattern), index})
			}
		}
	}
	slices.SortFunc(occurrences, func(a, b occurrence) int { return a.start - b.start })
	limit := SensitiveRuleMaxRunes
	if len(seen) == 1 {
		limit = len(runes)
	}
	counts := make([]int, len(seen))
	covered := 0
	left := 0
	bestLength := limit + 1
	for right, hit := range occurrences {
		if counts[hit.word] == 0 {
			covered++
		}
		counts[hit.word]++
		for covered == len(seen) {
			if hit.start-occurrences[left].start < limit {
				maxEnd := 0
				for i := left; i <= right; i++ {
					maxEnd = max(maxEnd, occurrences[i].end)
				}
				if length := maxEnd - occurrences[left].start; length < bestLength {
					start, end, bestLength = occurrences[left].start, maxEnd, length
					found = true
				}
			}
			counts[occurrences[left].word]--
			if counts[occurrences[left].word] == 0 {
				covered--
			}
			left++
		}
	}
	return start, end, found
}
