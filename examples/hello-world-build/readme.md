# Hello world build

This folder is to demonstrate how you can build a single binary executable from rye script. 

You need to have Rye source on your computer and RYE_HOME environmental variable defined, and you need Go(lang) installed.

The whole system is still in a proof of concept stage. It was tested on Linux.

1. Setup

set RYE_HOME=/path/to/rye/source/

Use setup command for more info 

`$RYE_HOME/bin/l.rye setup`

2. Build a binary for main.rye

`$RYE_HOME/bin/l.rye build embed_main`

in the folder a binary will be added. If you run it the Rye script should evaluate. You can move binary or remove main.rye and it will still work.
