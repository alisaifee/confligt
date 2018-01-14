package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
)

var docCmd = &cobra.Command{
	Use:    "doc",
	Short:  "Generate README",
	Long:   `Generate README`,
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		if nil != doc.GenMarkdownTree(RootCmd, "/tmp") {
			L.Fatal("unable to generate readme")
		}
		in, _ := ioutil.ReadFile("/tmp/confligt.md")
		// Add installation
		installationString, _ := ioutil.ReadFile("INSTALLATION.md")
		readme := string(in)
		end := strings.Index(readme, RootCmd.Short) + len(RootCmd.Short)
		finalReadme := readme[:end] + "\n" + string(installationString) + readme[end:]
		// Fix sillyness
		re := regexp.MustCompile(`(Number of branches to check concurrently \(default) (\d+)\)`)
		finalReadme = re.ReplaceAllString(finalReadme, `$1 NUMCPUs/2)`)
		file, _ := os.Create("README.md")
		defer file.Close()
		file.Write([]byte(finalReadme))
	},
}

func init() {
	RootCmd.AddCommand(docCmd)
}
