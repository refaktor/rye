//go:build !no_git
// +build !no_git

package evaldo

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/refaktor/rye/env"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

// fileInfo represents a file with its path and modification time
type fileInfo struct {
	path  string
	mtime time.Time
}

var Builtins_git = map[string]*env.Builtin{

	//
	// ##### Git ##### "Git repository functions"
	//
	// Tests:
	// equal { git-repo//open "." |type? } 'native
	// Args:
	// * path: path to Git repository
	// Returns:
	// * native Git repository object
	"init": {
		Argsn: 1,
		Doc:   "Initializes a new Git repository at the specified path.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch path := arg0.(type) {
			case env.String:
				repo, err := git.PlainInit(path.Value, false)
				if err != nil {
					return MakeBuiltinError(ps, fmt.Sprintf("failed to initialize repository: %v", err), "git-init")
				}
				return *env.NewNative(ps.Idx, repo, "git-repo")
			case env.Uri:
				repo, err := git.PlainInit(path.GetPath(), false)
				if err != nil {
					return MakeBuiltinError(ps, fmt.Sprintf("failed to initialize repository: %v", err), "git-init")
				}
				return *env.NewNative(ps.Idx, repo, "git-repo")
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType, env.UriType}, "git-init")
			}
		},
	},

	"open": {
		Argsn: 1,
		Doc:   "Opens a Git repository at the specified path.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch path := arg0.(type) {
			case env.String:
				repo, err := git.PlainOpen(path.Value)
				if err != nil {
					return MakeBuiltinError(ps, fmt.Sprintf("failed to open repository: %v", err), "git-repo//open")
				}
				return *env.NewNative(ps.Idx, repo, "git-repo")
			case env.Uri:
				repo, err := git.PlainOpen(path.GetPath())
				if err != nil {
					return MakeBuiltinError(ps, fmt.Sprintf("failed to open repository: %v", err), "git-repo//open")
				}
				return *env.NewNative(ps.Idx, repo, "git-repo")
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType, env.UriType}, "git-repo//open")
			}
		},
	},

	// Tests:
	// equal { repo: git-repo//open "." , repo |git-repo//clone "https://github.com/example/repo.git" "target-dir" |type? } 'native
	// Args:
	// * url: URL of the Git repository to clone
	// * path: path where to clone the repository
	// Returns:
	// * native Git repository object
	"clone": {
		Argsn: 2,
		Doc:   "Clones a Git repository from the specified URL to the specified path.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch url := arg0.(type) {
			case env.String:
				switch path := arg1.(type) {
				case env.String:
					repo, err := git.PlainClone(path.Value, false, &git.CloneOptions{
						URL: url.Value,
					})
					if err != nil {
						return MakeBuiltinError(ps, fmt.Sprintf("failed to clone repository: %v", err), "git-repo//clone")
					}
					return *env.NewNative(ps.Idx, repo, "git-repo")
				case env.Uri:
					repo, err := git.PlainClone(path.GetPath(), false, &git.CloneOptions{
						URL: url.Value,
					})
					if err != nil {
						return MakeBuiltinError(ps, fmt.Sprintf("failed to clone repository: %v", err), "git-repo//clone")
					}
					return *env.NewNative(ps.Idx, repo, "git-repo")
				default:
					return MakeArgError(ps, 2, []env.Type{env.StringType, env.UriType}, "git-repo//clone")
				}
			case env.Uri:
				switch path := arg1.(type) {
				case env.String:
					repo, err := git.PlainClone(path.Value, false, &git.CloneOptions{
						URL: url.GetPath(),
					})
					if err != nil {
						return MakeBuiltinError(ps, fmt.Sprintf("failed to clone repository: %v", err), "git-repo//clone")
					}
					return *env.NewNative(ps.Idx, repo, "git-repo")
				case env.Uri:
					repo, err := git.PlainClone(path.GetPath(), false, &git.CloneOptions{
						URL: url.GetPath(),
					})
					if err != nil {
						return MakeBuiltinError(ps, fmt.Sprintf("failed to clone repository: %v", err), "git-repo//clone")
					}
					return *env.NewNative(ps.Idx, repo, "git-repo")
				default:
					return MakeArgError(ps, 2, []env.Type{env.StringType, env.UriType}, "git-repo//clone")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType, env.UriType}, "git-repo//clone")
			}
		},
	},

	// Tests:
	// equal { repo: git-repo//open "." , repo |git-repo//worktree |type? } 'native
	// Args:
	// * repo: Git repository object
	// Returns:
	// * native Git worktree object
	"git-repo//worktree?": {
		Argsn: 1,
		Doc:   "Gets the worktree for a Git repository.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch repo := arg0.(type) {
			case env.Native:
				if ps.Idx.GetWord(repo.Kind.Index) != "git-repo" {
					return MakeBuiltinError(ps, "expected a Git repository object", "git-repo//worktree?")
				}

				worktree, err := repo.Value.(*git.Repository).Worktree()
				if err != nil {
					return MakeBuiltinError(ps, fmt.Sprintf("failed to get worktree: %v", err), "git-repo//worktree?")
				}
				return *env.NewNative(ps.Idx, worktree, "git-worktree")
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "git-repo//worktree?")
			}
		},
	},

	// Tests:
	// equal { repo: git-repo//open "." , wt: repo |git-repo//worktree , wt |git-worktree//status |type? } 'dict
	// Args:
	// * worktree: Git worktree object
	// Returns:
	// * dict containing the status of the worktree
	"git-worktree//status?": {
		Argsn: 1,
		Doc:   "Gets the status of a Git worktree.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch wt := arg0.(type) {
			case env.Native:
				if ps.Idx.GetWord(wt.Kind.Index) != "git-worktree" {
					return MakeBuiltinError(ps, "expected a Git worktree object", "git-worktree//status?")
				}

				status, err := wt.Value.(*git.Worktree).Status()
				if err != nil {
					return MakeBuiltinError(ps, fmt.Sprintf("failed to get status: %v", err), "git-worktree//status?")
				}

				// Convert status to a Rye dict
				dict := env.NewDict(make(map[string]any))
				for filePath, fileStatus := range status {
					fileDict := env.NewDict(make(map[string]any))
					fileDict.Data["staging"] = string(fileStatus.Staging)
					fileDict.Data["worktree"] = string(fileStatus.Worktree)
					dict.Data[filePath] = fileDict
				}

				return *dict
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "git-worktree//status?")
			}
		},
	},

	// Tests:
	// equal { repo: git-repo//open "." , repo |git-repo//untracked-files |type? } 'block
	// Args:
	// * repo: Git repository object
	// Returns:
	// * block of untracked files sorted by modification time (newest first)
	"git-repo//untracked-files?": {
		Argsn: 1,
		Doc:   "Lists all untracked files in the repository sorted by last modification time (newest first).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch repo := arg0.(type) {
			case env.Native:
				if ps.Idx.GetWord(repo.Kind.Index) != "git-repo" {
					return MakeBuiltinError(ps, "expected a Git repository object", "git-repo//untracked-files")
				}

				gitRepo := repo.Value.(*git.Repository)

				// Get the worktree
				worktree, err := gitRepo.Worktree()
				if err != nil {
					return MakeBuiltinError(ps, fmt.Sprintf("failed to get worktree: %v", err), "git-repo//untracked-files")
				}

				// Get the status of the worktree
				status, err := worktree.Status()
				if err != nil {
					return MakeBuiltinError(ps, fmt.Sprintf("failed to get status: %v", err), "git-repo//untracked-files")
				}

				// Get the repository path
				repoPath := "."
				// Try to get the repository path from the worktree
				if wt, err := gitRepo.Worktree(); err == nil {
					path := wt.Filesystem.Root()
					if path != "" {
						repoPath = path
					}
				}

				// Collect untracked files
				var untrackedFiles []fileInfo
				for filePath, fileStatus := range status {
					// Check if the file is untracked (denoted by '??' in git status)
					if string(fileStatus.Worktree) == "?" && string(fileStatus.Staging) == "?" {
						// Get the full path of the file
						fullPath := filepath.Join(repoPath, filePath)

						// Get file info to retrieve modification time
						info, err := os.Stat(fullPath)
						if err != nil {
							fmt.Fprintf(os.Stderr, "warning: could not stat file %s: %v\n", fullPath, err)
							continue
						}

						// Skip directories
						if info.IsDir() {
							continue
						}

						// Add file info to the list
						untrackedFiles = append(untrackedFiles, fileInfo{
							path:  filePath,
							mtime: info.ModTime(),
						})
					}
				}

				// Sort files by modification time (newest first)
				sort.Slice(untrackedFiles, func(i, j int) bool {
					return untrackedFiles[i].mtime.After(untrackedFiles[j].mtime)
				})

				// Create a block of file information
				result := make([]env.Object, 0, len(untrackedFiles))
				for _, file := range untrackedFiles {
					fileDict := env.NewDict(make(map[string]any))
					fileDict.Data["path"] = file.path
					fileDict.Data["mtime"] = file.mtime.Format(time.RFC3339)
					result = append(result, *fileDict)
				}

				return *env.NewBlock(*env.NewTSeries(result))
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "git-repo//untracked-files")
			}
		},
	},

	// Tests:
	// equal { repo: git-repo//open "." , repo |git-repo//commits |type? } 'block
	// Args:
	// * repo: Git repository object
	// Returns:
	// * block of commit information
	"git-repo//commits?": {
		Argsn: 1,
		Doc:   "Gets the commit history for a Git repository.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch repo := arg0.(type) {
			case env.Native:
				if ps.Idx.GetWord(repo.Kind.Index) != "git-repo" {
					return MakeBuiltinError(ps, "expected a Git repository object", "git-repo//commits")
				}

				gitRepo := repo.Value.(*git.Repository)

				// Get the reference to HEAD
				ref, err := gitRepo.Head()
				if err != nil {
					return MakeBuiltinError(ps, fmt.Sprintf("failed to get HEAD reference: %v", err), "git-repo//commits")
				}

				// Get the commit history
				commitIter, err := gitRepo.Log(&git.LogOptions{From: ref.Hash()})
				if err != nil {
					return MakeBuiltinError(ps, fmt.Sprintf("failed to get commit history: %v", err), "git-repo//commits")
				}

				// Collect commits
				commits := make([]env.Object, 0)
				err = commitIter.ForEach(func(c *object.Commit) error {
					commitDict := env.NewDict(make(map[string]any))
					commitDict.Data["hash"] = c.Hash.String()
					commitDict.Data["author"] = c.Author.Name
					commitDict.Data["email"] = c.Author.Email
					commitDict.Data["message"] = c.Message
					commitDict.Data["date"] = c.Author.When.Format(time.RFC3339)

					commits = append(commits, *commitDict)
					return nil
				})

				if err != nil {
					return MakeBuiltinError(ps, fmt.Sprintf("failed to iterate commits: %v", err), "git-repo//commits")
				}

				return *env.NewBlock(*env.NewTSeries(commits))
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "git-repo//commits")
			}
		},
	},

	// Tests:
	// equal { repo: git-repo//open "." , repo |git-repo//branches |type? } 'block
	// Args:
	// * repo: Git repository object
	// Returns:
	// * block of branch names
	"git-repo//branches?": {
		Argsn: 1,
		Doc:   "Gets the list of branches in a Git repository.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch repo := arg0.(type) {
			case env.Native:
				if ps.Idx.GetWord(repo.Kind.Index) != "git-repo" {
					return MakeBuiltinError(ps, "expected a Git repository object", "git-repo//branches")
				}

				gitRepo := repo.Value.(*git.Repository)

				// Get the branches
				branchIter, err := gitRepo.Branches()
				if err != nil {
					return MakeBuiltinError(ps, fmt.Sprintf("failed to get branches: %v", err), "git-repo//branches")
				}

				// Collect branches
				branches := make([]env.Object, 0)
				err = branchIter.ForEach(func(ref *plumbing.Reference) error {
					branchDict := env.NewDict(make(map[string]any))
					branchDict.Data["name"] = ref.Name().Short()
					branchDict.Data["hash"] = ref.Hash().String()

					branches = append(branches, *branchDict)
					return nil
				})

				if err != nil {
					return MakeBuiltinError(ps, fmt.Sprintf("failed to iterate branches: %v", err), "git-repo//branches")
				}

				return *env.NewBlock(*env.NewTSeries(branches))
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "git-repo//branches")
			}
		},
	},

	// Tests:
	// equal { repo: git-repo//open "." , repo |git-repo//remotes |type? } 'block
	// Args:
	// * repo: Git repository object
	// Returns:
	// * block of remote information
	"git-repo//remotes?": {
		Argsn: 1,
		Doc:   "Gets the list of remotes in a Git repository.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch repo := arg0.(type) {
			case env.Native:
				if ps.Idx.GetWord(repo.Kind.Index) != "git-repo" {
					return MakeBuiltinError(ps, "expected a Git repository object", "git-repo//remotes")
				}

				gitRepo := repo.Value.(*git.Repository)

				// Get the remotes
				remotes, err := gitRepo.Remotes()
				if err != nil {
					return MakeBuiltinError(ps, fmt.Sprintf("failed to get remotes: %v", err), "git-repo//remotes")
				}

				// Collect remotes
				remotesBlock := make([]env.Object, 0, len(remotes))
				for _, remote := range remotes {
					remoteDict := env.NewDict(make(map[string]any))
					remoteDict.Data["name"] = remote.Config().Name

					urls := make([]env.Object, 0, len(remote.Config().URLs))
					for _, url := range remote.Config().URLs {
						urls = append(urls, *env.NewString(url))
					}

					remoteDict.Data["urls"] = *env.NewBlock(*env.NewTSeries(urls))
					remotesBlock = append(remotesBlock, *remoteDict)
				}

				return *env.NewBlock(*env.NewTSeries(remotesBlock))
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "git-repo//remotes")
			}
		},
	},

	// Tests:
	// equal { repo: git-repo//open "." , repo |git-repo//tags |type? } 'block
	// Args:
	// * repo: Git repository object
	// Returns:
	// * block of tag information
	"git-repo//tags?": {
		Argsn: 1,
		Doc:   "Gets the list of tags in a Git repository.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch repo := arg0.(type) {
			case env.Native:
				if ps.Idx.GetWord(repo.Kind.Index) != "git-repo" {
					return MakeBuiltinError(ps, "expected a Git repository object", "git-repo//tags")
				}

				gitRepo := repo.Value.(*git.Repository)

				// Get the tags
				tagIter, err := gitRepo.Tags()
				if err != nil {
					return MakeBuiltinError(ps, fmt.Sprintf("failed to get tags: %v", err), "git-repo//tags")
				}

				// Collect tags
				tags := make([]env.Object, 0)
				err = tagIter.ForEach(func(ref *plumbing.Reference) error {
					tagDict := env.NewDict(make(map[string]any))
					tagDict.Data["name"] = ref.Name().Short()
					tagDict.Data["hash"] = ref.Hash().String()

					tags = append(tags, *tagDict)
					return nil
				})

				if err != nil {
					return MakeBuiltinError(ps, fmt.Sprintf("failed to iterate tags: %v", err), "git-repo//tags")
				}

				return *env.NewBlock(*env.NewTSeries(tags))
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "git-repo//tags")
			}
		},
	},

	// Tests:
	// equal { repo: git-repo//open "." , repo |git-repo//checkout "branch-name" |type? } 'native
	// Args:
	// * repo: Git repository object
	// * branch: Branch name to checkout
	// Returns:
	// * Git repository object
	"git-repo//checkout": {
		Argsn: 2,
		Doc:   "Checks out a branch in a Git repository.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch repo := arg0.(type) {
			case env.Native:
				if ps.Idx.GetWord(repo.Kind.Index) != "git-repo" {
					return MakeBuiltinError(ps, "expected a Git repository object", "git-repo//checkout")
				}

				gitRepo := repo.Value.(*git.Repository)

				// Get the worktree
				worktree, err := gitRepo.Worktree()
				if err != nil {
					return MakeBuiltinError(ps, fmt.Sprintf("failed to get worktree: %v", err), "git-repo//checkout")
				}

				// Get the branch name
				var branchName string
				switch branch := arg1.(type) {
				case env.String:
					branchName = branch.Value
				default:
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "git-repo//checkout")
				}

				// Checkout the branch
				err = worktree.Checkout(&git.CheckoutOptions{
					Branch: plumbing.NewBranchReferenceName(branchName),
				})
				if err != nil {
					return MakeBuiltinError(ps, fmt.Sprintf("failed to checkout branch: %v", err), "git-repo//checkout")
				}

				return arg0
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "git-repo//checkout")
			}
		},
	},

	// Tests:
	// equal { repo: git-repo//open "." , repo |git-repo//create-branch "new-branch" |type? } 'native
	// Args:
	// * repo: Git repository object
	// * branch: Branch name to create
	// Returns:
	// * Git repository object
	"git-repo//create-branch": {
		Argsn: 2,
		Doc:   "Creates a new branch in a Git repository.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch repo := arg0.(type) {
			case env.Native:
				if ps.Idx.GetWord(repo.Kind.Index) != "git-repo" {
					return MakeBuiltinError(ps, "expected a Git repository object", "git-repo//create-branch")
				}

				gitRepo := repo.Value.(*git.Repository)

				// Get the worktree
				worktree, err := gitRepo.Worktree()
				if err != nil {
					return MakeBuiltinError(ps, fmt.Sprintf("failed to get worktree: %v", err), "git-repo//create-branch")
				}

				// Get the branch name
				var branchName string
				switch branch := arg1.(type) {
				case env.String:
					branchName = branch.Value
				default:
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "git-repo//create-branch")
				}

				// Get the HEAD reference
				headRef, err := gitRepo.Head()
				if err != nil {
					return MakeBuiltinError(ps, fmt.Sprintf("failed to get HEAD reference: %v", err), "git-repo//create-branch")
				}

				// Create the branch
				branchRef := plumbing.NewBranchReferenceName(branchName)
				// Create a reference for the new branch
				err = gitRepo.Storer.SetReference(plumbing.NewHashReference(branchRef, headRef.Hash()))
				if err != nil {
					return MakeBuiltinError(ps, fmt.Sprintf("failed to create branch: %v", err), "git-repo//create-branch")
				}

				// Checkout the branch
				err = worktree.Checkout(&git.CheckoutOptions{
					Branch: branchRef,
				})
				if err != nil {
					return MakeBuiltinError(ps, fmt.Sprintf("failed to checkout branch: %v", err), "git-repo//create-branch")
				}

				return arg0
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "git-repo//create-branch")
			}
		},
	},

	// Tests:
	// equal { repo: git-repo//open "." , wt: repo |git-repo//worktree , wt |git-worktree//add "file.txt" |type? } 'native
	// Args:
	// * worktree: Git worktree object
	// * path: Path of the file to add
	// Returns:
	// * Git worktree object
	"git-worktree//add": {
		Argsn: 2,
		Doc:   "Adds a file to the Git index.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch wt := arg0.(type) {
			case env.Native:
				if ps.Idx.GetWord(wt.Kind.Index) != "git-worktree" {
					return MakeBuiltinError(ps, "expected a Git worktree object", "git-worktree//add")
				}

				worktree := wt.Value.(*git.Worktree)

				// Get the file path
				var filePath string
				switch path := arg1.(type) {
				case env.String:
					filePath = path.Value
				case env.Uri:
					filePath = path.GetPath()
				default:
					return MakeArgError(ps, 2, []env.Type{env.StringType, env.UriType}, "git-worktree//add")
				}

				// Add the file to the index
				_, err := worktree.Add(filePath)
				if err != nil {
					return MakeBuiltinError(ps, fmt.Sprintf("failed to add file: %v", err), "git-worktree//add")
				}

				return arg0
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "git-worktree//add")
			}
		},
	},

	// Tests:
	// equal { repo: git-repo//open "." , wt: repo |git-repo//worktree , wt |git-worktree//commit "Commit message" |type? } 'native
	// Args:
	// * worktree: Git worktree object
	// * message: Commit message
	// * author: Optional author name (default: "Rye User")
	// * email: Optional author email (default: "rye@example.com")
	// Returns:
	// * Git worktree object
	"git-worktree//commit": {
		Argsn: 2,
		Doc:   "Commits changes to the Git repository.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch wt := arg0.(type) {
			case env.Native:
				if ps.Idx.GetWord(wt.Kind.Index) != "git-worktree" {
					return MakeBuiltinError(ps, "expected a Git worktree object", "git-worktree//commit")
				}

				worktree := wt.Value.(*git.Worktree)

				// Get the commit message
				var message string
				switch msg := arg1.(type) {
				case env.String:
					message = msg.Value
				default:
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "git-worktree//commit")
				}

				// Get optional author name and email
				authorName := "Rye User"
				authorEmail := "rye@example.com"

				if arg2 != nil {
					switch author := arg2.(type) {
					case env.String:
						authorName = author.Value
					default:
						return MakeArgError(ps, 3, []env.Type{env.StringType}, "git-worktree//commit")
					}
				}

				if arg3 != nil {
					switch email := arg3.(type) {
					case env.String:
						authorEmail = email.Value
					default:
						return MakeArgError(ps, 4, []env.Type{env.StringType}, "git-worktree//commit")
					}
				}

				// Commit the changes
				_, err := worktree.Commit(message, &git.CommitOptions{
					Author: &object.Signature{
						Name:  authorName,
						Email: authorEmail,
						When:  time.Now(),
					},
				})
				if err != nil {
					return MakeBuiltinError(ps, fmt.Sprintf("failed to commit changes: %v", err), "git-worktree//commit")
				}

				return arg0
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "git-worktree//commit")
			}
		},
	},
}
