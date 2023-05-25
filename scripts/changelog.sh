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

GITHUB_COMMIT_URL="https://github.com/Stride-Labs/stride/commit"
GITHUB_PR_URL="https://github.com/org/repo/pull"
ON_CHAIN_FILES='"x/**/*.go" "app/**/*.go" ":(exclude)**/*_test.go"'
CURRENT_DATE=$(date +'%Y-%m-%d')
NEW_MAJOR_VERSION=$(echo "$NEW_VERSION" | cut -d '.' -f 1)

TEMP_CHANGELOG="TEMP_CHANGELOG.md"
UPGRADE_CHANGELOG="app/upgrades/$NEW_MAJOR_VERSION/README.md"
MAIN_CHANGELOG="CHANGELOG.md"
MAIN_CHANGELOG_INSERT_STATEMENT="<!-- GH ACTIONS TEMPLATE - INSERT NEW VERSION HERE -->"

# First write all changes to the main changelog
# Write the version summary
# TODO: Build On-Chain vs Off-Chain sections dynamically. I'm not sure how to 
# create the filter for Off-Chain
echo "### Changelog" > $TEMP_CHANGELOG
echo "## [$NEW_VERSION](https://github.com/Stride-Labs/stride/releases/tag/$NEW_VERSION) - $CURRENT_DATE" >> $TEMP_CHANGELOG
echo "!!!ACTION ITEM: Move the following to the On-Chain vs Off-chain sections!!!" >> $TEMP_CHANGELOG

i=1
git log --pretty=format:"%h %H %s" ${OLD_VERSION}..main | while read LINE; do
  SHORT_COMMIT_HASH=$(echo $LINE | cut -d' ' -f1)
  LONG_COMMIT_HASH=$(echo $LINE | cut -d' ' -f2)
  COMMIT_TITLE=$(echo $LINE | cut -d' ' -f3-)
  PR_NUMBER=$(echo $COMMIT_TITLE | grep -oP '#\K\w+')
  COMMIT_DESCRIPTION=$(echo $COMMIT_TITLE | sed "s|#$PR_NUMBER|[#$PR_NUMBER]($GITHUB_PR_URL/$PR_NUMBER)|")
  echo "$i. $COMMIT_DESCRIPTION [[${SHORT_COMMIT_HASH}]($GITHUB_COMMIT_URL/${LONG_COMMIT_HASH})]" >> $TEMP_CHANGELOG
  i=$((i+1))
done

echo -e "\n### On-Chain changes" >> $TEMP_CHANGELOG
echo -e "\n### Off-Chain changes" >> $TEMP_CHANGELOG
echo "These changes do not affect any on-chain functionality, but have been implemented since \`$OLD_VERSION\`" >> $TEMP_CHANGELOG

sed -i -e "/$MAIN_CHANGELOG_INSERT_STATEMENT/r $TEMP_CHANGELOG" $MAIN_CHANGELOG
rm $TEMP_CHANGELOG

# Next write all the on chain changes to the upgrade changelog
i=1
git log --pretty=format:"%h %H %s" ${OLD_VERSION}..main -- "x/**/*.go" "app/**/*.go" ":(exclude)**/*_test.go" | while read LINE; do
  if [[ "$i" == "1" ]]; then
    echo "# Upgrade $NEW_MAJOR_VERSION Changelog" > $UPGRADE_CHANGELOG
  fi
  SHORT_COMMIT_HASH=$(echo $LINE | cut -d' ' -f1)
  LONG_COMMIT_HASH=$(echo $LINE | cut -d' ' -f2)
  COMMIT_TITLE=$(echo $LINE | cut -d' ' -f3-)
  PR_NUMBER=$(echo $COMMIT_TITLE | grep -oP '#\K\w+')
  COMMIT_DESCRIPTION=$(echo $COMMIT_TITLE | sed "s|#$PR_NUMBER|[#$PR_NUMBER]($GITHUB_PR_URL/$PR_NUMBER)|")
  echo "$i. $COMMIT_DESCRIPTION [[${SHORT_COMMIT_HASH}]($GITHUB_COMMIT_URL/${LONG_COMMIT_HASH})]" >> $UPGRADE_CHANGELOG
  i=$((i+1))
done
