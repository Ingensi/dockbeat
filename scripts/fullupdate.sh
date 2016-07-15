#!/usr/bin/env bash

set -e

BEATNAME=$1
BEATPATH=$2

# Setup
if [ -z $BEATNAME ]; then
    echo "beat name must be set"
    exit;
fi

if [ -z $BEATPATH ]; then
    echo "beat path must be set"
    exit;
fi

DIR=../$BEATNAME

if [ ! -d "$DIR" ]; then
  echo "Beat dir does not exist: $DIR"
  exit;
fi

echo "Beat name: $BEATNAME"
echo "Beat path: $DIR"

cd $DIR

echo "Start modifying beat"

# Update config
echo "Update docker config file"
rm -f $BEATNAME-docker.yml
cp $BEATNAME.yml $BEATNAME-docker.yml
cat $BEATNAME.yml | sed -e "s/hosts:\ \[\"localhost:9200\"\]/hosts:\ \[\"elasticsearch:9200\"\]/g" > $BEATNAME-docker.yml
