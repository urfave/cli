# greet fish shell completion

function __fish_greet_no_subcommand --description 'Test if there has been any subcommand yet'
    for i in (commandline -opc)
        if contains -- $i config c info i in some-command hidden-command usage u
            return 1
        end
    end
    return 0
end

complete -c greet -n '__fish_greet_no_subcommand' -l socket -s s -r -d 'some \'usage\' text'
complete -c greet -n '__fish_greet_no_subcommand' -f -l flag -s fl -s f -r
complete -c greet -n '__fish_greet_no_subcommand' -f -l another-flag -s b -d 'another usage text'
complete -c greet -n '__fish_greet_no_subcommand' -l logfile -r
complete -c greet -n '__fish_greet_no_subcommand' -l foofile -r
complete -x -c greet -n '__fish_greet_no_subcommand' -a 'config c' -d 'another usage test'
complete -c greet -n '__fish_seen_subcommand_from config c' -l flag -s fl -s f -r
complete -c greet -n '__fish_seen_subcommand_from config c' -f -l another-flag -s b -d 'another usage text'
complete -x -c greet -n '__fish_seen_subcommand_from config c; and not __fish_seen_subcommand_from sub-config s ss' -a 'sub-config s ss' -d 'another usage test'
complete -c greet -n '__fish_seen_subcommand_from sub-config s ss' -f -l sub-flag -s sub-fl -s s -r
complete -c greet -n '__fish_seen_subcommand_from sub-config s ss' -f -l sub-command-flag -s s -d 'some usage text'
complete -x -c greet -n '__fish_greet_no_subcommand' -a 'info i in' -d 'retrieve generic information'
complete -x -c greet -n '__fish_greet_no_subcommand' -a 'some-command'
complete -c greet -n '__fish_seen_subcommand_from hidden-command' -f -l completable
complete -x -c greet -n '__fish_greet_no_subcommand' -a 'usage u' -d 'standard usage text'
complete -c greet -n '__fish_seen_subcommand_from usage u' -l flag -s fl -s f -r
complete -c greet -n '__fish_seen_subcommand_from usage u' -f -l another-flag -s b -d 'another usage text'
complete -x -c greet -n '__fish_seen_subcommand_from usage u; and not __fish_seen_subcommand_from sub-usage su' -a 'sub-usage su' -d 'standard usage text'
complete -c greet -n '__fish_seen_subcommand_from sub-usage su' -f -l sub-command-flag -s s -d 'some usage text'
