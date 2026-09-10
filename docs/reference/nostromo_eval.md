---
title: eval
---

# nostromo eval

Show eval command from manifest

## Synopsis

Show eval command from manifest.
After adding commands you can run them through nostromo. As long as
a command can be found in the manifest it will provide a command to eval.

If you create key paths to commands like this:
  nostromo add cmd foo.bar "./crazy-long-command with-args"
  nostromo add cmd foo.baz "./another-crazy-long-command with-args"
  
It will create command entries at each level. Nostromo will alias these 
commands in your shell profile to be evaluated. 
For example, it will add:
  alias foo=eval 'nostromo eval foo "$*"'

The power of these commands are that you can take complicated combinations
of commands and build up a more intelligent way to run things as if you wrote
your own tool. Imagine composing commands to simplify a workflow:
  build thing1
  build thing2

The root "build" command can do things like cd to a folder, set env vars, and
run the main command. Lastly, substitutions can further shorten any sets of
commands that need to be run across the scope of the command.

```
nostromo eval [command] [args] [flags]
```

## Options

```
  -h, --help   help for eval
```

## Options inherited from parent commands

```
  -v, --verbose   show verbose logging
```

## SEE ALSO

* [nostromo](nostromo.md)	 - nostromo is a tool to manage aliases

