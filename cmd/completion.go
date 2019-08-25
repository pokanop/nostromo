/*
Copyright © 2019 Sahel Jalal <sahel.jalal@icloud.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"fmt"

	"github.com/pokanop/nostromo/pathutil"
	"github.com/spf13/cobra"
)

// completionCmd represents the completion command
var completionCmd = &cobra.Command{
	Use:   "completion",
	Short: "Generates bash completion scripts",
	Long: `To load completion now, run

. ~/.nostromo/completion

To configure your bash shell to load completions for each session add to your shell init files

# In ~/.bashrc or ~/.profile
. ~/.nostromo/completion`,
	Run: func(cmd *cobra.Command, args []string) {
		err := rootCmd.GenBashCompletionFile(pathutil.Abs("~/.nostromo/completion"))
		if err != nil {
			fmt.Println(err)
			printUsage(cmd)
			return
		}
		fmt.Println(cmd.Long)
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)
}
