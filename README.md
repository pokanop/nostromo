[![Travis (.org)](https://img.shields.io/travis/pokanop/nostromo)](https://travis-ci.org/pokanop/nostromo)
[![Coveralls github](https://img.shields.io/coveralls/github/pokanop/nostromo)](https://coveralls.io/github/pokanop/nostromo)
[![GitHub](https://img.shields.io/github/license/pokanop/nostromo)](https://github.com/pokanop/nostromo/blob/master/LICENSE)

# nostromo
`nostromo` is a CLI to manage aliases through simple commands to add and remove scoped aliases and substitutions.

Managing aliases can be tedius and difficult to set up. `nostromo` makes this process easy and reliable. The tool adds shortcuts to your `.bash_profile` that call into the `nostromo` binary. It reads and manages all aliases within its own manifest. This is used to find and execute the actual command as well as swap any substitutions to simplify calls.

`nostromo` can potentially help you build complex tools in a declarative way. Tools commonly allow you to run multi-level commands like `git rebase master branch` or `docker rmi b750fe78269d` which seem clear to use. Imagine if you could wrap your aliases / commands / workflow into custom commands that describe things you do often.

With `nostromo` you can take some aliases like these:
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
test ios --with-some-flag
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

## Key Features
- Simplified alias management
- Scoped commands and substitutions
- Build complex command trees
- Bash completion support
- Preserves flags and arguments

## Usage

### Managing Aliases

### Scoped Commands & Substitutions

### Complex Command Tree

### Bash Completion

## Credits

## Contibuting
Contributions are what make the open source community such an amazing place to be learn, inspire, and create. Any contributions you make are **greatly appreciated**.

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## License
Distributed under the MIT License.