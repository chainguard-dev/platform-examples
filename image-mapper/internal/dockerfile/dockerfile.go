package dockerfile

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/chainguard-dev/customer-success/scripts/image-mapper/internal/mapper"
	"github.com/moby/buildkit/frontend/dockerfile/parser"
)

// Map images in a Dockerfile to their Chainguard equivalents
func Map(ctx context.Context, input []byte, opts ...mapper.Option) ([]byte, error) {
	m, err := NewMapper(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("constructing mapper: %w", err)
	}

	return mapDockerfile(m, input)
}

func mapDockerfile(m mapper.Mapper, input []byte) ([]byte, error) {
	res, err := parser.Parse(bytes.NewReader(input))
	if err != nil {
		return nil, fmt.Errorf("parse dockerfile: %w", err)
	}

	// Keep track of the name of the stages in the Dockerfile so we don't
	// mistake a stage name for an image name in `COPY --from=<image>` or
	// RUN `--mount=type=bind,from=<image>` style instructions.
	stages := map[string]struct{}{}

	// Keep track of args so we can resolve them in `FROM` instructions.
	args := map[string]string{}

	// Track when we hit the first `FROM` instruction, because any ARGs after that
	// point aren't usable in `FROM` instructions.
	beforeFrom := true

	// We'll compose the output by replacing lines in the input
	output := string(input)

	// Track the number of lines we've removed so we can adjust the line
	// numbers we modify accordingly
	offset := 0

	for _, child := range res.AST.Children {
		var replacement string

		switch strings.ToLower(child.Value) {

		// ARG EXAMPLE=<image>
		case "arg":
			if child.Next == nil {
				continue
			}
			if !beforeFrom {
				continue
			}

			// Save the args, if there's a value
			for n := child.Next; n != nil; n = n.Next {
				parts := strings.Split(n.Value, "=")
				if len(parts) == 2 {
					args[parts[0]] = strings.Trim(parts[1], "\"")
				}
			}

		// FROM <image> [AS <stage>]
		case "from":
			beforeFrom = false
			if child.Next == nil {
				continue
			}

			// Save the stage name, if there is one
			for n := child.Next; n != nil; n = n.Next {
				if strings.ToLower(n.Value) != "as" {
					continue
				}
				if n.Next == nil {
					continue
				}

				stages[n.Next.Value] = struct{}{}
			}

			// Resolve args in the FROM line
			from := resolveArgs(args, child.Next.Value)

			// Map the image to Chainguard
			img, err := mapper.MapImage(m, from)
			if err != nil {
				log.Printf("WARN: error mapping image: %s: %s", from, err)
				continue
			}

			replacement = strings.ReplaceAll(child.Original, child.Next.Value, img.String())

		// COPY --from=<image>
		case "copy":
			for _, flag := range child.Flags {
				if !strings.HasPrefix(flag, "--from=") {
					continue
				}

				from := strings.TrimPrefix(flag, "--from=")

				// Skip if --from refers to a stage, rather than an image
				if _, ok := stages[from]; ok {
					continue
				}

				img, err := mapper.MapImage(m, from)
				if err != nil {
					log.Printf("WARN: error mapping image: %s: %s", from, err)
					continue
				}

				replacement = strings.ReplaceAll(child.Original, flag, fmt.Sprintf("--from=%s", img))

				break
			}

		// RUN --mount=type=bind,target=/usr/bin,from=python
		case "run":
			original := child.Original

			for _, flag := range child.Flags {
				if !strings.HasPrefix(flag, "--mount=") {
					continue
				}

				// Extract the image from a from=<image> option
				match := fromPattern.FindStringSubmatch(flag)
				if len(match) < 2 {
					continue
				}
				from := match[1]

				// Skip if from= refers to a stage, rather than
				// an image
				if _, ok := stages[from]; ok {
					continue
				}

				img, err := mapper.MapImage(m, from)
				if err != nil {
					log.Printf("WARN: error mapping image: %s: %s", from, err)
					continue
				}
				// Replace the from= option in the flag
				// with the mapped image
				modifiedFlag := strings.ReplaceAll(flag, fmt.Sprintf("from=%s", from), fmt.Sprintf("from=%s", img))

				// Replace the flag with the modified flag in
				// the original line
				original = strings.ReplaceAll(original, flag, modifiedFlag)

			}

			if original != child.Original {
				replacement = original
			}
		}

		if replacement == "" {
			continue
		}

		// If the original instruction was spread over multiple lines
		// then we will flatten it onto one line and remove the other
		// lines to keep things tidy.
		//
		// One consequence of this is that we need to adjust the line
		// numbers we write to after that point to account for the
		// offset.
		output = replaceLines(output, child.StartLine-offset, child.EndLine-offset, replacement)
		offset = offset + (child.EndLine - child.StartLine)
	}

	return []byte(output), nil
}

// fromPattern extracts images in `from=` options in `RUN --mount` instructions
var fromPattern = regexp.MustCompile(`\bfrom=([^,]+)`)

// argPattern identifies arguments like `${ARG_NAME}`
var argPattern = regexp.MustCompile(`\$\{([^}]+)\}`)

// resolveArgs resolves args in a Dockerfile line
func resolveArgs(args map[string]string, line string) string {
	return argPattern.ReplaceAllStringFunc(line, func(match string) string {
		// Extract the inside of ${...}
		content := match[2 : len(match)-1]

		// Check for default syntax: VAR:-default
		argName := content
		argDefault := ""

		parts := strings.Split(content, ":-")
		if len(parts) > 1 {
			argName = parts[0]
			argDefault = parts[1]
		}

		// If the variable exists in map, use it
		if val, ok := args[argName]; ok {
			return val
		}

		// Otherwise, if a default is provided, use it
		if argDefault != "" {
			return argDefault
		}

		// No variable and no default → return original pattern unchanged
		return match
	})
}

// replaceLines replaces the indicated lines in the output with the replacement
// value
func replaceLines(output string, start, end int, replacement string) string {
	lines := strings.Split(output, "\n")
	if start < 0 || end >= len(lines) || start > end {
		return output
	}

	lines[start-1] = replacement

	lines = append(lines[:start], lines[end:]...)

	return strings.Join(lines, "\n")
}
