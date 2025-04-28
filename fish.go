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

	// Add global flags
	completions := cmd.prepareFishFlags(cmd.VisibleFlags(), []string{})

	// Add commands and their flags
	completions = append(
		completions,
		cmd.prepareFishCommands(cmd.Commands, []string{})...,
	)

	toplevelCommandNames := []string{}
	for _, child := range cmd.Commands {
		toplevelCommandNames = append(toplevelCommandNames, child.Names()...)
	}

	return t.ExecuteTemplate(w, name, &fishCommandCompletionTemplate{
		Command:     cmd,
		Completions: completions,
		AllCommands: toplevelCommandNames,
	})
}

func (cmd *Command) prepareFishCommands(commands []*Command, previousCommands []string) []string {
	completions := []string{}
	for _, command := range commands {
		if !command.Hidden {
			var completion strings.Builder
			fmt.Fprintf(&completion,
				"complete -x -c %s -n '%s' -a '%s'",
				cmd.Name,
				cmd.fishSubcommandHelper(previousCommands, commands),
				strings.Join(command.Names(), " "),
			)

			if command.Usage != "" {
				fmt.Fprintf(&completion,
					" -d '%s'",
					escapeSingleQuotes(command.Usage))
			}
			completions = append(completions, completion.String())
		}
		completions = append(
			completions,
			cmd.prepareFishFlags(command.VisibleFlags(), command.Names())...,
		)

		// recursively iterate subcommands
		if len(command.Commands) > 0 {
			completions = append(
				completions,
				cmd.prepareFishCommands(
					command.Commands, command.Names(),
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
		fmt.Fprintf(completion,
			"complete -c %s -n '%s'",
			cmd.Name,
			cmd.fishFlagHelper(previousCommands),
		)

		fishAddFileFlag(f, completion)

		for idx, opt := range f.Names() {
			if idx == 0 {
				fmt.Fprintf(completion,
					" -l %s", strings.TrimSpace(opt),
				)
			} else {
				fmt.Fprintf(completion,
					" -s %s", strings.TrimSpace(opt),
				)
			}
		}

		if flag, ok := f.(DocGenerationFlag); ok {
			if flag.TakesValue() {
				completion.WriteString(" -r")
			}

			if flag.GetUsage() != "" {
				fmt.Fprintf(completion,
					" -d '%s'",
					escapeSingleQuotes(flag.GetUsage()))
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

func (cmd *Command) fishSubcommandHelper(allCommands []string, siblings []*Command) string {
	fishHelper := fmt.Sprintf("__fish_%s_no_subcommand", cmd.Name)
	if len(allCommands) > 0 {
		var siblingNames []string
		for _, command := range siblings {
			siblingNames = append(siblingNames, command.Names()...)
		}
		fishHelper = fmt.Sprintf(
			"__fish_seen_subcommand_from %s; and not __fish_seen_subcommand_from %s",
			strings.Join(allCommands, " "),
			strings.Join(siblingNames, " "),
		)
	}
	return fishHelper
}

func (cmd *Command) fishFlagHelper(allCommands []string) string {
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
	return strings.ReplaceAll(input, `'`, `\'`)
}
