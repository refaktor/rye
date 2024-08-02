# Contributing

When trying to contribute try to follow directives below, to make the progress of the project streamlined and Rye up to certain quality.

## Before opening a PR
* Open and issue so we can make a discussion first

## When making a PR
* PRs that are meant for merging should pass all checks
* Each new builtin should have a docstring written and at least few tests

## In general
* Main branch shouldn't be left in a failed state (the red cross)

## Q&A

### How to create commits / PR-s / merge them so that feed and release notes are as informative as possible?

If there is only one commit per PR then the message is stored and used.

If there are multiple commits's then messages get lost unless we merge with *Squash and merge*. This joins messages together and uses them (positive) but also joins all code changes into one commit. 
A) If reason for multiple commits is iterating on a set of changes, or making them comple (adding tests, making golint-ci pass, ...)  then this makes sense. 
B) If PR is composed of multiple commits each for different set of changes then some information is lost with Squash, maybe these should be multiple PR-a.

_This is work in progress, you can propose improvements._
