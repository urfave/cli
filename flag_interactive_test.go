package cli

import (
	"bytes"
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MockPrompter struct {
	Inputs     []string
	InputIndex int
	Outputs    []string
}

func NewMockPrompter(inputs ...string) *MockPrompter {
	return &MockPrompter{
		Inputs:     inputs,
		InputIndex: 0,
		Outputs:    []string{},
	}
}

func (p *MockPrompter) nextInput() string {
	if p.InputIndex < len(p.Inputs) {
		input := p.Inputs[p.InputIndex]
		p.InputIndex++
		return input
	}
	return ""
}

func (p *MockPrompter) Prompt(prompt string, defaultValue string) (string, error) {
	p.Outputs = append(p.Outputs, prompt)
	input := p.nextInput()
	if input == "" {
		return defaultValue, nil
	}
	return input, nil
}

func (p *MockPrompter) PromptRequired(prompt string) (string, error) {
	for {
		p.Outputs = append(p.Outputs, prompt)
		input := p.nextInput()
		if input != "" {
			return input, nil
		}
	}
}

func (p *MockPrompter) PromptConfirm(prompt string, defaultValue bool) (bool, error) {
	p.Outputs = append(p.Outputs, prompt)
	input := p.nextInput()
	if input == "" {
		return defaultValue, nil
	}
	if input == "y" || input == "Y" || input == "yes" || input == "Yes" {
		return true, nil
	}
	return false, nil
}

func (p *MockPrompter) PromptSelect(prompt string, options []string) (int, string, error) {
	p.Outputs = append(p.Outputs, prompt)
	input := p.nextInput()
	if input == "" {
		return 0, options[0], nil
	}
	index, err := strconv.Atoi(input)
	if err != nil {
		return 0, options[0], nil
	}
	if index >= 1 && index <= len(options) {
		return index - 1, options[index-1], nil
	}
	return 0, options[0], nil
}

func TestDefaultPrompter_Prompt(t *testing.T) {
	var output bytes.Buffer
	input := bytes.NewBufferString("test input\n")

	prompter := &DefaultPrompter{
		Reader: input,
		Writer: &output,
	}

	result, err := prompter.Prompt("Enter something", "default")
	require.NoError(t, err)
	assert.Equal(t, "test input", result)
	assert.Contains(t, output.String(), "Enter something [default]")
}

func TestDefaultPrompter_Prompt_WithDefault(t *testing.T) {
	var output bytes.Buffer
	input := bytes.NewBufferString("\n")

	prompter := &DefaultPrompter{
		Reader: input,
		Writer: &output,
	}

	result, err := prompter.Prompt("Enter something", "default value")
	require.NoError(t, err)
	assert.Equal(t, "default value", result)
}

func TestDefaultPrompter_PromptConfirm(t *testing.T) {
	testCases := []struct {
		name         string
		input        string
		defaultValue bool
		expected     bool
	}{
		{"yes input", "y\n", false, true},
		{"no input", "n\n", true, false},
		{"empty with default true", "\n", true, true},
		{"empty with default false", "\n", false, false},
		{"Yes input", "Yes\n", false, true},
		{"NO input", "NO\n", true, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var output bytes.Buffer
			input := bytes.NewBufferString(tc.input)

			prompter := &DefaultPrompter{
				Reader: input,
				Writer: &output,
			}

			result, err := prompter.PromptConfirm("Confirm?", tc.defaultValue)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestDefaultPrompter_PromptSelect(t *testing.T) {
	options := []string{"Option 1", "Option 2", "Option 3"}

	testCases := []struct {
		name          string
		input         string
		expectedIndex int
		expectedValue string
	}{
		{"select first", "1\n", 0, "Option 1"},
		{"select second", "2\n", 1, "Option 2"},
		{"select third", "3\n", 2, "Option 3"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var output bytes.Buffer
			input := bytes.NewBufferString(tc.input)

			prompter := &DefaultPrompter{
				Reader: input,
				Writer: &output,
			}

			index, value, err := prompter.PromptSelect("Choose an option", options)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedIndex, index)
			assert.Equal(t, tc.expectedValue, value)
			assert.Contains(t, output.String(), "Choose an option")
		})
	}
}

func TestInteractiveStringFlag_PromptValue(t *testing.T) {
	flag := &InteractiveStringFlag{
		StringFlag: StringFlag{
			Name:  "username",
			Value: "default_user",
		},
		Prompt:   "Enter username",
		Required: false,
	}

	err := flag.PreParse()
	require.NoError(t, err)

	mockPrompter := NewMockPrompter("test_user")

	err = flag.PromptValue(mockPrompter)
	require.NoError(t, err)

	assert.Equal(t, "test_user", flag.Get())
	assert.Contains(t, mockPrompter.Outputs[0], "Enter username")
}

func TestInteractiveStringFlag_PromptValue_WithDefault(t *testing.T) {
	flag := &InteractiveStringFlag{
		StringFlag: StringFlag{
			Name:  "username",
			Value: "default_user",
		},
		Prompt:   "Enter username",
		Required: false,
	}

	err := flag.PreParse()
	require.NoError(t, err)

	mockPrompter := NewMockPrompter("")

	err = flag.PromptValue(mockPrompter)
	require.NoError(t, err)

	assert.Equal(t, "default_user", flag.Get())
}

func TestInteractiveIntFlag_PromptValue(t *testing.T) {
	flag := &InteractiveIntFlag{
		Int64Flag: Int64Flag{
			Name:  "age",
			Value: 18,
		},
		Prompt:   "Enter age",
		Required: false,
	}

	err := flag.PreParse()
	require.NoError(t, err)

	mockPrompter := NewMockPrompter("25")

	err = flag.PromptValue(mockPrompter)
	require.NoError(t, err)

	assert.Equal(t, int64(25), flag.Get())
}

func TestInteractiveBoolFlag_PromptValue(t *testing.T) {
	flag := &InteractiveBoolFlag{
		BoolFlag: BoolFlag{
			Name:  "subscribe",
			Value: false,
		},
		Prompt: "Subscribe to newsletter",
	}

	err := flag.PreParse()
	require.NoError(t, err)

	mockPrompter := NewMockPrompter("y")

	err = flag.PromptValue(mockPrompter)
	require.NoError(t, err)

	assert.Equal(t, true, flag.Get())
}

func TestInteractiveFloatFlag_PromptValue(t *testing.T) {
	flag := &InteractiveFloatFlag{
		FloatFlag: FloatFlag{
			Name:  "price",
			Value: 9.99,
		},
		Prompt:   "Enter price",
		Required: false,
	}

	err := flag.PreParse()
	require.NoError(t, err)

	mockPrompter := NewMockPrompter("19.99")

	err = flag.PromptValue(mockPrompter)
	require.NoError(t, err)

	assert.Equal(t, 19.99, flag.Get())
}

func TestInteractiveDurationFlag_PromptValue(t *testing.T) {
	flag := &InteractiveDurationFlag{
		DurationFlag: DurationFlag{
			Name:  "timeout",
			Value: 30 * time.Second,
		},
		Prompt:   "Enter timeout",
		Required: false,
	}

	err := flag.PreParse()
	require.NoError(t, err)

	mockPrompter := NewMockPrompter("1m")

	err = flag.PromptValue(mockPrompter)
	require.NoError(t, err)

	assert.Equal(t, 1*time.Minute, flag.Get())
}

func TestCommand_RunInteractive(t *testing.T) {
	cmd := &Command{
		Flags: []Flag{
			&InteractiveStringFlag{
				StringFlag: StringFlag{
					Name: "name",
				},
				Prompt:   "Enter your name",
				Required: true,
			},
			&InteractiveIntFlag{
				Int64Flag: Int64Flag{
					Name:  "age",
					Value: 18,
				},
				Prompt:   "Enter your age",
				Required: false,
			},
		},
		Action: func(ctx context.Context, cmd *Command) error {
			return nil
		},
	}

	mockPrompter := NewMockPrompter("John", "25")

	err := cmd.RunInteractive(context.Background(), mockPrompter)
	require.NoError(t, err)

	assert.Equal(t, 2, len(mockPrompter.Outputs))
}

func TestInteractivePrompterFlag_Interface(t *testing.T) {
	var _ InteractivePrompterFlag = &InteractiveStringFlag{}
	var _ InteractivePrompterFlag = &InteractiveIntFlag{}
	var _ InteractivePrompterFlag = &InteractiveBoolFlag{}
	var _ InteractivePrompterFlag = &InteractiveFloatFlag{}
	var _ InteractivePrompterFlag = &InteractiveDurationFlag{}
}

func TestInteractiveStringFlag_ShouldPrompt(t *testing.T) {
	flagWithPrompt := &InteractiveStringFlag{
		Prompt: "Enter something",
	}
	assert.True(t, flagWithPrompt.ShouldPrompt())

	flagWithoutPrompt := &InteractiveStringFlag{}
	assert.False(t, flagWithoutPrompt.ShouldPrompt())
}

func TestCommand_InteractiveMethod(t *testing.T) {
	cmd := &Command{
		Flags: []Flag{
			&BoolFlag{
				Name:  "interactive",
				Value: true,
			},
		},
	}

	for _, f := range cmd.allFlags() {
		err := f.PreParse()
		require.NoError(t, err)
		err = f.Set("interactive", "true")
		require.NoError(t, err)
	}

	assert.True(t, cmd.shouldRunInteractive())
}

func TestCommand_InteractiveMethod_NotSet(t *testing.T) {
	cmd := &Command{
		Flags: []Flag{
			&StringFlag{
				Name:  "name",
				Value: "test",
			},
		},
	}

	for _, f := range cmd.allFlags() {
		err := f.PreParse()
		require.NoError(t, err)
	}

	assert.False(t, cmd.shouldRunInteractive())
}

func TestInteractiveStringFlag_GetDefaultValue(t *testing.T) {
	flag := &InteractiveStringFlag{
		StringFlag: StringFlag{
			Name:  "test",
			Value: "default_value",
		},
	}

	assert.Equal(t, "default_value", flag.GetDefaultValue())
}

func TestInteractiveIntFlag_GetDefaultValue(t *testing.T) {
	flag := &InteractiveIntFlag{
		Int64Flag: Int64Flag{
			Name:  "test",
			Value: 42,
		},
	}

	assert.Equal(t, "42", flag.GetDefaultValue())
}

func TestInteractiveBoolFlag_GetDefaultValue(t *testing.T) {
	flag := &InteractiveBoolFlag{
		BoolFlag: BoolFlag{
			Name:  "test",
			Value: true,
		},
	}

	assert.Equal(t, "true", flag.GetDefaultValue())
}

func TestInteractiveFloatFlag_GetDefaultValue(t *testing.T) {
	flag := &InteractiveFloatFlag{
		FloatFlag: FloatFlag{
			Name:  "test",
			Value: 3.14,
		},
	}

	assert.Equal(t, "3.14", flag.GetDefaultValue())
}

func TestInteractiveDurationFlag_GetDefaultValue(t *testing.T) {
	flag := &InteractiveDurationFlag{
		DurationFlag: DurationFlag{
			Name:  "test",
			Value: 5 * time.Second,
		},
	}

	assert.Equal(t, "5s", flag.GetDefaultValue())
}
