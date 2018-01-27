package main

import (
	"log"
	"fmt"
	"bufio"
	"os"
	"strings"

	"github.com/urfave/cli"

	"github.com/lostatc/reddup/input"
	"github.com/lostatc/reddup/paths"
)

const appHelpTemplate = `Usage:
    {{if .UsageText}}{{.UsageText}}{{else}}{{.HelpName}} {{if .VisibleFlags}}[global_options]{{end}}{{if .Commands}} command [command_args]{{end}}{{end}}{{if .VisibleFlags}}

Global Options:
    {{range $index, $option := .VisibleFlags}}{{if $index}}
    {{end}}{{$option}}{{end}}{{end}}{{if .VisibleCommands}}

Commands:{{range .VisibleCategories}}{{if .Name}}
    {{.Name}}:{{end}}{{range .VisibleCommands}}
    {{join .Names ", "}}{{"\t"}}{{.Usage}}{{end}}{{end}}{{end}}
`

const commandHelpTemplate = `Usage:
    {{if .UsageText}}{{.UsageText}}{{else}}{{.HelpName}}{{if .VisibleFlags}} [command_options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}{{end}}{{if .Category}}

Category:
    {{.Category}}{{end}}{{if .Description}}

Description:
    {{.Description}}{{end}}{{if .VisibleFlags}}

Options:
    {{range .VisibleFlags}}{{.}}
    {{end}}{{end}}
`

func main() {
	cli.AppHelpTemplate = appHelpTemplate
	cli.CommandHelpTemplate = commandHelpTemplate
	cli.HelpFlag = cli.BoolFlag{
		Name: "help",
		Usage: "Show help.",
	}
	cli.VersionFlag = cli.BoolFlag{
		Name: "version",
		Usage: "Display the version.",
	}

	app := cli.NewApp()
	app.Name = "reddup"
	app.Version = "0.1.0"
	app.Authors = []cli.Author{
		cli.Author{
			Name: "Garrett Powell",
			Email: "garrett@gpowell.net",
		},
	}
	app.Copyright = "Copyright © 2017-2018 Garrett Powell <garrett@gpowell.net>"
	app.Usage = "Clean up unused files."

	app.Flags = []cli.Flag{
		cli.StringSliceFlag{
			Name: "exclude",
			Usage: "Exclude files that match this `pattern`.",
		},
		cli.StringFlag{
			Name: "exclude-from",
			Usage: "Exclude files that match patterns in this `file`.",
		},
		cli.StringFlag{
			Name:  "min-time, t",
			Usage: "Only include files which were last modified at least this much `time` in the past. This accepts the units 'h,' 'd,' 'm'  and 'y.'",
			Value: "0h",
		},
		cli.HelpFlag,
	}

	app.Commands = []cli.Command{
		cli.Command{
			Name: "list",
			Usage: "Print a list of files that should be cleaned up.",
			Description: "Print a list of up to size bytes of files in the directory source that should be cleaned up.",
			ArgsUsage: "size source",
			Before: enforceArgs(2),
			Action: list,
		},
		cli.Command{
			Name: "move",
			Usage: "Move files that should be cleaned up, prompting the user for confirmation first.",
			Description: "Move up to size bytes of files that should be cleaned up from source to dest. Accepted units for size are 'KB,' 'MB,' 'GB,' 'KiB,' 'MiB' and 'GiB.'",
			ArgsUsage: "size source dest",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name: "structure",
					Usage: "Preserve the file structure.",
				},
				cli.BoolFlag{
					Name: "no-prompt",
					Usage: "Don't prompt the user for confirmation before moving files.",
				},
			},
			Before: enforceArgs(3),
			Action: move,
		},
		cli.Command{
			Name: "help",
			Usage: "Show a list of commands or help for one command.",
			ArgsUsage: "[command]",
			Action: func(c *cli.Context) error {
				args := c.Args()
				if args.Present() {
					return cli.ShowCommandHelp(c, args.First())
				}

				cli.ShowAppHelp(c)
				return nil
			},
		},
	}

	app.Run(os.Args)
}

// enforceArgs returns a function that enforces a specific number of arguments.
func enforceArgs(numArgs int) cli.BeforeFunc {
	return func(context *cli.Context) error {
		if len(context.Args()) < numArgs {
			return fmt.Errorf("not enough arguments")
		} else if len(context.Args()) > numArgs {
			return fmt.Errorf("too many arguments")
		}
		return nil
	}
}

// list executes the 'list' command.
func list(context *cli.Context) (err error) {
	delPaths := getPaths(context)
	for _, filePath := range delPaths {
		fmt.Println(filePath.Path)
	}
	return nil
}

// move executes the 'move' command.
func move(context *cli.Context) (err error) {
	delPaths := getPaths(context)
	sourceDir := context.Args()[1]
	destDir := context.Args()[2]

	moveFiles := true
	if context.Bool("no-prompt") == false {
		// Print the files to transfer.
		for _, filePath := range delPaths {
			fmt.Println(filePath.Path)
		}

		// Prompt the user to confirm the file transfer.
		fmt.Printf("Move these %d files? [y/N] ", len(delPaths))
		reader := bufio.NewReader(os.Stdin)
		confirmation, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		confirmation = strings.ToLower(confirmation)
		confirmation = strings.TrimSuffix(confirmation, "\n")
		switch confirmation {
		case "y", "yes":
		default:
			moveFiles = false
		}
	}

	if moveFiles {
		// Move the files.
		if context.Bool("structure") {
			paths.MoveStructuredFiles(sourceDir, delPaths, destDir)
		} else {
			paths.MoveFiles(delPaths, destDir)
		}
		fmt.Printf("%d files moved\n", len(delPaths))
	} else {
		fmt.Println("0 files moved")
	}

	return nil
}

// getPaths returns the paths of files that should be cleaned up based on the
// given arguments.
func getPaths(context *cli.Context) (delPaths paths.FilePaths) {
	// Parse arguments.
	maxSize, err := input.FileSize(context.Args()[0])
	if err != nil {
		log.Fatal(err)
	}
	startDir := context.Args()[1]
	minDuration, err := input.Duration(context.GlobalString("min-time"))
	if err != nil {
		log.Fatal(err)
	}

	// Find all paths in the directory.
	allPaths, err := paths.ScanTree(startDir, paths.ModeFile)
	if err != nil {
		log.Fatal(err)
	}

	// Parse the exclude patterns or the exclude pattern file if given.
	var exclude *paths.Exclude
	if context.GlobalString("exclude-from") == "" {
		exclude = new(paths.Exclude)
	} else {
		exclude, err = paths.NewExcludeFromFile(context.GlobalString("exclude-from"))
		if err != nil {
			log.Fatal(err)
		}
	}
	for _, pattern := range context.GlobalStringSlice("exclude") {
		exclude.Patterns = append(exclude.Patterns, pattern)
	}

	// Ignore paths that match exclude patterns.
	var nonExcludedPaths paths.FilePaths
	for _, filePath := range allPaths {
		if !exclude.CheckMatch(filePath.Path, startDir) {
			nonExcludedPaths = append(nonExcludedPaths, filePath)
		}
	}

	// Select paths to be cleaned up. Duplicate files are selected first.
	duplicatePaths := paths.GetNewestDuplicates(nonExcludedPaths)
	delPaths = append(delPaths, duplicatePaths...)
	maxSize -= duplicatePaths.TotalSize()
	delPaths = append(delPaths, paths.Filter(nonExcludedPaths.Difference(duplicatePaths), maxSize, minDuration)...)

	return delPaths
}
