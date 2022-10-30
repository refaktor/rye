# Idea for distribution of rye

## The start

It all starts with tinyrye, a binary that you can download locally or setup to global path yourself or copy to /usr/local/bin/. 

`tinyrye install` downloads the signed install script and runs it. Tinyrye is just rye build with b_tiny.

## The instalation

Install script is interactive script, it will create the ~/.rye directory. Into it it will copy the tinyrye binary, so it's doesn't rely of being on path.

It will ask user for some basic info to create the ~/.rye/.profile .

It will ask user if it should make tinyrye globaly accessible. If yet it will make a link to tinyrye in /usr/local/bin.

It will create the rye and rye-assist files in /usr/local/bin.

It will download the script assist.rye for rye-assist.

It will install golang and needed modules for tinyrye.

You can then use tinyrye for general use around the shell. Tinyrye should have some basic builtins so that it is usable for this.

## Use in projects

When inside a project folder, you create rye.mod file where you list modules you want local rye to have.

Then you use `rye-assist build` to build a local rye with those modules. It should also do go-get commands later, in first version you need to make them.

When adding modules you add them to rye.mod and rebuild.

## Rye assist

`rye-assist`

Rye assist is a tool for local use of rye. If you call it, it should greet you and explain what it is, and basic usage.

`rye-assist help`

Should display it's basic info and all options it has.

`rye-assist build`

Should build a local rye binary with rye.mod modules.

`rye-assist modules`

Should list modules for local rye build.

`rye-assist modules add xxx`

Should add module to local rye.mod.

### Question

Maybe rye-assist should also be used to install rye?

