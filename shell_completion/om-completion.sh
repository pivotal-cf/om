#!/usr/bin/env bash

_om_completions() {
  COMPREPLY=($(compgen -W "$(om 2>/dev/null | grep -e '^[[:space:]][[:space:]]*[[:alnum:]]' | sed -e 's/^[[:space:]]*//' | awk '{print $1}')" -- "${COMP_WORDS[1]}"))
}

complete -F _om_completions om
