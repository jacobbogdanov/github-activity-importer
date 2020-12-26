// Package main provides a command line interface to find commits by a git author in one repo and
// create dummy commits by a new author in another repository.
package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jacobbogdanov/github-activity-importer/pkg/importer"
)

func main() {
	app := importer.Importer{}

	flag.StringVar(&app.SourceRepo, "source-repo", "", "File path to the git repository where commits will be read from.")
	flag.StringVar(&app.DestRepo, "dest-repo", "", "File path to the git repository where commits will be saved to.")

	startFlag := flag.String("start-date", "", "Limit the search in the source-repo to be after a certain date.")
	endFlag := flag.String("end-date", endDateDefault(), "Limit the search in the source-repo to be before a certain date.")

	flag.StringVar(&app.SourceAuthor.Name, "source-author-name", "", "The name of the author in the source-repo to find commits for.\nAt least one of source-author-name and source-author-email are required.")
	flag.StringVar(&app.SourceAuthor.Email, "source-author-email", "", "The email of the author in the source-repo to find commits for.\nAt least one of source-author-name and source-author-email are required.")

	flag.StringVar(&app.DestAuthor.Name, "dest-author-name", "", "The name of the author to save the commits as.\nIf omitted this will be the same as the source-author-name.")
	flag.StringVar(&app.DestAuthor.Email, "dest-author-email", "", "The email of the author to save the commits as.\nIf omitted this will be the same as the source-author-email.")

	flag.Parse()

	if app.SourceAuthor.Name == "" && app.SourceAuthor.Email == "" {
		fmt.Println("at least one of 'source-author-name' and 'source-author-email' is required")
		os.Exit(2)
	}

	if app.DestAuthor.Name == "" {
		app.DestAuthor.Name = app.SourceAuthor.Name
	}

	if app.DestAuthor.Email == "" {
		app.DestAuthor.Email = app.SourceAuthor.Email
	}

	if *startFlag != "" {
		start, err := parseTime(*startFlag)
		if err != nil {
			fmt.Printf("failed to parse start-date: %s\n", err)
			os.Exit(2)
		}
		app.Start = start
	}

	end, err := parseTime(*endFlag)
	if err != nil {
		fmt.Printf("failed to parse end-date: %s\n", err)
		os.Exit(2)
	}
	app.End = end

	if app.Start.After(app.End) {
		fmt.Println("start-date must before end-date")
		os.Exit(2)
	}

	if err := app.Run(); err != nil {
		fmt.Printf("gh-activity-importer encountered an error: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("successfully transferred git activity!\n")
}

func parseTime(value string) (time.Time, error) {
	parts := strings.Split(value, "/")
	if len(parts) != 3 {
		return time.Time{}, fmt.Errorf("failed to parse timestamp: expected slash-delimited format 'year/month/day', got '%s'", value)
	}

	year, err := strconv.Atoi(parts[0])
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse year '%s': %s", parts[0], err)
	}

	month, err := strconv.Atoi(parts[1])
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse month '%s': %s", parts[1], err)
	}

	day, err := strconv.Atoi(parts[2])
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse day '%s': %s", parts[2], err)
	}

	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC), nil
}

func endDateDefault() string {
	now := time.Now()
	return fmt.Sprintf("%d/%d/%d", now.Year(), now.Month(), now.Day())
}
