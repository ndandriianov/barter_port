package seed_demo

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"unicode"
)

var (
	offerPhotoIndexOnce sync.Once
	offerPhotoIndex     map[string]string
	offerPhotoIndexErr  error
)

func resolveOfferPhotoPath(userKey string, spec offerSpec) (string, error) {
	if spec.SkipPhoto {
		return "", nil
	}

	index, err := loadOfferPhotoIndex()
	if err != nil {
		return "", err
	}

	for _, candidate := range offerPhotoCandidates(spec) {
		if path, ok := index[normalizeOfferPhotoLookup(userKey+"_"+candidate)]; ok {
			return path, nil
		}
	}

	return "", fmt.Errorf("photo for %s/%s not found in resources", userKey, spec.Key)
}

func loadOfferPhotoIndex() (map[string]string, error) {
	offerPhotoIndexOnce.Do(func() {
		dir, err := offerResourcesDir()
		if err != nil {
			offerPhotoIndexErr = err
			return
		}

		entries, err := os.ReadDir(dir)
		if err != nil {
			offerPhotoIndexErr = fmt.Errorf("read resources dir: %w", err)
			return
		}

		index := make(map[string]string, len(entries))
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			name := entry.Name()
			index[normalizeOfferPhotoLookup(name)] = filepath.Join(dir, name)
		}

		offerPhotoIndex = index
	})

	return offerPhotoIndex, offerPhotoIndexErr
}

func offerResourcesDir() (string, error) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("resolve seed-demo source path")
	}

	dir := filepath.Clean(filepath.Join(filepath.Dir(filename), "..", "resources"))
	if _, err := os.Stat(dir); err != nil {
		return "", fmt.Errorf("stat resources dir %s: %w", dir, err)
	}

	return dir, nil
}

func offerPhotoCandidates(spec offerSpec) []string {
	candidates := []string{spec.Name}

	if trimmed := trimWantedPrefix(spec.Name); trimmed != "" && trimmed != spec.Name {
		candidates = append(candidates, trimmed)
	}

	candidates = append(candidates, spec.Key)
	candidates = append(candidates, spec.PhotoAliases...)

	seen := make(map[string]struct{}, len(candidates))
	result := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		candidate = strings.TrimSpace(candidate)
		if candidate == "" {
			continue
		}
		key := normalizeOfferPhotoLookup(candidate)
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, candidate)
	}

	return result
}

func trimWantedPrefix(value string) string {
	trimmed := strings.TrimSpace(value)
	lower := strings.ToLower(trimmed)
	if strings.HasPrefix(lower, "ищу ") {
		return strings.TrimSpace(lower[len("ищу "):])
	}

	return trimmed
}

func normalizeOfferPhotoLookup(value string) string {
	base := strings.TrimSuffix(filepath.Base(strings.TrimSpace(value)), filepath.Ext(value))

	var b strings.Builder
	prevSeparator := false
	for _, r := range strings.ToLower(base) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			prevSeparator = false
			continue
		}

		if prevSeparator {
			continue
		}

		b.WriteByte('_')
		prevSeparator = true
	}

	return strings.Trim(b.String(), "_")
}
