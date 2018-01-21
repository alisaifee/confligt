## confligt

Find conflicting branches in git repositories
#### Installation

If you have a working golang development environment: `go get github.com/alisaifee/confligt`

If you are using a mac & use [homebrew](https://brew.sh/): `brew tap alisaifee/homebrew-tap && brew install confligt`

If you are on a debian based OS, grab a `.deb` from [here](https://github.com/alisaifee/confligt/releases/latest)


### Synopsis

Confligt finds conflicting branches in git repositories.

Without any arguments or flags, confligt will inspect all local & remote branches in the current working
directory - that have commits since 7 days ago - against each other and other remote branches
(from the default origin) to find conflicting pairs.

```
confligt [flags]
```

### Examples

```

# Filter by branches that were updated a day ago
$ confligt --since='1 day'

# Filter by branches that start with foo- or bar-
$ confligt --filter='\b(foo|bar)-'

# Inspect branches in the remote named alice. Use develop as the default branch.
$ confligt --remote=alice --main=develop
	
```

### Options

```
      --concurrency int   Number of branches to check concurrently (default NUMCPUs/2)
      --config string     config file (default is $HOME/.confligt.yaml)
      --fetch             Fetch from remote before inspecting
      --filter string     Regular expression to match branch names against
  -h, --help              help for confligt
      --include-local     Include local branches when finding conflicts (default true)
      --include-remote    Include remote branches when finding conflicts (default true)
  -m, --main string       Name of main branch (default "master")
      --mine              Inspect only your own branches
  -r, --remote string     Name of remote (default "origin")
  -s, --since string      Consider branches with commits since (default "7 days")
  -v, --verbose           Display verbose logging
```

