package cmd

import (
	"bytes"
	"os/exec"
	"strings"

	"github.com/mitchellh/go-homedir"
	"gopkg.in/src-d/go-git.v4"
	config2 "gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"io/ioutil"
	"os"
	"path"
)

type ExRepository struct {
	*git.Repository
}

func (r *ExRepository) ExecuteCommand(arguments ...string) (string, error) {
	var out bytes.Buffer
	cmd := exec.Command("git", arguments...)
	w, _ := r.Worktree()
	cmd.Dir = w.Filesystem.Root()
	cmd.Stdout = &out
	resp := ""
	err := cmd.Run()
	if err == nil {
		resp = strings.TrimSuffix(out.String(), "\n")
	}
	return resp, err
}

func (r *ExRepository) MergeBase(references ...*plumbing.Reference) (*object.Commit, error) {
	args := []string{"merge-base"}
	for _, ref := range references {
		args = append(args, ref.Name().String())
	}
	mergeBase, err := r.ExecuteCommand(args...)
	if err != nil {
		return nil, err
	}
	return r.CommitObject(plumbing.NewHash(mergeBase))
}

func (r *ExRepository) MergeTree(mergeBase *object.Commit, references ...*plumbing.Reference) (string, error) {
	args := []string{"merge-tree", mergeBase.Hash.String()}
	for _, ref := range references {
		args = append(args, ref.Name().String())
	}
	return r.ExecuteCommand(args...)
}

func (r *ExRepository) LocalUserEmail() string {
	config, _ := r.Config()
	email := config.Raw.Section("user").Option("email")
	if email != "" {
		return email
	} else {
		globalConfig := config2.NewConfig()

		home, _ := homedir.Dir()
		globalConfigPath := path.Join(home, ".gitconfig")
		if _, err := os.Stat(globalConfigPath); err == nil {
			content, err := ioutil.ReadFile(globalConfigPath)
			if err == nil {
				globalConfig.Unmarshal(content)
				return globalConfig.Raw.Section("user").Option("email")
			}
		}
	}
	return ""
}

type ConflictResult struct {
	Conflicts int
	Error     error
}

func checkConflict(repo *ExRepository, source *plumbing.Reference, target *plumbing.Reference) (<-chan int, <-chan error) {
	resultChannel := make(chan int)
	errorChannel := make(chan error)
	go func() {
		count := 0
		mergeBase, err := repo.MergeBase(source, target)
		var mergeStatus string
		if err == nil {
			mergeStatus, err = repo.MergeTree(mergeBase, source, target)
			count = strings.Count(mergeStatus, "<<<<<<<")
			if err == nil {
				resultChannel <- count
			}
		}
		errorChannel <- err
	}()
	return resultChannel, errorChannel
}
