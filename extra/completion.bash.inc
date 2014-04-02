#!/bin/bash


# you will need to add the following line to your profile
# source /path/to/completion.bash.inc

_cx_commands()
{
    cx help commands|cut -f 2 -d ' '
}

_cx()
{
    cur=${COMP_WORDS[COMP_CWORD]}
    prev=${COMP_WORDS[COMP_CWORD-1]}
    if [ $COMP_CWORD -eq 1 ]; then
        COMPREPLY=( $( compgen -W "$(_cx_commands)" $cur ) )
    elif [ $COMP_CWORD -eq 2 ]; then
        case "$prev" in
        help)
            COMPREPLY=( $( compgen -W "$(_cx_commands)" $cur ) )
            ;;
        esac
    fi
}

complete -F _cx -o default cx
