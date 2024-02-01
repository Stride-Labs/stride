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

MAIN_CHANGELOG="CHANGELOG.md"
MAIN_CHANGELOG_INSERT_STATEMENT="<!-- GH ACTIONS TEMPLATE - INSERT NEW VERSION HERE -->"

# Checks if a given commit hash modified an on-chain file
# If so, returns true (0), otherwise returns false (1)
modified_on_chain_file() {
  commit=$1

  # Get all modified files
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

# Gather the list of all commits between the old version and main
# The output on each line will be in the format: {commit_hash} {commit_title}
# The last line does not end with a new line character, so we have to append a new empty line
# so that the full output is propogated into the while loop
git log --pretty=format:"%H %s" ${OLD_VERSION}..main --reverse > $TEMP_COMMITS
echo "" >> $TEMP_COMMITS

# Loop through the commits and build out the commit descriptions for the on-chain and off-chain sections
on_chain_index=1
off_chain_index=1
cat $TEMP_COMMITS | while read line; do
  commit_hash=$(echo $line | cut -d' ' -f1)
  commit_title=$(echo $line | cut -d' ' -f2-)

  # Pull out the PR number (e.g. "Added LSM Support (#803)" -> "803")
  pr_number=$(echo $commit_title | sed -n 's/.*#\([0-9]*\).*/\1/p')

  # Build the commit description by replacing the PR number with the full url
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

# Build a temporary changelog file with the just changes from this upgrade
# This will later be inserted into the main changelog file
echo -e "\n## [$NEW_VERSION](https://github.com/Stride-Labs/stride/releases/tag/$NEW_VERSION) - $CURRENT_DATE" > $TEMP_MAIN_CHANGELOG

# If there were on-chain changes, add the "On-Chain" section
if [[ -s "$TEMP_ON_CHAIN_CHANGELOG" ]]; then
  echo -e "\n### On-Chain changes" >> $TEMP_MAIN_CHANGELOG
  cat "$TEMP_ON_CHAIN_CHANGELOG" >> $TEMP_MAIN_CHANGELOG
fi

# If there were off-chain changes, add the "Off-Chain" section
if [[ -s "$TEMP_OFF_CHAIN_CHANGELOG" ]]; then
  echo -e "\n### Off-Chain changes" >> $TEMP_MAIN_CHANGELOG
  cat "$TEMP_OFF_CHAIN_CHANGELOG" >> $TEMP_MAIN_CHANGELOG
fi

# Insert the temporary changelog into the main file
echo "" >> $TEMP_MAIN_CHANGELOG
sed -i -e "/$MAIN_CHANGELOG_INSERT_STATEMENT/r $TEMP_MAIN_CHANGELOG" $MAIN_CHANGELOG

# Finally, cleanup all the temp files
rm -f $TEMP_COMMITS
rm -f $TEMP_MAIN_CHANGELOG
rm -f $TEMP_ON_CHAIN_CHANGELOG
rm -f $TEMP_OFF_CHAIN_CHANGELOG
