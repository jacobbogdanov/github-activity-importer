// Package importer provides the ability to find commits by a git author in one repo and create dummy
// commits by a new author in another repository.
package importer

import (
	"fmt"
	"log"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// User represents a git user.
type User struct {
	Name  string
	Email string
}

// Importer provides the ability to find commits by a git author in one repo and create dummy commits
// by a new author in another repository.
type Importer struct {
	// Information about the source git repository.
	SourceRepo   string
	SourceAuthor User
	Start        time.Time
	End          time.Time

	// Information about the destination git repository.
	DestRepo   string
	DestAuthor User

	// Control where the importer will log to.
	Logger *log.Logger
}

// Run is the main entry point to run the importer application.
func (app *Importer) Run() error {
	source, err := app.openRepo(app.SourceRepo)
	if err != nil {
		return fmt.Errorf("error loading source repo: %v", err)
	}

	dest, err := app.openRepo(app.DestRepo)
	if err != nil {
		return fmt.Errorf("error loading dest repo: %v", err)
	}

	timestamps, err := app.find(source)
	if err != nil {
		return err
	}

	if len(timestamps) == 0 {
		return fmt.Errorf(
			"source repository contains zero commits from author '%s' between %s and %s",
			app.SourceAuthor, app.Start, app.End)
	}

	// go-git has no equivalent to `git log --reverse`. See https://github.com/go-git/go-git/issues/60
	for i := len(timestamps) - 1; i >= 0; i-- {
		if err := app.saveOne(dest, timestamps[i]); err != nil {
			return err
		}
	}

	app.Logger.Printf(
		"successfully transferred %d commits from %s to %s\n\n"+
			"Verify the output of the destination repository (by running something like `git log`.)\n"+
			"If the output looks good push to remote, wait a few minutes, then check your github profile!",
		len(timestamps), app.SourceRepo, app.DestRepo)

	return nil
}

// openRepo takes either a URL or a filepath and opens the git repository for reading or writing.
func (app *Importer) openRepo(filepathOrURL string) (*git.Repository, error) {
	parts := strings.Split(filepathOrURL, "://")
	if len(parts) == 1 {
		return openLocalRepo(filepathOrURL)
	}
	if len(parts) != 2 {
		return nil, fmt.Errorf("failed to determine protocol from URL or filepath '%s': found protocol separator '://' more than once", filepathOrURL)
	}
	if parts[0] == "file" {
		return openLocalRepo(parts[1])
	}

	app.Logger.Println("cloning remote repository into local folder 'dest'")
	repo, err := git.PlainClone(
		"dest",
		false,
		&git.CloneOptions{
			URL:      filepathOrURL,
			Progress: app.Logger.Writer(),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to clone remote repository '%s': %v", filepathOrURL, err)
	}

	return repo, nil
}

func openLocalRepo(filepath string) (*git.Repository, error) {
	repo, err := git.PlainOpen(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open local git repository '%s': %v", filepath, err)
	}

	return repo, nil
}

func (app *Importer) find(repo *git.Repository) ([]time.Time, error) {
	ref, err := repo.Head()
	if err != nil {
		return nil, err
	}

	cIter, err := repo.Log(&git.LogOptions{
		From:  ref.Hash(),
		Order: git.LogOrderCommitterTime,
		Since: &app.Start,
		Until: &app.End,
	})
	if err != nil {
		return nil, err
	}

	timestamps := []time.Time{}
	err = cIter.ForEach(func(c *object.Commit) error {
		if app.matchesSource(c.Author) {
			timestamps = append(timestamps, c.Author.When)
		}

		return nil
	})

	return timestamps, err
}

func (app *Importer) matchesSource(author object.Signature) bool {
	if app.SourceAuthor.Name != "" && app.SourceAuthor.Name != author.Name {
		return false
	}

	if app.SourceAuthor.Email != "" && app.SourceAuthor.Email != author.Email {
		return false
	}

	return true
}

func (app *Importer) saveOne(repo *git.Repository, timestamp time.Time) error {
	worktree, err := repo.Worktree()
	if err != nil {
		return err
	}

	dir := filepath.Join(strconv.Itoa(timestamp.Year()), strconv.Itoa(int(timestamp.Month())))

	err = worktree.Filesystem.MkdirAll(dir, 0600)
	if err != nil {
		return err
	}

	// TODO this doesn't work if two commits happen at the same time.
	_, err = worktree.Filesystem.Create(filepath.Join(dir, timestamp.String()))
	if err != nil {
		return err
	}

	author := object.Signature{
		Name:  app.DestAuthor.Name,
		Email: app.DestAuthor.Email,
		When:  timestamp,
	}

	commit, err := worktree.Commit(
		timestamp.String(),
		&git.CommitOptions{Author: &author, Committer: &author},
	)
	if err != nil {
		return err
	}

	if _, err = repo.CommitObject(commit); err != nil {
		return err
	}
	return nil
}
