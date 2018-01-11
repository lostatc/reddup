package main

import (
	"log"

	"github.com/urfave/cli"

	"github.com/lostatc/reddup/input"
	"github.com/lostatc/reddup/paths"
	"fmt"
)

func main() {
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
	app.ArgsUsage = "[global_opts] command [command_args]"

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
	}

	app.Commands = []cli.Command{
		cli.Command{
			Name:  "list",
			Usage: "size source",
			Description: "Print a list of files that should be cleaned up.",
			ArgsUsage: "Print a list of up to size bytes of files in the directory source that should be cleaned up.",
			Action: func(context *cli.Context) (err error) {
				return err
			},
		},
		cli.Command{
			Name: "move",
			Usage: "size source dest",
			Description: "Move files that should be cleaned up, prompting the user for confirmation first.",
			ArgsUsage: "Move up to size bytes of files that should be cleaned up from source to dest. Accepted units for size are 'KB,' 'MB,' 'GB,' 'KiB,' 'MiB' and 'GiB.'",
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
			Action: func(context *cli.Context) (err error) {
				return err
			},
		},
	}
}

// getPaths returns the paths of files that should be cleaned up based on the
// given arguments.
func getPaths(context *cli.Context) (delPaths []paths.FilePath) {
	// Parse arguments.
	startDir := context.Args()[0]
	maxSize, err := input.FileSize(context.Args()[1])
	if err != nil {
		log.Fatal(err)
	}
	minDuration, err := input.Duration(context.String("min-time"))
	if err != nil {
		log.Fatal(err)
	}

	// Find all paths in the directory.
	allPaths, err := paths.ScanTree(startDir, paths.ModeFile)
	if err != nil {
		log.Fatal(err)
	}

	// Parse the exclude patterns or exclude pattern file if given.
	var exclude *paths.Exclude
	if context.String("exclude-from") == "" {
		exclude = new(paths.Exclude)
	} else {
		exclude, err = paths.NewExcludeFromFile(context.String("exclude-from"))
		if err != nil {
			log.Fatal(err)
		}
	}
	for _, pattern := range context.StringSlice("exclude") {
		exclude.Patterns = append(exclude.Patterns, pattern)
	}

	// Ignore paths that match exclude patterns.
	var nonExcludedPaths []paths.FilePath
	for _, filePath := range allPaths {
		if !exclude.CheckMatch(filePath.Path, startDir) {
			nonExcludedPaths = append(nonExcludedPaths, filePath)
		}
	}

	delPaths = paths.Filter(nonExcludedPaths, maxSize, minDuration)

	return delPaths
}
