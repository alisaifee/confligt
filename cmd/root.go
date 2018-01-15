package cmd

import (
	"fmt"
	"os"

	"path"
	"regexp"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/mitchellh/go-homedir"
	"github.com/remeh/sizedwaitgroup"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/urjitbhatia/gohumantime"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"math"
	"runtime"
)

var cfgFile string

var RootCmd = &cobra.Command{
	Use:   "confligt",
	Short: "Find conflicting branches in git repositories",
	Long: `Confligt finds conflicting branches in git repositories.

Without any arguments or flags, confligt will inspect all local branches in the current working
directory - that have commits since 7 days ago - against each other and other remote branches
(from the default origin) to find conflicting pairs.`,
	Example: `
# Filter by branches that were updated a day ago
$ confligt --since='1 day'

# Filter by branches that start with foo- or bar-
$ confligt --filter='\b(foo|bar)-'

# Inspect branches in the remote named alice. Use develop as the default branch.
$ confligt --remote=alice --main=develop
	`,
	DisableAutoGenTag: true,
	Args:              cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var secondsSince float64
		if millisSince, err := gohumantime.ToMilliseconds(viper.GetString("since")); err == nil && millisSince != 0 {
			secondsSince = float64(millisSince) / 1000
		} else {
			L.Fatalf("Unable to parse: %s", viper.GetString("since"))
		}
		var repoPath string
		if len(args) > 0 {
			repoPath = args[0]
		} else {
			repoPath = "."
		}
		refBranchName := path.Join("refs/remotes", viper.GetString("remote"), viper.GetString("main"))
		remoteName := viper.GetString("remote")
		bare, err := git.PlainOpen(repoPath)
		repository := &ExRepository{bare}
		if err != nil {
			L.Fatalf("Unable to open git repository at %s", repoPath)
		}

		var currentUserEmail string
		if viper.GetBool("mine") {
			currentUserEmail = repository.LocalUserEmail()
			if currentUserEmail == "" {
				L.Fatal("Unable to infer current user")
			}
		}
		if viper.GetBool("fetch") {
			if V {
				L.Printf("Fetching from remote %s...", remoteName)
			}
			_, err := repository.ExecuteCommand("fetch")
			if err != nil {
				L.Fatalf("Error fetching from %s", remoteName)
			}
		}

		var mainBranch *plumbing.Reference

		conflicts := 0
		references, err := repository.References()
		remoteBranches := make(map[string]*plumbing.Reference)
		localBranches := make(map[string]*plumbing.Reference)
		// filter out branches that don't match search criteria
		references.ForEach(func(reference *plumbing.Reference) error {
			if reference.Name().String() == refBranchName {
				mainBranch = reference
			} else {
				if commit, err := repository.CommitObject(reference.Hash()); err == nil && time.Now().Sub(commit.Author.When).Seconds() < secondsSince {
					if viper.GetBool("mine") && commit.Author.Email != currentUserEmail {
						return nil
					}
					if viper.GetString("filter") != "" {
						match, _ := regexp.MatchString(viper.GetString("filter"), reference.Name().Short())
						if !match {
							return nil
						}
					}
					if reference.Name().IsBranch() {
						localBranches[reference.Name().String()] = reference
					} else {
						if strings.HasPrefix(reference.Name().String(), "refs/remotes/"+remoteName) &&
							!strings.Contains(reference.Name().String(), "HEAD") {
							remoteBranches[reference.Name().String()] = reference
						}
					}
				}
			}
			return nil
		})
		var branchesToCheck map[string]*plumbing.Reference
		if mainBranch == nil {
			L.Fatalf("Unable to find main branch with name %s", refBranchName)
		} else {

			if viper.GetBool("local-only") {
				branchesToCheck = localBranches
			} else {
				branchesToCheck = localBranches
				for k, v := range remoteBranches {
					branchesToCheck[k] = v
				}
			}
			// filter out branches that have already been merged to mainBranch
			for name, branch := range branchesToCheck {
				commit, _ := repository.MergeBase(branch, mainBranch)
				if commit.Hash == branch.Hash() {
					if V && !strings.Contains(mainBranch.Name().Short(), branch.Name().Short()) {
						L.Printf(
							"%s is already merged. Consider deleting the branch",
							yellow(branch.Name().Short()),
						)
					}
					delete(branchesToCheck, name)
				}
			}
		}

		if V {
			L.Printf("Inspecting %d branch(es)...", len(branchesToCheck))
		}
		conflictingWithMaster := make([]*plumbing.Reference, 0)
		rebasedWithMaster := make([]*plumbing.Reference, 0)
		wg := sizedwaitgroup.New(viper.GetInt("concurrency"))

		// skip branches that already conflict with master
		for _, reference := range branchesToCheck {
			wg.Add()
			go func(source *plumbing.Reference, target *plumbing.Reference) {
				resultC, errC := checkConflict(repository, source, target)
				select {
				case _ = <-errC:
					wg.Done()
				case res := <-resultC:
					if res > 0 {
						fmt.Printf(
							"%v conflicts with %s [%d conflict(s)]\n",
							boolColor(target.Name().Short(), res == 1, color.FgYellow, color.FgRed),
							cyan(source.Name().Short()),
							res,
						)
						conflicts = +1
						conflictingWithMaster = append(conflictingWithMaster, target)
					} else {
						rebasedWithMaster = append(rebasedWithMaster, target)
					}
					wg.Done()

				}
			}(mainBranch, reference)
		}
		wg.Wait()
		if V {
			masterConflicts := len(conflictingWithMaster)
			L.Printf(
				"Found %v branch(es) conflicting with %v",
				boolColor(masterConflicts, masterConflicts == 0),
				mainBranch.Name().Short(),
			)
			L.Printf("Inspecting %d branch combinations...", (len(rebasedWithMaster)*(len(rebasedWithMaster)-1))/2)
		}

		// find sad remote branches.
		sadBranches := make(map[string]byte)

		for i, source := range rebasedWithMaster {
			for _, target := range rebasedWithMaster[1+i:] {
				wg.Add()
				go func(source *plumbing.Reference, target *plumbing.Reference) {
					resultC, errC := checkConflict(repository, source, target)
					select {
					case _ = <-errC:
						wg.Done()
					case res := <-resultC:
						if res > 0 {
							L.Printf(
								"%v conflicts with %v [%d conflict(s)]\n",
								boolColor(source.Name().Short(), res == 1, color.FgYellow, color.FgRed),
								boolColor(target.Name().Short(), res == 1, color.FgYellow, color.FgRed),
								res,
							)
							sadBranches[source.Name().String()] = 1
							sadBranches[target.Name().String()] = 1
							conflicts = +1
						}
						wg.Done()
					}
				}(source, target)
			}
		}
		wg.Wait()
		if V {
			if conflicts == 0 {
				L.Println(green("No conflicting branches found"))
			} else {
				L.Printf("Found %s branch(es) conflicting with each other", boolColor(len(sadBranches), len(sadBranches) == 0))

			}
		}

	},
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.confligt.yaml)")
	RootCmd.PersistentFlags().StringP("remote", "r", "origin", "Name of remote")
	RootCmd.PersistentFlags().StringP("main", "m", "master", "Name of main branch")
	RootCmd.PersistentFlags().StringP("since", "s", "7 days", "Consider branches with commits since")
	RootCmd.PersistentFlags().StringP("filter", "", "", "Regular expression to match branch names against")
	RootCmd.PersistentFlags().BoolP("verbose", "v", false, "Display verbose logging")
	RootCmd.PersistentFlags().BoolP("fetch", "", false, "Fetch from remote before inspecting")
	RootCmd.PersistentFlags().BoolP("mine", "", false, "Inspect only your own branches")
	RootCmd.PersistentFlags().BoolP("local-only", "", true, "Find conflicts for local branches only")
	RootCmd.PersistentFlags().IntP("concurrency", "", int(math.Max(1, float64(runtime.NumCPU()/2))), "Number of branches to check concurrently")

	for _, flag := range []string{"remote", "main", "since", "filter", "fetch", "verbose", "mine", "local-only", "concurrency"} {
		viper.BindPFlag(flag, RootCmd.PersistentFlags().Lookup(flag))
	}

}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		viper.AddConfigPath(home)
		viper.SetConfigName(".confligt")
	}

	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err == nil && viper.GetBool("verbose") {
		fmt.Println(yellow("Using config file:"), viper.ConfigFileUsed())
	}
	V = viper.GetBool("verbose")
}
