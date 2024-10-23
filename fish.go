package cli

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"text/template"
)

// ToFishCompletion creates a fish completion string for the `*App`
// The function errors if either parsing or writing of the string fails.
func (cmd *Command) ToFishCompletion() (string, error) {
	var w bytes.Buffer
	if err := cmd.writeFishCompletionTemplate(&w); err != nil {
		return "", err
	}
	return w.String(), nil
}

type fishCommandCompletionTemplate struct {
	Command     *Command
	Completions []string
	AllCommands []string
}

func (cmd *Command) writeFishCompletionTemplate(w io.Writer) error {
	const name = "cli"
	t, err := template.New(name).Parse(FishCompletionTemplate)
	if err != nil {
		return err
	}
	allCommands := []string{}

	// Add global flags
	completions := cmd.prepareFishFlags(cmd.VisibleFlags(), allCommands)

	// Add help flag
	if !cmd.HideHelp {
		completions = append(
			completions,
			cmd.prepareFishFlags([]Flag{HelpFlag}, allCommands)...,
		)
	}

	// Add version flag
	if !cmd.HideVersion {
		completions = append(
			completions,
			cmd.prepareFishFlags([]Flag{VersionFlag}, allCommands)...,
		)
	}

	// Add commands and their flags
	completions = append(
		completions,
		cmd.prepareFishCommands(cmd.VisibleCommands(), &allCommands, []string{})...,
	)

	return t.ExecuteTemplate(w, name, &fishCommandCompletionTemplate{
		Command:     cmd,
		Completions: completions,
		AllCommands: allCommands,
	})
}

func (cmd *Command) prepareFishCommands(commands []*Command, allCommands *[]string, previousCommands []string) []string {
	completions := []string{}
	for _, command := range commands {
		var completion strings.Builder
		completion.WriteString(fmt.Sprintf(
			"complete -r -c %s -n '%s' -a '%s'",
			cmd.Name,
			cmd.fishSubcommandHelper(previousCommands),
			strings.Join(command.Names(), " "),
		))

		if command.Usage != "" {
			completion.WriteString(fmt.Sprintf(" -d '%s'",
				escapeSingleQuotes(command.Usage)))
		}

		if !command.HideHelp {
			completions = append(
				completions,
				cmd.prepareFishFlags([]Flag{HelpFlag}, command.Names())...,
			)
		}

		*allCommands = append(*allCommands, command.Names()...)
		completions = append(completions, completion.String())
		completions = append(
			completions,
			cmd.prepareFishFlags(command.VisibleFlags(), command.Names())...,
		)

		// recursively iterate subcommands
		if len(command.Commands) > 0 {
			completions = append(
				completions,
				cmd.prepareFishCommands(
					command.Commands, allCommands, command.Names(),
				)...,
			)
		}
	}

	return completions
}

func (cmd *Command) prepareFishFlags(flags []Flag, previousCommands []string) []string {
	completions := []string{}
	for _, f := range flags {
		completion := &strings.Builder{}
		completion.WriteString(fmt.Sprintf(
			"complete -c %s -n '%s'",
			cmd.Name,
			cmd.fishSubcommandHelper(previousCommands),
		))

		fishAddFileFlag(f, completion)

		for idx, opt := range f.Names() {
			if idx == 0 {
				completion.WriteString(fmt.Sprintf(
					" -l %s", strings.TrimSpace(opt),
				))
			} else {
				completion.WriteString(fmt.Sprintf(
					" -s %s", strings.TrimSpace(opt),
				))
			}
		}

		if flag, ok := f.(DocGenerationFlag); ok {
			if flag.TakesValue() {
				completion.WriteString(" -r")
			}

			if flag.GetUsage() != "" {
				completion.WriteString(fmt.Sprintf(" -d '%s'",
					escapeSingleQuotes(flag.GetUsage())))
			}
		}

		completions = append(completions, completion.String())
	}

	return completions
}

func fishAddFileFlag(flag Flag, completion *strings.Builder) {
	switch f := flag.(type) {
	case *StringFlag:
		if f.TakesFile {
			return
		}
	case *StringSliceFlag:
		if f.TakesFile {
			return
		}
	}
	completion.WriteString(" -f")
}

func (cmd *Command) fishSubcommandHelper(allCommands []string) string {
	fishHelper := fmt.Sprintf("__fish_%s_no_subcommand", cmd.Name)
	if len(allCommands) > 0 {
		fishHelper = fmt.Sprintf(
			"__fish_seen_subcommand_from %s",
			strings.Join(allCommands, " "),
		)
	}
	return fishHelper
}

func escapeSingleQuotes(input string) string {
	return strings.Replace(input, `'`, `\'`, -1)
}
