package cli

import "fmt"

// Author represents someone who has contributed to a cli project.
type Author struct {
	Name  string // The Authors name
	Email string // The Authors email
}

// String makes Author comply to the Stringer interface, to allow an easy print in the templating process
func (a *Author) String() string {
	e := ""
	if a.Email != "" {
		e = " <" + a.Email + ">"
	}

	return fmt.Sprintf("%v%v", a.Name, e)
}
