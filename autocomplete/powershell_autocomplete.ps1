$fn = $($MyInvocation.MyCommand.Name)
$name = $fn -replace "(.*)\.ps1$", '$1'
Register-ArgumentCompleter -Native -CommandName $name -ScriptBlock {
    param($wordToComplete, $commandAst, $cursorPosition)

    # The request names the completion in its first argument, where a "--" typed on
    # the command line cannot turn it into a positional argument of whatever the
    # command runs. The word under the cursor is sent as the last argument, empty or
    # not, so that "cmd --<TAB>" and "cmd -- <TAB>" can be told apart.
    $elements = @($commandAst.CommandElements | ForEach-Object { $_.ToString() })
    $command = $elements[0]
    $words = @()
    if ($elements.Count -gt 1) {
        $words = $elements[1..($elements.Count - 1)]
    }
    # Once the word under the cursor has any character it is an element of its own,
    # so it is dropped here and sent as the last argument instead.
    if ($wordToComplete -and $words.Count -gt 0 -and $words[-1] -eq $wordToComplete) {
        if ($words.Count -eq 1) {
            $words = @()
        } else {
            $words = $words[0..($words.Count - 2)]
        }
    }

    & $command __complete @words $wordToComplete 2>$null | ForEach-Object {
        $parts = $_.Split(':', 2)
        if ($parts.Count -eq 2) {
            $completion = $parts[0].Trim()
            $description = $parts[1].Trim()
            [System.Management.Automation.CompletionResult]::new($completion, $completion, 'ParameterValue', $description)
        } else {
            [System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterValue', $_)
        }
    }
}
