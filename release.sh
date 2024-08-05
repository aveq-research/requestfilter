#!/bin/bash
#
# Bump the version, run auto-changelog, and push to Git

# ==============================================================================
# PRINTING STUFF

RED="\033[1;31m"
GREEN="\033[0;32m"
YELLOW="\033[1;33m"
WHITE="\033[1;37m"
RESET="\033[0m"

QUESTION_FLAG="❔"
WARNING_FLAG="❕"
ERROR_FLAG="🛑"
NOTICE_FLAG="❯"

# ==============================================================================
# FUNCTIONS

_warn() {
  echo -e "$YELLOW$WARNING_FLAG $1$RESET"
}

_info() {
  echo -e "$WHITE$NOTICE_FLAG $1$RESET"
}

_question() {
  echo -e "$GREEN$QUESTION_FLAG  $1$RESET"
}

_error() {
  echo -e "$RED$ERROR_FLAG $1$RESET"
}

# ==============================================================================

set -e

cd "$(dirname "$0")"

FORCE=false

usage() {
    echo "Usage: $0 [options] VERSION"
    echo
    echo "VERSION:"
    echo "  major: bump major version number"
    echo "  minor: bump minor version number"
    echo "  patch: bump patch version number"
    echo
    echo "Options:"
    echo "  -f, --force:  force release"
    echo "  -h, --help:   show this help message"
    exit 1
}

# parse args
while [ "$#" -gt 0 ]; do
    case "$1" in
    -f | --force)
        FORCE=true
        shift
        ;;
    -h | --help)
        usage
        ;;
    *)
        break
        ;;
    esac
done

# check if version is specified
if [ "$#" -ne 1 ]; then
    usage
fi

if [ "$1" != "major" ] && [ "$1" != "minor" ] && [ "$1" != "patch" ]; then
    usage
fi

# check if git is clean and force is not enabled
if ! git diff-index --quiet HEAD -- && [ "$FORCE" = false ]; then
    _error "Error: git is not clean. Please commit all changes first."
    exit 1
fi


if ! command -v gitchangelog &> /dev/null; then
    _error "gitchangelog is not installed or not configured properly."
fi

VERSION_FILE="VERSION"

_info "Bumping $VERSION_FILE"

current_version=$(grep -Eo '[0-9]+\.[0-9]+\.[0-9]+' "$VERSION_FILE")

MAJOR=$(cut -d "." -f 1 <<<"$current_version")
MINOR=$(cut -d "." -f 2 <<<"$current_version")
PATCH=$(cut -d "." -f 3 <<<"$current_version")

PREVIOUS_MAJOR=$MAJOR
PREVIOUS_MINOR=$MINOR
PREVIOUS_PATCH=$PATCH

_info "Current version: $MAJOR.$MINOR.$PATCH"

if [ "$1" == "major" ]; then
    MAJOR=$((MAJOR + 1))
    MINOR=0
    PATCH=0
elif [ "$1" == "minor" ]; then
    MINOR=$((MINOR + 1))
    PATCH=0
elif [ "$1" == "patch" ]; then
    PATCH=$((PATCH + 1))
fi

_info "New version: $MAJOR.$MINOR.$PATCH"

# prompt for confirmation
if [ "$FORCE" = false ]; then
    read -p "Do you want to release? [yY] " -n 1 -r
    echo
else
    REPLY="y"
fi

if [[ $REPLY =~ ^[Yy]$ ]]; then
  # replace the version
  new_version="$MAJOR.$MINOR.$PATCH"
  perl -pi -e "s/$current_version/$new_version/g" "$VERSION_FILE"

  git add "$VERSION_FILE"

  # bump initially but to not push yet
  git commit --no-verify -m "bump version to ${new_version}"
  git tag -a -m "Tag version ${new_version}" "v$new_version"

  # generate the changelog
  _info "Generating changelog ..."
  gitchangelog > CHANGELOG.md

  # add the changelog
  git add CHANGELOG.md
  git commit --no-verify --amend --no-edit
  git tag -a -f -m "Tag version ${new_version}" "v$new_version"

  # push to remote
  _info "Pushing to remote ..."
  git push && git push --tags
else
  _warn "Aborted."
  exit 1
fi
