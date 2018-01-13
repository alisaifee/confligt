## confligt

Find conflicting branches

#### Installation
`go get github.com/alisaifee/confligt`

### Synopsis


Confligt finds conflicting branches in your git repository.

Without any arguments or flags, confligt will inspect all local branches in the current working
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
      --local-only        Find conflicts for local branches only (default true)
  -m, --main string       Name of main branch (default "master")
      --mine              Inspect only your own branches
  -r, --remote string     Name of remote (default "origin")
  -s, --since string      Consider branches with commits since (default "7 days")
  -v, --verbose           Display verbose logging
```

