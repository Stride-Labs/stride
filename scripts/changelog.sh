#!/bin/bash

set -eu

VERSION_REGEX='v[0-9]{1,2}\.[0-9]{1}\.[0-9]{1}$'

# Validate script parameters
if [ -z "$OLD_VERSION" ]; then
    echo "OLD_VERSION must be set (e.g. v8.0.0). Exiting..."
    exit 1
fi

if [ -z "$NEW_VERSION" ]; then
    echo "NEW_VERSION must be set (e.g. v8.0.0). Exiting..."
    exit 1
fi

if ! echo $OLD_VERSION | grep -Eq $VERSION_REGEX; then 
    echo "OLD_VERSION must be of form {major}.{minor}.{patch} (e.g. 8.0.0). Exiting..."
    exit 1
fi 

if ! echo $NEW_VERSION | grep -Eq $VERSION_REGEX; then 
    echo "NEW_VERSION must be of form {major}.{minor}.{patch} (e.g. 8.0.0). Exiting..."
    exit 1
fi 

GITHUB_PR_URL="https://github.com/Stride-Labs/stride/pull"
GO_FILES='^x/.*\.go$|^app/.*\.go$'
TEST_FILES='.*_test\.go$'
CURRENT_DATE=$(date +'%Y-%m-%d')
NEW_MAJOR_VERSION=$(echo "$NEW_VERSION" | cut -d '.' -f 1)

TEMP_COMMITS="TEMP_COMMITS.txt"
TEMP_ON_CHAIN_CHANGELOG="TEMP_ON_CHAIN_CHANGELOG.md"
TEMP_OFF_CHAIN_CHANGELOG="TEMP_OFF_CHAIN_CHANGELOG.md"
TEMP_MAIN_CHANGELOG="TEMP_MAIN_CHANGELOG.md"

UPGRADE_CHANGELOG="app/upgrades/$NEW_MAJOR_VERSION/README.md"
MAIN_CHANGELOG="CHANGELOG.md"
MAIN_CHANGELOG_INSERT_STATEMENT="<!-- GH ACTIONS TEMPLATE - INSERT NEW VERSION HERE -->"

# Checks if a given commit hash modified an on-chain file
# If so, returns true (0), otherwise returns false (1)
modified_on_chain_file() {
  commit=$1

  # Get all modified 
  modified_files=$(git diff-tree --no-commit-id --name-only -r $commit)

  # Filter for go files in x/ or x/app, that were not test files
  modified_on_chain_files=$(echo "$modified_files" | grep -E "$GO_FILES" | grep -vE "$TEST_FILES")

  # If there were any modified on-chain files, return true
  if [[ -n "$modified_on_chain_files" ]]; then
    return 0 # true
  fi

  # Otherwise, return false
  return 1 # false
}

# First we'll gather the list of all commits between the new and old version
# The last line does not end with a new line character, so we have to append a new empty line
# so that the last line is picked up
git log --pretty=format:"%H %s" ${OLD_VERSION}..${NEW_VERSION} --reverse > $TEMP_COMMITS
echo "" >> $TEMP_COMMITS

# Then we'll loop through those commits and build out the commit descriptions for the on-chain and off-chain sections
on_chain_index=1
off_chain_index=1
cat $TEMP_COMMITS | while read line; do
  commit_hash=$(echo $line | cut -d' ' -f1)
  commit_title=$(echo $line | cut -d' ' -f2-)

  # Pull out the PR number (e.g. "Added LSM Support (#803)" -> 803)
  pr_number=$(echo $commit_title | sed -n 's/.*#\([0-9]*\).*/\1/p')

  # Build the commit description by replacing the PR number with the url
  # (e.g. "Added LSM Support (#803)" -> "Added LSM Support ([#803](https://github.com/Stride-Labs/stride/pull/803))"
  description=$(echo $commit_title | sed "s|#$pr_number|[#$pr_number]($GITHUB_PR_URL/$pr_number)|")

  # Write the description to the relevant file, based on whether it changed anything on chain
  if modified_on_chain_file $commit_hash; then
    echo "$on_chain_index. $description" >> $TEMP_ON_CHAIN_CHANGELOG
    on_chain_index=$((on_chain_index+1))
  else
    echo "$off_chain_index. $description" >> $TEMP_OFF_CHAIN_CHANGELOG
    off_chain_index=$((off_chain_index+1))
  fi
done 

# Now build the main and upgrade changelogs
# For the main changelog, we'll first write out just the upgrade section to a temporary file and then
# insert it into the main file
echo "## [$NEW_VERSION](https://github.com/Stride-Labs/stride/releases/tag/$NEW_VERSION) - $CURRENT_DATE" >> $TEMP_MAIN_CHANGELOG

# If there were on-chain changes, add the relevant sections to the main and upgrade changelog
if [[ $on_chain_index -gt 1 ]]; then
  echo -e "\n### On-Chain changes" >> $TEMP_MAIN_CHANGELOG
  echo "$TEMP_ON_CHAIN_CHANGELOG" >> $TEMP_MAIN_CHANGELOG

  echo "# Upgrade $NEW_MAJOR_VERSION Changelog" > $UPGRADE_CHANGELOG
  echo "$TEMP_ON_CHAIN_CHANGELOG" >> $UPGRADE_CHANGELOG
fi

# If there were off-chain changes, only add them to the main change log
if [[ $off_chain_index -gt 1 ]]; then
  echo -e "\n### Off-Chain changes" >> $TEMP_MAIN_CHANGELOG
  echo "$TEMP_OFF_CHAIN_CHANGELOG" >> $TEMP_MAIN_CHANGELOG
fi

# Insert the temporary main changelog into the actual file
sed -i -e "/$MAIN_CHANGELOG_INSERT_STATEMENT/r $TEMP_MAIN_CHANGELOG" $MAIN_CHANGELOG

# Finally, cleanup all the temp files
rm $TEMP_COMMITS
rm $TEMP_MAIN_CHANGELOG
rm $TEMP_ON_CHAIN_CHANGELOG
rm $TEMP_OFF_CHAIN_CHANGELOG
