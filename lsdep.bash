#!/usr/bin/env bash
# Usage: lsdep [PACKAGE...]
#
# Example (list github.com/foo/bar and package dir deps [the . argument])
# $ lsdep github.com/foo/bar .
#
# By default, this will list dependencies (imports), test imports, and test
# dependencies (imports made by test imports).  You can recurse further by
# setting TESTIMPORTS to an integer greater than one, or to skip test
# dependencies, set TESTIMPORTS to 0 or a negative integer.

: "${TESTIMPORTS:=1}"

lsdep_impl__ () {
    local txtestimps='{{range $v := .TestImports}}{{print . "\n"}}{{end}}'
    local txdeps='{{range $v := .Deps}}{{print . "\n"}}{{end}}'

    {
        go list -f "${txtestimps}${txdeps}" "$@"
        if [[ -n "${TESTIMPORTS}" ]] && [[ "${TESTIMPORTS:-1}" -gt 0 ]]
        then
            go list -f "${txtestimps}" "$@" |
            sort | uniq |
            comm -23 - <(go list std | sort) |
                TESTIMPORTS=$((TESTIMPORTS - 1)) xargs bash -c 'lsdep_impl__ "$@"' "$0"
        fi
    } |
    sort | uniq |
    comm -23 - <(go list std | sort)
}
export -f lsdep_impl__

lsdep_impl__ "$@"
