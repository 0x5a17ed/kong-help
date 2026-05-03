package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"

	konghelp "github.com/0x5a17ed/kong-help"
)

type (
	InitCmd struct {
		Quiet     bool   `short:"q" help:"Only print error and warning messages; all other output will be suppressed."`
		Directory string `help:"If you provide a <directory>, the command is run inside it."`
	}
	StatusCmd struct{}
	AddCmd    struct {
		Paths []string `arg:"" help:"The path to the file to add."`
	}
	CommitCmd struct {
		All     bool   `aliases:"branches" help:"Automatically stage files that have been moddified and deleted, but new files you have not told Git about are not affected."`
		Message string `short:"m" placeholder:"<msg>" help:"Use <msg> as the commit message."`
		File    string `arg:"" default:"test.txt" type:"existingfile" help:"File to commit."`
	}
	PushCmd struct {
		Repository string `arg:"" optional:"" help:"The remote rerpository that is the destination of a push operation."`
		All        bool   `aliases:"branches" default:"false" help:"Push all branches; cannot be used with other <refspec>."`
		Delete     bool   `short:"d" help:"All listed refs are deleted from the remote repository."`
	}
)

func (InitCmd) Run() error   { return nil }
func (StatusCmd) Run() error { return nil }
func (AddCmd) Run() error    { return nil }
func (CommitCmd) Run() error { return nil }
func (PushCmd) Run() error   { return nil }

type CMD struct {
	Version bool   `short:"v" help:"Print the Git suite version that the git program came from."`
	GitDir  string `type:"path" help:"Set the path to the reposity (\".git\" directory)."`

	Init   InitCmd   `cmd:"" help:"Create an empty Git repository or reinitialize an existing one."`
	Status StatusCmd `cmd:"" default:"1" help:"Show the git status of the Git repository."`
	Add    AddCmd    `cmd:"" help:"Add file contents to the index."`
	Commit CommitCmd `cmd:"" help:"Record changes to the repository."`
	Push   PushCmd   `cmd:"" help:"Update remote rerfs along with associated objects."`
}

func main() {
	var cli CMD
	kongCtx := kong.Parse(
		&cli,
		konghelp.Help(),
		kong.Name("git"),
		kong.Description("Subset of git commands to showcase kong-help's output!"),
		kong.ConfigureHelp(kong.HelpOptions{
			WrapUpperBound: -1, // Uses terminal width
			// Tree: true,
			// NoExpandSubcommands: true,
		}),
	)

	if err := kongCtx.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
