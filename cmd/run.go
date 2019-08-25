// Copyright © 2019 Sahel Jalal <sahel.jalal@icloud.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"github.com/pokanop/nostromo/task"
	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run [command] [args]",
	Short: "Run a command from manifest",
	Long: `Run a command from manifest.
After adding commands you can run them through nostromo. As long as
a command can be found in the manifest it will be executed.

If you create key paths to commands like this:
  nostromo add cmd foo.bar "./crazy-long-command with-args"
  nostromo add cmd foo.baz "./another-crazy-long-command with-args"
  
It will create command entries at each level. Now you can run the following
it will execute the actual commands:
  nostromo run foo bar
  nostromo run foo baz
  
Since nostromo will alias these commands in your shell profile you can omit
the "nostromo run" command itself. For example, it will add:
  alias foo='nostromo run foo "$*"'

The power of these commands are that you can take complicated combinations
of commands and build up a more intelligent way to run things as if you wrote
your own tool. Imagine composing commands to simplify a workflow:
  build thing1
  build thing2
  
The root "build" command can do things like cd to a folder, set env vars, and
run the main command. Lastly, substitutions can further shorten any sets of
commands that need to be run across the scope of the command.`,
	Args:               cobra.MinimumNArgs(1),
	DisableFlagParsing: true,
	Run: func(cmd *cobra.Command, args []string) {
		task.Run(args)
	},
}

func init() {
	RootCmd.AddCommand(runCmd)
}
