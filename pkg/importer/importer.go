// Package importer provides the ability to find commits by a git author in one repo and create dummy
// commits by a new author in another repository.
package importer

import (
	"fmt"
	"path/filepath"
	"strconv"
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
}

// Run is the main entry point to run the importer application.
func (app *Importer) Run() error {
	source, err := git.PlainOpen(app.SourceRepo)
	if err != nil {
		return fmt.Errorf("failed to open source repository '%s': %s", app.SourceRepo, err)
	}

	dest, err := git.PlainOpen(app.DestRepo)
	if err != nil {
		return fmt.Errorf("failed to open destination repository '%s': %s", app.DestRepo, err)
	}

	timestamps, err := app.find(source)
	if err != nil {
		return err
	}

	// go-git has no equivalent to `git log --reverse`. See https://github.com/go-git/go-git/issues/60
	for i := len(timestamps) - 1; i >= 0; i-- {
		if err := app.saveOne(dest, timestamps[i]); err != nil {
			return err
		}
	}

	return nil
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
