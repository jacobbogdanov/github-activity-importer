# github-activity-importer

Import anonymized git contributions from one private repo into another.

This is useful if you contribute code to a git repo that is **not** on GitHub, but you want your contributor graph to reflect all your hard work.

## Usage

```text
$ gh-activity-importer -help
  -dest-author-email string
        The email of the author to save the commits as.
        If omitted this will be the same as the source-author-email.
  -dest-author-name string
        The name of the author to save the commits as.
        If omitted this will be the same as the source-author-name.
  -dest-repo string
        File path to the git repository where commits will be saved to.
  -end-date string
        Limit the search in the source-repo to be before a certain date. (default "today")
  -source-author-email string
        The email of the author in the source-repo to find commits for.
        At least one of source-author-name and source-author-email are required.
  -source-author-name string
        The name of the author in the source-repo to find commits for.
        At least one of source-author-name and source-author-email are required.
  -source-repo string
        File path to the git repository where commits will be read from.
  -start-date string
        Limit the search in the source-repo to be after a certain date.
```

## Example

First you have to chose your destination repository.

Either,

### Connect to GitHub

If you would like to set this up with your github account then you can follow
the instructions on how to [create a git repository on github](https://docs.github.com/en/free-pro-team@latest/github/creating-cloning-and-archiving-repositories/creating-a-new-repository) and
then [clone it locally](https://docs.github.com/en/free-pro-team@latest/github/creating-cloning-and-archiving-repositories/cloning-a-repository).

OR

### Create a local repository

If you would like to just mess around with `gh-activity-importer` then you can
create a repository locally by running

```sh
$ git init ~/my-activity
Initialized empty Git repository in my-activity/.git/
```

### Running `gh-activity-importer`

```sh
$ gh-activity-importer \
    -source-repo=/company/private/repo \
    -dest-repo=~/my-activity \
    -source-author-email=username@company.com \
    -dest-auth-email=usename@gmail.com \
    -dest-user-name="User Name"
```

This will collect all the dates from the git commits in `/company/private/repo`
authored using the email `username@company.com`, the create commits in the
`~/my-activity` git repository using the _same dates_ but **no other information**.
