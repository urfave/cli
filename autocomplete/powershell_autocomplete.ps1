$fn = $($MyInvocation.MyCommand.Name)
$name = $fn -replace "(.*)\.ps1$", '$1'
Register-ArgumentCompleter -Native -CommandName $name -ScriptBlock {
    param($wordToComplete, $commandAst, $cursorPosition)

    # One level of quoting is taken off each word, the way the shell would before
    # handing it to a command. The value is not expanded: nothing on the command line
    # is evaluated to answer a completion, so a "$(...)" reaches the command as the
    # text it is rather than being run by pressing the tab key.
    function __cliCompletionText($element) {
        if ($element -is [System.Management.Automation.Language.StringConstantExpressionAst] -or
            $element -is [System.Management.Automation.Language.ExpandableStringExpressionAst]) {
            return $element.Value
        }
        return $element.Extent.Text
    }

    $elements = $commandAst.CommandElements
    if ($elements.Count -eq 0) {
        return
    }

    # The command name itself is the shell's to complete, not the command's.
    if ($cursorPosition -le $elements[0].Extent.EndOffset) {
        return
    }

    # The request names the completion in its first argument, where a "--" typed on
    # the command line cannot turn it into a positional argument of whatever the
    # command runs. The word under the cursor is sent as the last argument, empty or
    # not, so that "cmd --<TAB>" and "cmd -- <TAB>" can be told apart.
    #
    # Which word that is comes from the cursor rather than from a comparison with
    # $wordToComplete, which PowerShell hands over normalized: an unfinished "hello
    # arrives here as "hello", matches no element as written, and would be sent both
    # as a word of its own and as the word being completed. Reading the cursor also
    # leaves out what follows it, so completing in the middle of a line asks about
    # the line up to that point.
    $command = __cliCompletionText $elements[0]
    $words = @()
    $word = ''
    for ($i = 1; $i -lt $elements.Count; $i++) {
        $extent = $elements[$i].Extent
        if ($cursorPosition -gt $extent.StartOffset -and $cursorPosition -le $extent.EndOffset) {
            $word = __cliCompletionText $elements[$i]
        } elseif ($extent.EndOffset -lt $cursorPosition) {
            $words += __cliCompletionText $elements[$i]
        }
    }

    & $command __complete @words $word 2>$null | ForEach-Object {
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
