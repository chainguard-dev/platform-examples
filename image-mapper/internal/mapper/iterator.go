package mapper

import (
	"bufio"
	"errors"
	"io"
)

// ErrIteratorDone indicates when an iterator is finished
var ErrIteratorDone = errors.New("done")

// Iterator iterates over images
type Iterator interface {
	Next() (string, error)
}

type readerIterator struct {
	scanner *bufio.Scanner
}

// NewReaderIterator iterates over images in the given reader. It expects an
// image per line.
func NewReaderIterator(r io.Reader) Iterator {
	return &readerIterator{
		scanner: bufio.NewScanner(r),
	}
}

// Next returns the next line
func (it *readerIterator) Next() (string, error) {
	if !it.scanner.Scan() {
		if it.scanner.Err() != nil {
			return "", it.scanner.Err()
		}

		return "", ErrIteratorDone
	}

	txt := it.scanner.Text()

	// If the line is empty, skip to the next one
	if txt == "" {
		return it.Next()
	}

	return txt, nil
}

type argsIterator struct {
	args  []string
	index int
}

// NewArgsIterator iterates over a list of images
func NewArgsIterator(args []string) Iterator {
	return &argsIterator{
		args: args,
	}
}

// Next returns the next image
func (it *argsIterator) Next() (string, error) {
	if it.index >= len(it.args) {
		return "", ErrIteratorDone
	}

	arg := it.args[it.index]
	it.index++

	return arg, nil
}
