/*
Copyright © 2018 Garrett Powell <garrett@gpowell.net>

This file is part of reddup.

reddup is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

reddup is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with reddup.  If not, see <http://www.gnu.org/licenses/>.
*/

/*
Reddup is a program for cleaning up unused files.
*/
package main

import (
	"log"
	"fmt"
	"bufio"
	"os"
	"io"
	"strings"
	"text/tabwriter"
	"sort"

	"github.com/urfave/cli"

	"github.com/lostatc/reddup/parse"
	"github.com/lostatc/reddup/paths"
)

const appHelpTemplate = `Usage:
    {{if .UsageText}}{{.UsageText}}{{else}}{{.HelpName}} {{if .VisibleFlags}}[global_options]{{end}}{{if .Commands}} <command> [command_args]{{end}}{{end}}{{if .VisibleFlags}}

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
	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Fprintf(c.App.Writer, "%v %v\n", c.App.Name, c.App.Version)
	}
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
			Usage: "Exclude files that match this shell globbing `<pattern>`.",
		},
		cli.StringFlag {
			Name: "exclude-from",
			Usage: "Exclude files that match shell globbing patterns in this `<file>`.",
		},
		cli.StringFlag {
			Name:  "min-time, t",
			Usage: "Only include files which were last modified at least this much `<time>` in the past. This accepts the units 'h,' 'd,' 'm'  and 'y.'",
			Value: "0h",
		},
		cli.BoolFlag {
			Name: "no-duplicates",
			Usage: "Don't automatically include duplicate files.",
		},
		cli.HelpFlag,
	}

	app.Commands = []cli.Command {
		cli.Command {
			Name: "list",
			Usage: "Print a list of files that should be cleaned up.",
			Description: "Print a list of up to <size> bytes of files (e.g. 10GiB) in the directory <source> that should be cleaned up. For each file, also print its size, last access time and whether it is a duplicate.",
			ArgsUsage: "<size> <source>",
			UseShortOptionHandling: true,
			Flags: []cli.Flag{
				cli.BoolFlag {
					Name: "paths-only",
					Usage: "Print only a list of newline-separated file paths.",
				},
			},
			Before: enforceArgs(2),
			Action: list,
		},
		cli.Command {
			Name: "move",
			Usage: "Move files that should be cleaned up, prompting the user for confirmation first.",
			Description: "Move up to <size> bytes of files (e.g. 10GiB) that should be cleaned up from <source> to <dest>. Prompt the user for confirmation before moving anything.",
			ArgsUsage: "<size> <source> <dest>",
			UseShortOptionHandling: true,
			Flags: []cli.Flag {
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
			UseShortOptionHandling: true,
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
		err := paths.MoveStructuredFiles(sourceDir, selectedPaths, destDir)
		if err != nil {
			return err
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

	// Find duplicate paths if applicable.
	var duplicatePaths paths.FilePaths
	if !c.GlobalBool("no-duplicates") {
		duplicatePaths = paths.GetOldestDuplicates(nonExcludedPaths)
		sort.Slice(duplicatePaths, func(i, j int) bool {
			return duplicatePaths[j].Stat.Size() < duplicatePaths[i].Stat.Size()
		})
		nonExcludedPaths = nonExcludedPaths.Difference(duplicatePaths)
	}

	// Select non-duplicate paths to be cleaned up.
	delPaths = append(duplicatePaths, paths.Filter(nonExcludedPaths, maxSize, minDuration)...)

	// Assign a piece of metadata to each file path so that they can retain
	// their original rank even if the returned slice is modified.
	for i := range delPaths {
		delPaths[i].Metadata.Rank = i + 1
	}

	return delPaths
}

// printPaths prints a formatted table of information about each FilePath in
// pathsToPrint to output. This includes the file's rank, path, size, last
// access time and whether the file is a duplicate.
func printPaths(output io.Writer, pathsToPrint paths.FilePaths) {
	writer := tabwriter.NewWriter(output, 0, 0, listPadding, ' ', 0)
	fmt.Fprintln(writer, "#\tSize\tLast Access\tDuplicate\tPath")

	for _, filePath := range pathsToPrint {
		var isDuplicate string
		if filePath.Metadata.Duplicate {
			isDuplicate = "Yes"
		} else {
			isDuplicate = "No"
		}

		fmt.Fprintf(
			writer, "%d\t%s\t%v\t%s\t%s\n",
			filePath.Metadata.Rank,
			parse.FormatFileSize(filePath.Stat.Size()),
			filePath.Time.AccessTime().Format("Jan 01 2006 15:04"),
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
