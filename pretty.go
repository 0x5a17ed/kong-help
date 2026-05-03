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
		if selected := ctx.Selected(); selected == nil {
			if !w.Options.NoAppSummary {
				w.Print("")
				w.Indent().Printf("%s: %s%s", ColorExample("Usage"), app.Name, app.Summary())
			}
			printNode(w, app.Node, true)
			w.Indent().Printf(`Use "%s --help" for more info`, app.Name)
		} else {
			if !w.Options.NoAppSummary {
				w.Print("")
				w.Indent().Printf("%s: %s", ColorExample("Usage"), selected.Summary())
			}
			printNode(w, selected, true)
			w.Indent().Printf(`Use "%s %s --help" for more info`, app.Name, selected.Name)
		}

		if _, err := w.WriteTo(ctx.Stdout); err != nil {
			return err
		}
		return nil
	}
}

func printNode(w *helpWriter, node *kong.Node, hide bool) {
	if node.Help != "" {
		w.Print("")
		w.Indent().PrintWrap(node.Help)
	}
	if w.Options.Summary {
		return
	}
	if node.Detail != "" {
		w.Print("")
		w.Indent().PrintWrap(node.Detail)
	}
	if len(node.Positional) > 0 {
		w.Print("")
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
		lines = append(lines, formatGroup(collection.Metadata)...)
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

	// groupedCmds := collectCommandGroups(cmds)
	// for _, group := range groupedCmds {
	// 	w.Print("")
	// 	if group.Metadata.Title != "" {
	// 		w.Wrap(group.Metadata.Title)
	// 	}
	// 	if group.Metadata.Description != "" {
	// 		w.Indent().Wrap(group.Metadata.Description)
	// 		w.Print("")
	// 	}
	//
	// 	if w.Compact {
	// 		writeCompactCommandList(group.Commands, iw)
	// 	} else {
	// 		writeCommandList(group.Commands, iw)
	// 	}
	// }
}

func printCard(w *helpWriter, header string, lines [][]string) {
	printCardHeader(w, header)
	w.CardSection().PrintColumns(lines)
	printCardFooter(w)
}

func printCardHeader(w *helpWriter, title string) {
	padding := strings.Repeat("─", w.width-len(title)-7)
	w.Printf("╭─ %s ─%s─╮", ColorLow(title), padding)
}

func printCardFooter(w *helpWriter) {
	padding := strings.Repeat("─", w.width-2)
	w.Print(fmt.Sprintf("╰%s╯", padding))
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
			Metadata: &kong.Group{Title: "Flags:"},
			Flags:    ungroupedFlags,
		})
	}
	for _, group := range groups {
		out = append(out, helpFlagGroup{Metadata: group, Flags: flagsByGroup[group.Key]})
	}
	return out
}
