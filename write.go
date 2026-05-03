package konghelp

import (
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/alecthomas/kong"
)

// Ensure we implement the correct interfaces
var _ io.WriterTo = &helpWriter{}

const _MaxIterations = 100
const columnSeparator = "  "

type helpWriter struct {
	Options kong.HelpOptions
	lines   *[]string
	prefix  string
	width   int
}

func newHelpWriter(ctx *kong.Context, options kong.HelpOptions, width int) *helpWriter {
	lines := []string{}
	return &helpWriter{
		Options: options,
		lines:   &lines,
		// TODO: Get this from somewhere else
		// TODO: Use $COLUMNS? golang.org/x/term .IsTerminal(os.Stdout.Fd())?
		width: width,
	}
}

// NOTE: The first part of each row must have an identical length.
//
//	Only the last part is truncated by word.
func (w *helpWriter) PrintColumns(lines [][]string) {
	if len(lines) == 0 {
		return
	}

	// Calculate the max width of each part (besides the last one)
	maxes := []int{-1}
	for _, parts := range lines {
		for i, part := range parts {
			if i == len(parts)-1 || i == 0 {
				continue
			}
			part = Visible(part)
			if len(maxes) == i {
				maxes = append(maxes, len(part))
			}
			if maxes[i] < len(part) {
				maxes[i] = len(part)
			}
		}
	}

	// Pad out the columns
	for j, parts := range lines {
		for i, part := range parts {
			if i == len(parts)-1 || i == 0 {
				continue
			}
			// NOTE: Naive. Part is the same as the last loop
			visible := Visible(part)
			if width := maxes[i] - len(visible); width >= 1 {
				padded := part + strings.Repeat(" ", width)
				lines[j][i] = padded
			}
		}
	}

	for _, parts := range lines {
		lines, err := AggregateIntoLines(parts, w.width)
		if err != nil {
			// TODO: Return errors
			panic(err)
		}
		for _, line := range lines {
			w.Print(strings.TrimRight(line, " "))
		}
	}
}

func AggregateIntoLines(parts []string, maxWidth int) ([]string, error) {
	// Base case: whole line < maxWidth
	line := strings.Join(parts, columnSeparator)
	visible := Visible(line)
	if len(visible) <= maxWidth {
		return []string{line}, nil
	}

	tail := len(parts) - 1

	// Find most number of columns that fit without wrapping
	var paddingSize int
	var lines []string
	var count int
	for {
		count++
		if count > _MaxIterations {
			return nil, errors.New("not implemented: truncating long flags: terminal too small")
		}
		if tail == 0 {
			return nil, fmt.Errorf("terminal too small: %v", parts)
		}
		newLine := strings.Join(parts[:tail], columnSeparator)
		if visible := Visible(newLine); len(visible) < maxWidth {
			paddingSize = len(visible)
			lines = []string{newLine}
			break
		}
		tail--
	}
	// Wrap columns that don't fit
	parts = parts[tail:]
	for _, part := range parts {
		i := 0
		words := strings.Split(part, " ")
		for i < len(words) {
			word := words[i]
			switch {
			case len(word) >= maxWidth:
				if VisibleLen(word) != len(word) {
					return nil, fmt.Errorf(
						"not implemented: truncating long lines with ANSI: %s", word,
					)
				}
				words[i] = TruncateWithSuffix(word, maxWidth, "...")
			case VisibleLen(lines[len(lines)-1])+VisibleLen(word) >= maxWidth:
				lines = append(lines, strings.Repeat(" ", paddingSize))
			default:
				oldLine := lines[len(lines)-1]
				lines[len(lines)-1] = oldLine + " " + word
				i++
			}
		}
	}
	return lines, nil
}

func (w *helpWriter) Print(line string) {
	*w.lines = append(*w.lines, fmt.Sprintf("%s%s", w.prefix, line))
}

func (w *helpWriter) PrintBlankLine() {
	lines := *w.lines
	if len(lines) == 0 {
		return
	}
	if strings.TrimSpace(Visible(lines[len(lines)-1])) == "" {
		return
	}
	w.Print("")
}

func (w *helpWriter) PrintWrap(line string) {
	// TODO: Implement
	w.Print(line)
}

func (w *helpWriter) Printf(format string, args ...any) {
	w.Print(fmt.Sprintf(format, args...))
}

func (w *helpWriter) WriteTo(writer io.Writer) (int64, error) {
	var count int64
	for _, line := range *w.lines {
		n, err := fmt.Fprintln(writer, line)
		if err != nil {
			return -1, err
		}
		count = count + (int64)(n)
	}
	return count, nil
}

func (w *helpWriter) Indent() *helpWriter {
	return &helpWriter{
		Options: w.Options,
		prefix:  w.prefix + "  ",
		lines:   w.lines,
		width:   w.width - 2,
	}
}

func (w *helpWriter) CardSection() *helpWriter {
	return &helpWriter{
		Options: w.Options,
		prefix:  w.prefix,
		lines:   w.lines,
		width:   w.width,
	}
}

// Calculates the length of a string only counting visible characters
func VisibleLen(str string) int {
	// TODO: Can probably make this function optimized with simd to check for control codes
	return len(Visible(str))
}

func Visible(str string) string {
	ansiRegex := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	return ansiRegex.ReplaceAllString(str, "")
}

func TruncateWithSuffix(line string, width int, suffix string) string {
	if len(line) <= width {
		return line
	}
	newLine := line[:width-len(suffix)] + suffix
	if len(newLine) > width {
		panic("coded this wrong")
	}
	return newLine
}
