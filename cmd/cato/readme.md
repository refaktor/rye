# cato - cat to

inline "editor" somewhat like cat, but you can traverse the text (all lines) and edit it, not just enter it in one shot.

has some sane emacs like keybindins and could have simple syntax highlighting, for example for rye code or markdown

## Use

```
cato file.txt

cato readme.rye     # uses syntax highlighting for rye

cato readme.md      # uses syntax highlighting for markdown

cato -e readme.md   # edit the file inline

cato -v readme.md   # view the file
```

## Note

this is a side-effect of makign Rye shell accept and work well with multiline inputs. Visit ryelang.org to find out more.
