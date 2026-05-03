package konghelp

import (
	"bytes"
	"strings"
	"testing"

	"github.com/alecthomas/kong"
)

type helpExit struct{}

func renderHelp(t *testing.T, cli any, args ...string) (output string) {
	t.Helper()

	var stdout bytes.Buffer
	app, err := kong.New(
		cli,
		Help(Options{UseWidth: 80}),
		kong.Name("test"),
		kong.Description("Root help."),
		kong.Writers(&stdout, &stdout),
		kong.Exit(func(int) {
			panic(helpExit{})
		}),
	)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		if value := recover(); value != nil {
			if _, ok := value.(helpExit); ok {
				output = Visible(stdout.String())
				return
			}
			panic(value)
		}
		t.Fatal("expected help to exit")
	}()
	if _, err := app.Parse(args); err != nil {
		t.Fatal(err)
	}
	return Visible(stdout.String())
}

func TestHelpSeparatesSummaryAndOptionsWithoutArguments(t *testing.T) {
	var cli struct {
		Verbose bool `short:"v" help:"Verbose output."`
	}

	output := renderHelp(t, &cli, "--help")

	assertSingleBlankLineBefore(t, output, "Options")
}

func TestHelpDoesNotDuplicateBlankLineBeforeOptionsWithArguments(t *testing.T) {
	var cli struct {
		Paths []string `arg:"" help:"The path to the file to add."`
	}

	output := renderHelp(t, &cli, "--help")

	assertSingleBlankLineBefore(t, output, "Options")
}

func TestHelpDoesNotEmitTrailingWhitespace(t *testing.T) {
	var cli struct {
		Verbose bool   `short:"v" help:"Verbose output."`
		Path    string `type:"path" help:"Path to inspect."`

		Add struct {
			Paths []string `arg:"" help:"The path to the file to add."`
		} `cmd:"" help:"Add file contents to the index."`
	}

	output := renderHelp(t, &cli, "--help")

	if strings.HasSuffix(output, "\n\n") {
		t.Fatalf("expected output not to end with a blank line:\n%s", output)
	}
	for _, line := range strings.Split(strings.TrimSuffix(output, "\n"), "\n") {
		if strings.TrimRight(line, " \t") != line {
			t.Fatalf("expected no trailing whitespace in line %q:\n%s", line, output)
		}
	}
}

func TestAggregateIntoLinesUsesTwoSpaceColumnSeparator(t *testing.T) {
	lines, err := AggregateIntoLines([]string{"*", "path", "STRING", "Path help."}, 80)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := lines[0], "*  path  STRING  Path help."; got != want {
		t.Fatalf("expected two spaces between columns:\nwant: %q\n got: %q", want, got)
	}
}

func assertSingleBlankLineBefore(t *testing.T, output, header string) {
	t.Helper()

	lines := strings.Split(output, "\n")
	for i, line := range lines {
		if strings.TrimSpace(line) != header {
			continue
		}
		if i < 2 {
			t.Fatalf("expected content and one blank line before %q:\n%s", header, output)
		}
		if strings.TrimSpace(lines[i-1]) != "" {
			t.Fatalf("expected blank line before %q:\n%s", header, output)
		}
		if strings.TrimSpace(lines[i-2]) == "" {
			t.Fatalf("expected only one blank line before %q:\n%s", header, output)
		}
		return
	}
	t.Fatalf("could not find %q in output:\n%s", header, output)
}
