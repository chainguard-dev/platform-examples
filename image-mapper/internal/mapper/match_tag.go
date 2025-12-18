package mapper

import (
	"regexp"
	"strconv"
)

// MatchTag returns the best matching active tag for the input tag. It'll return
// an empty string if it can't find an appropriate match.
func MatchTag(activeTags []string, tag string) string {
	for _, fn := range matchTagFns {
		match := fn(activeTags, tag)
		if match == "" {
			continue
		}

		return match
	}

	return ""
}

// MatchTagFn matches a tag to one of the provided active tags
type MatchTagFn func(activeTag []string, tag string) string

var matchTagFns = []MatchTagFn{
	matchEqualTag,
	matchClosestSemanticVersionTag,
}

// matchEqualTag identifies an exact match between the input tag and one of the
// activeTags
func matchEqualTag(activeTags []string, tag string) string {
	for _, activeTag := range activeTags {
		if activeTag != tag {
			continue
		}
		return tag
	}

	return ""
}

// matchClosestSemanticVersionTag finds the closest match to the input tag in
// the active tags.
//
// For instance:
//
//	2 -> 3
//	3.7 -> 3.9
//	3.11.1 -> 3.11.5
func matchClosestSemanticVersionTag(activeTags []string, tag string) string {
	parsedTag := parseTag(tag)
	if parsedTag == nil {
		return ""
	}

	var (
		bestMatch    *tagVersion
		bestMatchStr string
	)

	for _, activeTag := range activeTags {
		parsedActive := parseTag(activeTag)
		if parsedActive == nil {
			continue
		}

		// Must have same specificity (i.e major, minor, patch)
		if parsedActive.specificity != parsedTag.specificity {
			continue
		}

		// Must both have v prefix, or no v prefix
		if parsedActive.hasV != parsedTag.hasV {
			continue
		}

		// Active tag must be >= input tag
		if parsedActive.LessThan(parsedTag) {
			continue
		}

		// Active tag must be < the current best match we've found
		if bestMatch != nil && !parsedActive.LessThan(bestMatch) {
			continue
		}

		bestMatch = parsedActive
		bestMatchStr = activeTag

	}

	if bestMatch == nil {
		return ""
	}

	return bestMatchStr
}

type tagVersion struct {
	hasV        bool
	major       int
	minor       int
	patch       int
	specificity string
}

var tagRegex = regexp.MustCompile(`^(v?)(\d+)(?:\.(\d+))?(?:\.(\d+))?(?:-.*)?$`)

func parseTag(tag string) *tagVersion {
	matches := tagRegex.FindStringSubmatch(tag)
	if matches == nil {
		return nil
	}

	tv := &tagVersion{}

	// Check for v prefix
	tv.hasV = matches[1] == "v"

	// Parse major version (required)
	major, err := strconv.Atoi(matches[2])
	if err != nil {
		return nil
	}
	tv.major = major
	tv.specificity = "MAJOR"

	// Parse minor version if present
	if matches[3] != "" {
		minor, err := strconv.Atoi(matches[3])
		if err != nil {
			return nil
		}
		tv.minor = minor
		tv.specificity = "MINOR"
	}

	// Parse patch version if present
	if matches[4] != "" {
		patch, err := strconv.Atoi(matches[4])
		if err != nil {
			return nil
		}
		tv.patch = patch
		tv.specificity = "PATCH"
	}

	return tv
}

// LessThan tests whether this tag is less than the provided one
func (tv *tagVersion) LessThan(other *tagVersion) bool {
	return other != nil && tv.compare(other) < 0
}

// compare returns -1 if tv < other, 0 if equal, 1 if tv > other
func (tv *tagVersion) compare(other *tagVersion) int {
	if tv.major != other.major {
		if tv.major < other.major {
			return -1
		}
		return 1
	}

	if tv.minor != other.minor {
		if tv.minor < other.minor {
			return -1
		}
		return 1
	}

	if tv.patch != other.patch {
		if tv.patch < other.patch {
			return -1
		}
		return 1
	}

	return 0
}
