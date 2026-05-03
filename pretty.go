package konghelp

import (
	"fmt"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/fatih/color"
)

// TODO: Make these configurable
var ColorExample = color.New(color.FgYellow).SprintFunc()
var ColorRequired = color.New(color.FgRed).SprintFunc()
var ColorDefault = color.New(color.FgMagenta).SprintFunc()
var ColorPlaceHolder = ColorDefault
var ColorCommand = color.New(color.FgCyan).SprintFunc()
var ColorCardHeader = color.New(color.FgGreen).SprintFunc()
var ColorLow = color.HiBlackString
var ColorType = ColorExample
var ColorGroup = color.New(color.FgBlue).Add(color.Underline).SprintFunc()

func NewPrettyPrinter(printOpts Options) kong.HelpPrinter {
	return func(options kong.HelpOptions, ctx *kong.Context) error {
		if ctx.Empty() {
			options.Summary = false
		}
		// TODO: Have this controlled via an option
		options.ValueFormatter = PrettyValueFormatter(options.ValueFormatter)
		w := newHelpWriter(ctx, options, printOpts.width)

		app := ctx.Model

		selected := ctx.Selected()
		if selected == nil {
			selected = app.Node
		}

		printNode(w, app, selected, true)

		if _, err := w.WriteTo(ctx.Stdout); err != nil {
			return err
		}
		return nil
	}
}

func printNode(w *helpWriter, app *kong.Application, node *kong.Node, hide bool) {
	if node.Help != "" {
		w.PrintWrap(node.Help)
	}

	if !w.Options.NoAppSummary {
		printUsage(w, app, node)
	}
	if w.Options.Summary {
		return
	}
	if node.Detail != "" {
		w.PrintBlankLine()
		w.Indent().PrintWrap(node.Detail)
	}
	if len(node.Positional) > 0 {
		printPositionals(w, node.Positional)
	}
	if !w.Options.FlagsLast {
		printFlags(w, node.AllFlags(true))
	}

	if w.Options.NoExpandSubcommands {
		printCommands(w, node.Children)
	} else {
		printCommands(w, node.Leaves(hide))
	}

	if w.Options.FlagsLast {
		printFlags(w, node.AllFlags(true))
	}
}

func printUsage(w *helpWriter, app *kong.Application, node *kong.Node) {
	printCard(w, "Usage", [][]string{
		{"  ", fmt.Sprintf("%s %s", app.Name, strings.TrimSpace(node.Summary()))},
	})
}

func printPositionals(w *helpWriter, args []*kong.Positional) {
	lines := [][]string{}
	for _, arg := range args {
		line := formatPositional(arg, w.Options.ValueFormatter)
		lines = append(lines, line)
	}
	printCard(w, "Arguments", lines)
}

func printFlags(w *helpWriter, flags [][]*kong.Flag) {
	lines := [][]string{}
	for _, collection := range collectFlagGroups(flags) {
		if collection.Metadata != nil {
			lines = append(lines, formatGroup(collection.Metadata)...)
		}
		for _, flagset := range collection.Flags {
			for _, flag := range flagset {
				lines = append(lines, formatFlag(flag, w.Options.ValueFormatter))
			}
		}
	}
	printCard(w, "Options", lines)
}

func printCommands(w *helpWriter, cmds []*kong.Command) {
	if len(cmds) == 0 {
		return
	} else if w.Options.Tree {
		panic("Options.Tree not supported")
	}

	// TODO: Handle groups
	lines := [][]string{}
	for _, cmd := range cmds {
		if cmd.Hidden {
			continue
		}
		lines = append(lines, formatCommand(cmd, w.Options.Compact)...)
	}
	printCard(w, "Commands", lines)
}

func printCard(w *helpWriter, header string, lines [][]string) {
	w.PrintBlankLine()
	printCardHeader(w, header)
	w.CardSection().PrintColumns(lines)
	printCardFooter(w)
}

func printCardHeader(w *helpWriter, title string) {
	w.Print(ColorCardHeader(title))
}

func printCardFooter(_ *helpWriter) {
}

// Directly from kong source code:

type helpFlagGroup struct {
	Metadata *kong.Group
	Flags    [][]*kong.Flag
}

func collectFlagGroups(flags [][]*kong.Flag) []helpFlagGroup {
	// Group keys in order of appearance.
	groups := []*kong.Group{}
	// Flags grouped by their group key.
	flagsByGroup := map[string][][]*kong.Flag{}

	for _, levelFlags := range flags {
		levelFlagsByGroup := map[string][]*kong.Flag{}

		for _, flag := range levelFlags {
			key := ""
			if flag.Group != nil {
				key = flag.Group.Key
				groupAlreadySeen := false
				for _, group := range groups {
					if key == group.Key {
						groupAlreadySeen = true
						break
					}
				}
				if !groupAlreadySeen {
					groups = append(groups, flag.Group)
				}
			}

			levelFlagsByGroup[key] = append(levelFlagsByGroup[key], flag)
		}

		for key, flags := range levelFlagsByGroup {
			flagsByGroup[key] = append(flagsByGroup[key], flags)
		}
	}

	out := []helpFlagGroup{}
	// Ungrouped flags are always displayed first.
	if ungroupedFlags, ok := flagsByGroup[""]; ok {
		out = append(out, helpFlagGroup{
			Flags: ungroupedFlags,
		})
	}
	for _, group := range groups {
		out = append(out, helpFlagGroup{Metadata: group, Flags: flagsByGroup[group.Key]})
	}
	return out
}
