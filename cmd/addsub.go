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

// addsubCmd represents the addsub command
var addsubCmd = &cobra.Command{
	Use:   "sub [key.path] [name] [alias]",
	Short: "Add a substitution to nostromo manifest",
	Long: `Add a substitution to nostromo manifest for a given key path and arg.
A substitution allows any arguments as part of a command to be substituted
by that value, e.g., "original-cmd //some/long/arg1 //some/long/arg2" can
be substituted out with "cmd-alias sub1 sub2" by adding subs for the args.

This will create the substitution for scopes beneath levels in
the provided key path. A command scope can a tree of sub commands
and substitutions.`,
	Args: cobra.MinimumNArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		task.AddSubstitution(args[0], args[1], args[2])
	},
}

func init() {
	addCmd.AddCommand(addsubCmd)
}
