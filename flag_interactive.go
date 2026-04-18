package cli

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

type InteractiveFlag = FlagBase[bool, InteractiveConfig, interactiveValue]

type InteractiveConfig struct {
	Prompt       string
	DefaultValue string
	Required     bool
}

type interactiveValue struct {
	destination *bool
}

func (iv interactiveValue) Create(val bool, p *bool, c InteractiveConfig) Value {
	*p = val
	return &interactiveValue{
		destination: p,
	}
}

func (iv interactiveValue) ToString(value bool) string {
	iv.destination = &value
	return iv.String()
}

func (iv *interactiveValue) Set(s string) error {
	v, err := strconv.ParseBool(s)
	if err != nil {
		return err
	}
	*iv.destination = v
	return nil
}

func (iv *interactiveValue) Get() any { return *iv.destination }

func (iv *interactiveValue) String() string {
	return strconv.FormatBool(*iv.destination)
}

func (iv *interactiveValue) IsBoolFlag() bool { return true }

func (cmd *Command) Interactive(name string) bool {
	if v, ok := cmd.Value(name).(bool); ok {
		tracef("interactive available for flag name %[1]q with value=%[2]v (cmd=%[3]q)", name, v, cmd.Name)
		return v
	}
	tracef("interactive NOT available for flag name %[1]q (cmd=%[2]q)", name, cmd.Name)
	return false
}

type InteractivePrompter interface {
	Prompt(prompt string, defaultValue string) (string, error)
	PromptRequired(prompt string) (string, error)
	PromptConfirm(prompt string, defaultValue bool) (bool, error)
	PromptSelect(prompt string, options []string) (int, string, error)
}

type DefaultPrompter struct {
	Reader io.Reader
	Writer io.Writer
}

func NewDefaultPrompter() *DefaultPrompter {
	return &DefaultPrompter{
		Reader: os.Stdin,
		Writer: os.Stdout,
	}
}

func (p *DefaultPrompter) Prompt(prompt string, defaultValue string) (string, error) {
	if defaultValue != "" {
		fmt.Fprintf(p.Writer, "%s [%s]: ", prompt, defaultValue)
	} else {
		fmt.Fprintf(p.Writer, "%s: ", prompt)
	}

	scanner := bufio.NewScanner(p.Reader)
	if scanner.Scan() {
		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			return defaultValue, nil
		}
		return input, nil
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return defaultValue, nil
}

func (p *DefaultPrompter) PromptRequired(prompt string) (string, error) {
	for {
		result, err := p.Prompt(prompt, "")
		if err != nil {
			return "", err
		}
		if result != "" {
			return result, nil
		}
		fmt.Fprintln(p.Writer, "This field is required. Please enter a value.")
	}
}

func (p *DefaultPrompter) PromptConfirm(prompt string, defaultValue bool) (bool, error) {
	options := "y/N"
	if defaultValue {
		options = "Y/n"
	}

	for {
		fmt.Fprintf(p.Writer, "%s [%s]: ", prompt, options)

		scanner := bufio.NewScanner(p.Reader)
		if scanner.Scan() {
			input := strings.ToLower(strings.TrimSpace(scanner.Text()))
			if input == "" {
				return defaultValue, nil
			}
			if input == "y" || input == "yes" {
				return true, nil
			}
			if input == "n" || input == "no" {
				return false, nil
			}
			fmt.Fprintln(p.Writer, "Please enter 'y' or 'n'.")
		}

		if err := scanner.Err(); err != nil {
			return false, err
		}
	}
}

func (p *DefaultPrompter) PromptSelect(prompt string, options []string) (int, string, error) {
	if len(options) == 0 {
		return -1, "", fmt.Errorf("no options provided")
	}

	fmt.Fprintln(p.Writer, prompt)
	for i, opt := range options {
		fmt.Fprintf(p.Writer, "  %d. %s\n", i+1, opt)
	}

	for {
		fmt.Fprintf(p.Writer, "Please select an option (1-%d): ", len(options))

		scanner := bufio.NewScanner(p.Reader)
		if scanner.Scan() {
			input := strings.TrimSpace(scanner.Text())
			index, err := strconv.Atoi(input)
			if err == nil && index >= 1 && index <= len(options) {
				return index - 1, options[index-1], nil
			}
			fmt.Fprintf(p.Writer, "Invalid selection. Please enter a number between 1 and %d.\n", len(options))
		}

		if err := scanner.Err(); err != nil {
			return -1, "", err
		}
	}
}

