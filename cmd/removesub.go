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

// removesubCmd represents the removesub command
var removesubCmd = &cobra.Command{
	Use:   "sub [key.path] [alias]",
	Short: "Remove a substitution from nostromo manifest",
	Long: `Remove a substitution from nostromo manifest for a given key path and arg.
A substitution allows any arguments as part of a command to be substituted
by that value, e.g., "original-cmd //some/long/arg1 //some/long/arg2" can
be substituted out with "cmd-alias sub1 sub2" by adding subs for the args.

This will remove the substitution for scopes beneath levels in
the provided key path. A command scope can a tree of sub commands
and substitutions.`,
	Run: func(cmd *cobra.Command, args []string) {
		if !minArgs(2, cmd, args) {
			return
		}
		task.RemoveSubstitution(args[0], args[1])
	},
}

func init() {
	removeCmd.AddCommand(removesubCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// removesubCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// removesubCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
