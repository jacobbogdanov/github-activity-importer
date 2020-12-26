// Package main provides a command line interface to find commits by a git author in one repo and
// create dummy commits by a new author in another repository.
package main

import (
	"flag"
	"fmt"
	"log"
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

	app.Logger = log.New(os.Stderr, "", 0)

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
		start, err := parseTime(*startFlag, lowerBound)
		if err != nil {
			fmt.Printf("failed to parse start-date: %s\n", err)
			os.Exit(2)
		}
		app.Start = start
	}

	end, err := parseTime(*endFlag, upperBound)
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

// parseTime parses a string int a time.Time with the expected format of 'year/month/day'.
func parseTime(value string, bound Bound) (time.Time, error) {
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
	if month < 1 || month > 12 {
		return time.Time{}, fmt.Errorf("invalid month '%d': outside valid range 1-12", month)
	}

	day, err := strconv.Atoi(parts[2])
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse day '%s': %s", parts[2], err)
	}

	// This isn't proper validation, but time.Date can deal with the overflow for shorter months.
	if day < 1 || day > 31 {
		return time.Time{}, fmt.Errorf("invalid day '%d': outside valid range 1-31", month)
	}

	date := time.Date(year, time.Month(month), day, 24, 0, 0, 0, time.UTC)
	if bound == upperBound {
		return date.AddDate(0, 0, 1), nil
	}

	return date, nil
}

// Bound determines whether to round the date up or down.
type Bound int

const (
	// lowerBound means that the day should be interpreted as a starting point. This means it's the
	// START of the day.
	lowerBound Bound = 0
	// upperBound means that the day should be interpreted as an ending point. This means it's the
	// END of the day.
	upperBound Bound = 1
)

func endDateDefault() string {
	now := time.Now()
	return fmt.Sprintf("%d/%d/%d", now.Year(), now.Month(), now.Day())
}
