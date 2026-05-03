package konghelp

import (
	"fmt"
	"log/slog"
	"reflect"
	"strings"

	"github.com/alecthomas/kong"
)

func PrettyValueFormatter(formatter kong.HelpValueFormatter) kong.HelpValueFormatter {
	return func(value *kong.Value) string {
		parts := []string{formatter(value)}

		tag := value.Tag
		if tag == nil {
			return parts[0]
		}
		if tag.Required {
			parts = append(parts, ColorRequired("[required]"))
		}
		if value.Flag == nil && tag.HasDefault {
			parts = append(parts, ColorDefault(fmt.Sprintf("[default=%s]", tag.Default)))
		}
		return strings.Join(parts, " ")
	}
}

func formatValue(tag *kong.Tag, value reflect.Value, showBool bool) string {
	if tag != nil && tag.Type != "" {
		switch tag.Type {
		case "filecontent":
			return normalizeType("PATH")
		case "existingfile":
			return normalizeType("FILE")
		default:
			return normalizeType(tag.Type)
		}
	}
	switch value.Kind() {
	case reflect.Pointer:
		return formatType(value.Type().Elem())
	case reflect.Bool:
		if !showBool {
			return ""
		}
		fallthrough
	default:
		return formatType(value.Type())
	}
}

func formatType(t reflect.Type) string {
	name := t.Name()
	// Slices, maps etc
	if name == "" {
		name = t.String()
	}
	return normalizeType(name)
}

func normalizeType(name string) string {
	return ColorType(strings.ToUpper(name))
}

func formatGroup(group *kong.Group) [][]string {
	title, _ := strings.CutSuffix(group.Title, ":")
	return [][]string{
		{"  ", ColorGroup(title), group.Description},
	}
}

func formatPositional(arg *kong.Positional, format kong.HelpValueFormatter) []string {
	var prefix = "  "
	if arg.Tag != nil && arg.Tag.Required {
		prefix = ColorRequired("* ")
	}
	return []string{
		prefix,
		arg.Name,
		// TODO: Parse format/enum to do along the line of PATH[existing file] or STRING[enum]
		formatValue(arg.Tag, arg.Target, false),
		// TODO: Write a custom ValueFormatter to do: "Help Message. [required] [default=""] etc
		format(arg),
	}
}

func formatCommand(cmd *kong.Command, compact bool) [][]string {
	if compact {
		slog.Warn("Option.Compact currently not supported")
	}
	tags := " "
	if cmd.Tag != nil && (cmd.Tag.Default == "withargs" || cmd.Tag.Default == "1") {
		tags += ColorDefault("(default) ")
	}
	return [][]string{
		{
			"  ",
			ColorCommand(cmd.Path()),
			cmd.Help + tags,
		},
	}
}

func formatFlag(flag *kong.Flag, format kong.HelpValueFormatter) []string {
	if flag == nil {
		return []string{}
	}
	value := flag.Value
	if value == nil {
		return []string{}
	}

	var prefix = "  "
	if value.Tag != nil && value.Tag.Required {
		prefix = ColorRequired("* ")
	}
	var flagStr = "  "
	if flag.Short != 0 {
		flagStr = ColorFlag(fmt.Sprintf("-%c", flag.Short))
	}

	if value.Name != "" {
		long := ColorFlag("--" + value.Name)
		if flagStr == "  " {
			flagStr += "  " + long
		} else {
			flagStr += ", " + long
		}
		if len(flag.Aliases) > 0 {
			for _, alias := range flag.Aliases {
				flagStr += "," + ColorFlag("--"+alias)
			}
		}
		if placeholder := flag.PlaceHolder; placeholder != "" {
			flagStr += fmt.Sprintf(`=%s`, ColorPlaceHolder(placeholder))
		} else if tag := value.Tag; tag != nil && tag.HasDefault {
			var q string
			if value.Target.Kind() == reflect.String {
				q = `"`
			}
			flagStr += fmt.Sprintf(`=%s`, ColorDefault(q+tag.Default+q))
		}
	}

	return []string{
		prefix,
		flagStr,
		// TODO: Parse format/enum to do along the line of PATH[existing file] or STRING[enum]
		formatValue(value.Tag, value.Target, false),
		// TODO: Write a custom ValueFormatter to do: "Help Message. [required] [default=""] etc
		format(value),
	}
}
