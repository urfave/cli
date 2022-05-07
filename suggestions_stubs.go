//go:build urfave_cli_no_suggest || urfave_cli_core
// +build urfave_cli_no_suggest urfave_cli_core

package cli

func (a *App) suggestFlagFromError(err error, _ string) (string, error) {
	return "", err
}

func suggestCommand([]*Command, string) string {
	return ""
}
