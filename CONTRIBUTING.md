# Contributing

## Before PR-s
* Open and issue so we can make a discussion first

## Making PR-s
* Main branch shouldn't be left in a failed state (the red cross)
* PRs that are meant for merging should also pass all checks
* Each new builtin should have a docstring written and at least few tests

## Q&A

### How to create commits / PR-s / merge them so that feed and release notes are as informative as possible?

If there is only one commit per PR then the message is stored and used.

If there are multiple commits's then messages get lost unless we merge with *Squash and merge*. This joins messages together and uses them (positive) but also joins all code changes into one commit. 
A) If reason for multiple commits is iterating on a set of changes, or making them comple (adding tests, making golint-ci pass, ...)  then this makes sense. 
B) If PR is composed of multiple commits each for different set of changes then some information is lost with Squash, maybe these should be multiple PR-a.

_This is work in progress, you can propose better system_
