package domain

import (
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"
)

const MaxOfferTags = 5

var tagNamePattern = regexp.MustCompile(`^[A-Za-zА-Яа-яЁё]+$`)

func NormalizeTag(name string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(name))
	if normalized == "" {
		return "", ErrInvalidTagName
	}

	length := utf8.RuneCountInString(normalized)
	if length < 1 || length > 15 {
		return "", ErrInvalidTagName
	}
	if !tagNamePattern.MatchString(normalized) {
		return "", ErrInvalidTagName
	}

	return normalized, nil
}

func NormalizeTags(raw []string) ([]string, error) {
	if len(raw) == 0 {
		return []string{}, nil
	}

	seen := make(map[string]struct{}, len(raw))
	result := make([]string, 0, len(raw))
	for _, tag := range raw {
		normalized, err := NormalizeTag(tag)
		if err != nil {
			return nil, err
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, normalized)
	}

	if len(result) > MaxOfferTags {
		return nil, ErrTooManyTags
	}

	sort.Strings(result)
	return result, nil
}
