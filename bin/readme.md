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


### Adding project-specific builtins

You can compile extra Go builtins into your local rye by adding the `proj_builtins` marker to `lrye.mod` and dropping Go files into a `project_builtins/` directory next to it.

1. Add `proj_builtins` to your `lrye.mod` (it is not prefixed with `b_` like regular modules):

   ```
   tiny http proj_builtins
   ```

2. Create `project_builtins/` and put one or more `.go` files in it. Each file must be `package projectbuiltins` and define a single `RegisterProjectBuiltins` function. Register your builtins with `evaldo.RegisterBuiltins2` (flat) and/or `evaldo.RegisterBuiltinsInContext` (context/dialect):

   ```go
   // project_builtins/mybuiltins.go
   package projectbuiltins

   import (
       "github.com/refaktor/rye/env"
       "github.com/refaktor/rye/evaldo"
   )

   func RegisterProjectBuiltins(ps *env.ProgramState) {
       evaldo.RegisterBuiltins2(map[string]*env.Builtin{
           "my-hello": {
               Argsn: 0,
               Fn: func(ps *env.ProgramState, arg0, arg1, arg2, arg3, arg4 env.Object) env.Object {
                   return *env.NewString("hello from project builtins")
               },
           },
       }, ps, "project")
   }
   ```

3. Build as usual. l.rye copies your files into the Rye source and builds with the `proj_builtins` tag, registering them after the standard builtins:

   ```
   l.rye build
   ```

If `proj_builtins` is missing from `lrye.mod`, the `project_builtins/` directory is ignored. If the marker is present but the directory has no `.go` files, l.rye warns and builds normally.


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
