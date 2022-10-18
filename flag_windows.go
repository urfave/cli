package cli

func withEnvHint(envVars []string, str string) string {
	// if we are running is powershell this env var is set
	// and so we should use the default env format
	if os.Getenv("PSHOME") != "" {
		envText = defaultEnvFormat(envVars)
	} else {
		envText = envFormat(envVars, "%", "%, %", "%")
	}
	return str + envText
}
