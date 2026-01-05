# Bin folder

## l.rye

"local Rye" idea. This Rye script builds a local Rye interpreter with specified modules based on the bindings listed in lrye.mod file.

### Usage

Add lrye.mod and list modules if needed (build tags), then run:

`rye l.rye [command] [options]`

Since it uses a shebang line on Linux you can just use:

`./l.rye`

Environment Variables Required:

RYE_HOME       - Path to the Rye source directory

Instructions on how to setup this

`./l.rye setup`


### Building local (project specific) rye

add build tags to the lrye.mod (space separated or one per line). Run `$RYE_HOME/l.rye build`.

Builds a local lrye with modules (build tags) that you specified. This means you can exclude certain modules or add additional for specific project.

You can invoke ./lrye like any local executable, or with less typing just by calling `lrye` (requires $RYE_HOME/bin in PATH) script.


### Building executable with embeded main.rye

Example:

look at rye/examples/hello-world-build/

`$RYE_HOME/bin/l.rye build embed_main`

### More info

Run for CLI options:

`$RYE_HOME/bin/l.rye`

Run for information about setup:

`$RYE_HOME/bin/l.rye setup`

### Current limitations:

* Work in progress
* Proof of concept only
* Only embeds main.rye
* Only tested on Linux

Was already used to create a fyne+rye binary and also .apk (Android Fyne app).


## Optional PATH

If you setup PATH to this `bin` folder you get many benefits. For example in your $HOME/.profile file add:

```
export PATH=$PATH:$HOME/Work/rye/bin
```

Then you can run `l.rye` directly from anywhere. And you also get access to `lrye` script which runs your local ./lrye without the need to specify "./" part.
