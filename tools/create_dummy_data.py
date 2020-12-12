#!/usr/bin/env python3

"""A dumb utility to create dummy git data."""

# TODO re-write this using go-git... or better yet use an in-memory fs and write this in
# a go test!

import argparse
import datetime
import pathlib
import random
import subprocess
import sys
import uuid
from email import utils as email_utils
from typing import List, Union


def main(argv: List[str]) -> Union[int, str]:
    args = parse_args(argv)

    if args.weeks <= 0:
        return "weeks must be greather than zero"

    authors = [f"{args.user_name} <{args.email}>"]
    if args.include_other_users:
        authors.extend(
            (
                "Jimmy <jimmy@example.com>",
                "Bobby Tables <btables@example.com>",
                "Timmy <timmy@example.com>",
            )
        )

    now = datetime.datetime.now()
    current = now - datetime.timedelta(weeks=args.weeks)

    while current < now:
        try:
            create_commit(current, random.choice(authors))
        except subprocess.CalledProcessError as err:
            return str(err)

        current += datetime.timedelta(
            hours=random.randint(0, 24),
            minutes=random.randint(1, 60),
        )

    return 0


def parse_args(argv: List[str]) -> argparse.Namespace:
    parser = argparse.ArgumentParser()
    parser.add_argument(
        "--weeks",
        type=int,
        default=5,
        help="The number of weeks in the past to start creating git commits",
    )

    parser.add_argument("--email", required=True, help="The email of the git user")
    parser.add_argument("--user-name", required=True, help="The name of the git user")

    parser.add_argument(
        "--include-other-users",
        action="store_true",
        default=False,
        help="Add additional users to git history",
    )

    return parser.parse_args(argv[1:])


def create_commit(timestamp: datetime.datetime, author: str) -> None:
    """Create a randomly named file and commit it using the provided metadata."""
    name = str(uuid.uuid4())
    pathlib.Path(name).touch()
    subprocess.run(["git", "add", name])
    subprocess.run(
        [
            "git",
            "commit",
            "--no-verify",
            "-m",
            name,
            "--author",
            author,
            "--date",
            email_utils.format_datetime(timestamp),
        ],
        check=True,
    )


if __name__ == "__main__":
    sys.exit(main(sys.argv))