func (cmd *Command) RunInteractive(ctx context.Context, prompter InteractivePrompter) error {
	if prompter == nil {
		prompter = NewDefaultPrompter()
	}

	for _, flag := range cmd.allFlags() {
		if !flag.IsSet() {
			if ip, ok := flag.(InteractivePrompterFlag); ok {
				if ip.ShouldPrompt() {
					err := ip.PromptValue(prompter)
					if err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

type InteractivePrompterFlag interface {
	Flag
	ShouldPrompt() bool
	PromptValue(prompter InteractivePrompter) error
	GetPrompt() string
	GetDefaultValue() string
	IsPromptRequired() bool
}

type InteractiveStringFlag struct {
	StringFlag
	Prompt       string
	Required     bool
}

func (f *InteractiveStringFlag) ShouldPrompt() bool {
	return f.Prompt != ""
}

func (f *InteractiveStringFlag) GetPrompt() string {
	return f.Prompt
}

func (f *InteractiveStringFlag) GetDefaultValue() string {
	return f.Value
}

func (f *InteractiveStringFlag) IsPromptRequired() bool {
	return f.Required
}

func (f *InteractiveStringFlag) PromptValue(prompter InteractivePrompter) error {
	var value string
	var err error

	if f.Required {
		value, err = prompter.PromptRequired(f.Prompt)
	} else {
		value, err = prompter.Prompt(f.Prompt, f.Value)
	}

	if err != nil {
		return err
	}

	return f.Set(f.Name, value)
}

type InteractiveIntFlag struct {
	Int64Flag
	Prompt       string
	Required     bool
}

func (f *InteractiveIntFlag) ShouldPrompt() bool {
	return f.Prompt != ""
}

func (f *InteractiveIntFlag) GetPrompt() string {
	return f.Prompt
}

func (f *InteractiveIntFlag) GetDefaultValue() string {
	return strconv.FormatInt(f.Value, 10)
}

func (f *InteractiveIntFlag) IsPromptRequired() bool {
	return f.Required
}

func (f *InteractiveIntFlag) PromptValue(prompter InteractivePrompter) error {
	for {
		var value string
		var err error

		if f.Required {
			value, err = prompter.PromptRequired(f.Prompt)
		} else {
			value, err = prompter.Prompt(f.Prompt, strconv.FormatInt(f.Value, 10))
		}

		if err != nil {
			return err
		}

		if value == "" && !f.Required {
			return nil
		}

		_, err = strconv.ParseInt(value, 10, 64)
		if err == nil {
			return f.Set(f.Name, value)
		}

		fmt.Println("Invalid integer value. Please try again.")
	}
}

type InteractiveBoolFlag struct {
	BoolFlag
	Prompt       string
	Required     bool
}

func (f *InteractiveBoolFlag) ShouldPrompt() bool {
	return f.Prompt != ""
}

func (f *InteractiveBoolFlag) GetPrompt() string {
	return f.Prompt
}

func (f *InteractiveBoolFlag) GetDefaultValue() string {
	return strconv.FormatBool(f.Value)
}

func (f *InteractiveBoolFlag) IsPromptRequired() bool {
	return f.Required
}

func (f *InteractiveBoolFlag) PromptValue(prompter InteractivePrompter) error {
	value, err := prompter.PromptConfirm(f.Prompt, f.Value)
	if err != nil {
		return err
	}

	return f.Set(f.Name, strconv.FormatBool(value))
}

type InteractiveFloatFlag struct {
	FloatFlag
	Prompt       string
	Required     bool
}

func (f *InteractiveFloatFlag) ShouldPrompt() bool {
	return f.Prompt != ""
}

func (f *InteractiveFloatFlag) GetPrompt() string {
	return f.Prompt
}

func (f *InteractiveFloatFlag) GetDefaultValue() string {
	return strconv.FormatFloat(f.Value, 'f', -1, 64)
}

func (f *InteractiveFloatFlag) IsPromptRequired() bool {
	return f.Required
}

func (f *InteractiveFloatFlag) PromptValue(prompter InteractivePrompter) error {
	for {
		var value string
		var err error

		if f.Required {
			value, err = prompter.PromptRequired(f.Prompt)
		} else {
			value, err = prompter.Prompt(f.Prompt, strconv.FormatFloat(f.Value, 'f', -1, 64))
		}

		if err != nil {
			return err
		}

		if value == "" && !f.Required {
			return nil
		}

		_, err = strconv.ParseFloat(value, 64)
		if err == nil {
			return f.Set(f.Name, value)
		}

		fmt.Println("Invalid float value. Please try again.")
	}
}

type InteractiveDurationFlag struct {
	DurationFlag
	Prompt       string
	Required     bool
}

func (f *InteractiveDurationFlag) ShouldPrompt() bool {
	return f.Prompt != ""
}

func (f *InteractiveDurationFlag) GetPrompt() string {
	return f.Prompt
}

func (f *InteractiveDurationFlag) GetDefaultValue() string {
	return f.Value.String()
}

func (f *InteractiveDurationFlag) IsPromptRequired() bool {
	return f.Required
}

func (f *InteractiveDurationFlag) PromptValue(prompter InteractivePrompter) error {
	for {
		var value string
		var err error

		if f.Required {
			value, err = prompter.PromptRequired(f.Prompt)
		} else {
			value, err = prompter.Prompt(f.Prompt, f.Value.String())
		}

		if err != nil {
			return err
		}

		if value == "" && !f.Required {
			return nil
		}

		_, err = time.ParseDuration(value)
		if err == nil {
			return f.Set(f.Name, value)
		}

		fmt.Println("Invalid duration value (e.g., 1s, 2m, 3h). Please try again.")
	}
}

var InteractiveFlagInstance Flag = &BoolFlag{
	Name:    "interactive",
	Aliases: []string{"i"},
	Usage:   "enable interactive mode for missing parameters",
}

func (cmd *Command) shouldRunInteractive() bool {
	for _, flag := range cmd.allFlags() {
		for _, name := range flag.Names() {
			if name == "interactive" || name == "i" {
				if flag.IsSet() {
					if v, ok := flag.Get().(bool); ok {
						return v
					}
				}
			}
		}
	}
	return false
}

func (cmd *Command) handleInteractiveMode(ctx context.Context) error {
	prompter := NewDefaultPrompter()

	if cmd.Reader != nil {
		prompter.Reader = cmd.Reader
	}
	if cmd.Writer != nil {
		prompter.Writer = cmd.Writer
	}

	return cmd.RunInteractive(ctx, prompter)
}
