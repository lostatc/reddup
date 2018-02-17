package main

import (
	"log"
	"fmt"
	"bufio"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/urfave/cli"

	"github.com/lostatc/reddup/parse"
	"github.com/lostatc/reddup/paths"
	"io"
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

// This is the number of spaces of padding to put between columns in the output of "list."
const listPadding = 2

func main() {
	cli.AppHelpTemplate = appHelpTemplate
	cli.CommandHelpTemplate = commandHelpTemplate
	cli.HelpFlag = cli.BoolFlag {
		Name: "help",
		Usage: "Show help.",
	}
	cli.VersionFlag = cli.BoolFlag {
		Name: "version",
		Usage: "Display the version.",
	}

	app := cli.NewApp()
	app.Name = "reddup"
	app.Version = "0.1.0"
	app.Authors = []cli.Author {
		cli.Author {
			Name: "Garrett Powell",
			Email: "garrett@gpowell.net",
		},
	}
	app.Copyright = "Copyright © 2017-2018 Garrett Powell <garrett@gpowell.net>"
	app.Usage = "Clean up unused files."

	app.Flags = []cli.Flag {
		cli.StringSliceFlag {
			Name: "exclude",
			Usage: "Exclude files that match this `pattern`.",
		},
		cli.StringFlag {
			Name: "exclude-from",
			Usage: "Exclude files that match patterns in this `file`.",
		},
		cli.StringFlag {
			Name:  "min-time, t",
			Usage: "Only include files which were last modified at least this much `time` in the past. This accepts the units 'h,' 'd,' 'm'  and 'y.'",
			Value: "0h",
		},
		cli.HelpFlag,
	}

	app.Commands = []cli.Command {
		cli.Command {
			Name: "list",
			Usage: "Print a list of files that should be cleaned up.",
			Description: "Print a list of up to size bytes of files in the directory source that should be cleaned up. Also print the size and last access time of each file.",
			ArgsUsage: "size source",
			Flags: []cli.Flag{
				cli.BoolFlag {
					Name: "paths-only",
					Usage: "Print only a list of newline-delimited file paths.",
				},
			},
			Before: enforceArgs(2),
			Action: list,
		},
		cli.Command {
			Name: "move",
			Usage: "Move files that should be cleaned up, prompting the user for confirmation first.",
			Description: "Move up to size bytes of files that should be cleaned up from source to dest (e.g. 10GiB)t.",
			ArgsUsage: "size source dest",
			Flags: []cli.Flag {
				cli.BoolFlag {
					Name: "structure",
					Usage: "Preserve the file structure.",
				},
				cli.BoolFlag {
					Name: "no-prompt",
					Usage: "Don't prompt the user for confirmation before moving files.",
				},
			},
			Before: enforceArgs(3),
			Action: move,
		},
		cli.Command {
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
	return func(c *cli.Context) error {
		if len(c.Args()) < numArgs {
			return fmt.Errorf("not enough arguments")
		} else if len(c.Args()) > numArgs {
			return fmt.Errorf("too many arguments")
		}
		return nil
	}
}

// list executes the 'list' command.
func list(c *cli.Context) (err error) {
	delPaths := getPaths(c)

	if c.Bool("paths-only") {
		// Just print the file paths.
		for _, filePath := range delPaths {
			fmt.Println(filePath.Path)
		}
	} else {
		// Print additional information with the file paths.
		printPaths(os.Stdout, delPaths)
	}

	return nil
}

// move executes the 'move' command.
func move(c *cli.Context) (err error) {
	delPaths := getPaths(c)
	sourceDir := c.Args()[1]
	destDir := c.Args()[2]

	moveFiles := true
	var selectedPaths paths.FilePaths
	if c.Bool("no-prompt") {
		selectedPaths = delPaths
	} else {
		// Print all file paths.
		printPaths(os.Stdout, delPaths)

		// Prompt the user to choose the file to transfer.
		var selectedNumbers []int
		for {
			fmt.Println("\nSelect which files to transfer. You can specify comma-separated ranges of numbers (e.g. '1-9,15,17-20'). Leave blank to select all files.")
			fmt.Print("> ")
			numberRanges := readInput()
			selectedNumbers, err = parse.ReadNumberRanges(numberRanges)
			if err == nil {
				break
			}
		}
		if len(selectedNumbers) == 0 {
			selectedPaths = delPaths
		} else {
			for _, num := range selectedNumbers {
				selectedPaths = append(selectedPaths, delPaths[num - 1])
			}
		}

		// Prompt the user to confirm the file transfer.
		fmt.Println()
		printPaths(os.Stdout, selectedPaths)
		fmt.Printf("\nMove these %d files? [y/N] ", len(selectedPaths))
		confirmation := readInput()
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
		if c.Bool("structure") {
			paths.MoveStructuredFiles(sourceDir, selectedPaths, destDir)
		} else {
			paths.MoveFiles(selectedPaths, destDir)
		}
		fmt.Printf("%d files moved\n", len(selectedPaths))
	} else {
		fmt.Println("0 files moved")
	}

	return nil
}

// getPaths returns the paths of files that should be cleaned up based on the
// given arguments.
func getPaths(c *cli.Context) (delPaths paths.FilePaths) {
	// Parse arguments.
	maxSize, err := parse.ReadFileSize(c.Args()[0])
	if err != nil {
		log.Fatal(err)
	}
	startDir := c.Args()[1]
	minDuration, err := parse.ReadDuration(c.GlobalString("min-time"))
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
	if c.GlobalString("exclude-from") == "" {
		exclude = new(paths.Exclude)
	} else {
		exclude, err = paths.NewExcludeFromFile(c.GlobalString("exclude-from"))
		if err != nil {
			log.Fatal(err)
		}
	}
	for _, pattern := range c.GlobalStringSlice("exclude") {
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
	delPaths = paths.DuplicateFilter(nonExcludedPaths, maxSize, minDuration)

	return delPaths
}

// printPaths prints a formatted table of information about each path in
// pathsToPrint to output. This includes the path, size, last access time and
// whether the file is a duplicate.
func printPaths(output io.Writer, pathsToPrint paths.FilePaths) {
	writer := tabwriter.NewWriter(output, 0, 0, listPadding, ' ', 0)
	fmt.Fprintln(writer, "#\tSize\tLast Access\tDuplicate\tPath")

	for i, filePath := range pathsToPrint {
		var isDuplicate string
		if (filePath.Flags & paths.FlagDuplicate) == paths.FlagDuplicate {
			isDuplicate = "Yes"
		} else {
			isDuplicate = "No"
		}

		fmt.Fprintf(
			writer, "%d\t%s\t%v\t%s\t%s\n",
			i + 1,
			parse.FormatFileSize(filePath.Stat.Size()),
			filePath.Time.AccessTime().Format("Jan 1 2006 15:04"),
			isDuplicate,
			filePath.Path)
	}
	writer.Flush()
}

// readInput reads a line from stdin.
func readInput() string {
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}

	return input
}
