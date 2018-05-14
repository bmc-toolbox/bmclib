#!/bin/bash
set -e

LAST_VERSION=`git describe --abbrev=0 --tags`;
NEW_VERSION=$(TZ=UTC date +%Y%m%d_%H%M%S)

echo "Making sure repo is up to date";
git pull --rebase;
if [ $? -ne 0 ];
then
    exit 1
fi

echo "Pushing changes";
git push;
echo "*******************************";
echo "**** Bumping ${LAST_VERSION} to ${NEW_VERSION}";
echo "*******************************";
echo "Hit [Enter] to continue, Ctrl+C to abort:"
read USERINPUT;

echo "Creating new tag ${NEW_VERSION}";
git tag -a ${NEW_VERSION} -m ${NEW_VERSION};

echo "Pushing new tag upstream";
git push origin --tags;
