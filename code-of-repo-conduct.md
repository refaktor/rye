# Code of repository conduct

With recent HN post a lot of new people came to follow development of Rye. We probably all want Rye updates and general state of Rye repository to 
be "a good product" to follow too. 

Here we try to come up with some directives, we could follow to achieve this. Everybody is are welcome to propose changes.

We shouldn't have too many principles (or rules), we are not trying to create a bureaucracy here. Just the really necessary ones, the rest is advice or knowledge base ... 

## Principles we really try to follow 

1. Main branch shouldn't be left in a failed state (the red cross)
2. PRs that are meant for merging should also pass all checks
3. Each new builtin should have a docstring written and at least few tests

## Q&A

### How to create commits / PR-s / merge them so that feed and release notes are as informative as possible?

_work-in-progress-text-propose-changes_

If there is only one commit per PR then the message is stored and used (check this).

If there are multiple commits's then messages get lost unless we merge with *Squash and merge*. This joins messages together and uses them (positive) but also joins all code changes into one commit. 
A) If reason for multiple commits is iterating on a set of changes, or making them comple (adding tests, making golint-ci pass, ...)  then this makes sense. 
B) If PR is composed of multiple commits each for different set of changes then some information is lost with Squash.

* Should the choice be per-case based?
* What to do in case (B) then?
* Is there a third option?
* Should in case of (B) these be multiple PR-s anyway?



