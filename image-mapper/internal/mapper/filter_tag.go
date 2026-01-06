package mapper

import "strings"

// TagFilter is a function that filters tags
type TagFilter func(tags []string) []string

// TagFilterExcludeDev excludes -dev tags from a list of tags
func TagFilterExcludeDev(tags []string) []string {
	var out []string
	for _, tag := range tags {
		if strings.HasSuffix(tag, "-dev") {
			continue
		}

		out = append(out, tag)
	}

	return out
}

// TagFilterIncludeDev only includes -dev tags from a list of tags
func TagFilterIncludeDev(tags []string) []string {
	var out []string
	for _, tag := range tags {
		if !strings.HasSuffix(tag, "-dev") {
			continue
		}

		out = append(out, tag)
	}

	return out
}

// TagFilterPreferDev returns only -dev tags if any exist, otherwise returns all tags
func TagFilterPreferDev(tags []string) []string {
	var hasDev bool
	for _, tag := range tags {
		if strings.HasSuffix(tag, "-dev") {
			hasDev = true
			break
		}
	}
	if !hasDev {
		return tags
	}

	return TagFilterIncludeDev(tags)
}

func filterTags(repo Repo, filters ...TagFilter) []string {
	tags := repo.ActiveTags
	if len(repo.Tags) > 0 {
		tags = flattenTags(repo.Tags)
	}
	if len(filters) == 0 {
		return tags
	}

	for _, filter := range filters {
		tags = filter(tags)
	}

	return tags
}
