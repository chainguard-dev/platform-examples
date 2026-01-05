package mapper

// TagFilter is a function that filters tags
type TagFilter func(tag string) bool

func includeTags(tags []string, filters ...TagFilter) []string {
	if len(filters) == 0 {
		return tags
	}
	var output []string
	for _, tag := range tags {
		if !includeTag(tag, filters...) {
			continue
		}

		output = append(output, tag)
	}

	return output
}

func includeTag(tag string, filters ...TagFilter) bool {
	for _, filter := range filters {
		if !filter(tag) {
			continue
		}

		return true
	}

	return false
}
