package cli

type Command struct {
	Name        string
	ShortName   string
	Usage       string
	Description string
	Action      Handler
	Flags       []Flag
}
