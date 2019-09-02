[![Travis (.org)](https://img.shields.io/travis/pokanop/nostromo)](https://travis-ci.org/pokanop/nostromo)
[![Coveralls github](https://img.shields.io/coveralls/github/pokanop/nostromo)](https://coveralls.io/github/pokanop/nostromo)
[![GitHub](https://img.shields.io/github/license/pokanop/nostromo)](https://github.com/pokanop/nostromo/blob/master/LICENSE)

# nostromo
<img src="images/nostromo-manifest-show.png" alt="nostromo">

nostromo is a CLI to manage aliases through simple commands to add and remove scoped aliases and substitutions.

Managing aliases can be tedious and difficult to set up. nostromo makes this process easy and reliable. The tool adds shortcuts to your `.bashrc` that call into the nostromo binary. It reads and manages all aliases within its manifest. This is used to find and execute the actual command as well as swap any substitutions to simplify calls.

nostromo can potentially help you build complex tools in a declarative way. Tools commonly allow you to run multi-level commands like `git rebase master branch` or `docker rmi b750fe78269d` which seem clear to use. Imagine if you could wrap your aliases / commands / workflow into custom commands that describe things you do often.

With nostromo you can take aliases like these:
```sh
alias ios-build='pushd $IOS_REPO_PATH;xcodebuild -workspace Foo.xcworkspace -scheme foo_scheme'
alias ios-test='pushd $IOS_REPO_PATH;xcodebuild -workspace Foo.xcworkspace -scheme foo_test_scheme'
alias android-build='pushd $ANDROID_REPO_PATH;./gradlew build'
alias android-test='pushd $ANDROID_REPO_PATH;./gradlew test'
```
and turn them into declarative commands like this:
```sh
build ios
build android
test ios
test android
```
The possibilities are endless and up to your imagination with the ability to compose commands as you see fit.

## Getting Started

### Prerequisites
1. A working `go` installation with `GOPATH` and `PATH` set to run installed binaries
2. Works for MacOS and `bash` shell (other combinations untested)

### Installation
```sh
go get -u github.com/pokanop/nostromo
```

### Initialization
This command will initialize nostromo and create a manifest under `~/.nostromo`:
```sh
nostromo manifest init
```

## Key Features
- Simplified alias management
- Scoped commands and substitutions
- Build complex command trees
- Bash completion support
- Preserves flags and arguments

## Usage

### Managing Aliases
Aliases is one of the core functionalities provided by nostromo. Instead of constantly updating shell profiles manually, nostromo will automatically keep it updated with the latest additions. **Note** that commands won't take effect until you open a new shell or `source` your `.bashrc` again.

To add an alias (or command in nostromo parlance), simply run:
```sh
nostromo add cmd foo "echo doo"
```
Re-source your `.bashrc` and just like that you can run `foo` like any other alias.

#### Keypaths
nostromo uses the concept of keypaths to simplify building commands and accessing the command tree. A keypath is simply a `.` delimited string that represents the path to the command.

For example:
```sh
nostromo add cmd foo.bar.baz 'echo hello'
```
will build the command tree for `foo` -> `bar` -> `baz` such that any of these commands are now valid:
```sh
foo
foo bar
foo bar baz
```
where the last one will execute the `echo` command.

This is also how you can compose further commands by adding a **real** command to `foo.bar` for example that will get prepended for commands below that scope.

### Scoped Commands & Substitutions
Scope affects a tree of commands such that a parent scope is executed first and then each command in the keypath. If a command is run as follows:
```sh
foo bar baz
```
then the command associated with `foo` runs first, then `bar`, and finally `baz`. So if these commands were configured like this:
```sh
nostromo add cmd foo 'echo oof'
nostromo add cmd foo.bar 'rab'
nostromo add cmd foo.bar.baz 'zab'
```
then the actual execution would result in:
```sh
oof rab zab
```

nostromo also provides the ability to add substitutions at each one of these scopes in the command tree. So if you want to shorten common strings that are otherwise long into substitutions, you can attach them to a parent scope and nostromo will replace them at execution time for all instances.

A substitution can be added with:
```sh
nostromo add foo.bar sub //some/long/string sls
```
Subsequent calls to `foo bar` would replace the subs before running. This command:
```sh
foo bar baz sls
```
would finally run the following since the substitution is in scope:
```sh
oof rab zab //some/long/string
```

### Complex Command Tree
Given features like **keypaths** and **scope** you can build a complex set of commands and effectively your own tool that performs additive functionality with each command node.

You can get a quick snapshot of the command tree using:
```sh
nostromo manifest show
```

### Bash Completion
nostromo provides bash completion scripts to allow tab completion. You can get this by adding this to your shell init file:
```sh
eval "$(nostromo completion)"
```

## Credits
- This tool was bootstrapped using [cobra](https://github.com/spf13/cobra).
- Colored logging provided by [aurora](https://github.com/logrusorgru/aurora).

## Contibuting
Contributions are what makes the open-source community such an amazing place to learn, inspire, and create. Any contributions you make are **greatly appreciated**.

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## License
Distributed under the MIT License.