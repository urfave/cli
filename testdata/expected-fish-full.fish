# greet fish shell completion

function __fish_greet_no_subcommand --description 'Test if there has been any subcommand yet'
    for i in (commandline -opc)
        if contains -- $i config c sub-config s ss info i in some-command
            return 1
        end
    end
    return 0
end

complete -c greet -n '__fish_greet_no_subcommand' -l socket -s s -r -d 'some \'usage\' text'
complete -c greet -n '__fish_greet_no_subcommand' -f -l flag -s fl -s f -r
complete -c greet -n '__fish_greet_no_subcommand' -f -l another-flag -s b -d 'another usage text'
complete -c greet -n '__fish_greet_no_subcommand' -f -l help -s h -d 'show help'
complete -c greet -n '__fish_greet_no_subcommand' -f -l version -s v -d 'print the version'
complete -c greet -n '__fish_seen_subcommand_from config c' -f -l help -s h -d 'show help'
complete -r -c greet -n '__fish_greet_no_subcommand' -a 'config c' -d 'another usage test'
complete -c greet -n '__fish_seen_subcommand_from config c' -l flag -s fl -s f -r
complete -c greet -n '__fish_seen_subcommand_from config c' -f -l another-flag -s b -d 'another usage text'
complete -c greet -n '__fish_seen_subcommand_from sub-config s ss' -f -l help -s h -d 'show help'
complete -r -c greet -n '__fish_seen_subcommand_from config c' -a 'sub-config s ss' -d 'another usage test'
complete -c greet -n '__fish_seen_subcommand_from sub-config s ss' -f -l sub-flag -s sub-fl -s s -r
complete -c greet -n '__fish_seen_subcommand_from sub-config s ss' -f -l sub-command-flag -s s -d 'some usage text'
complete -c greet -n '__fish_seen_subcommand_from info i in' -f -l help -s h -d 'show help'
complete -r -c greet -n '__fish_greet_no_subcommand' -a 'info i in' -d 'retrieve generic information'
complete -c greet -n '__fish_seen_subcommand_from some-command' -f -l help -s h -d 'show help'
complete -r -c greet -n '__fish_greet_no_subcommand' -a 'some-command'
