#!/bin/bash -e

#function that makes the script exit, if any command fails
exit_if_failed () {
if [ $? -ne 0 ]
then
  exit 1
fi
}

echo "Check arch and export GOROOT & GOPATH"
uname -m
export GOROOT=/usr/local/go
export PATH=$GOPATH/bin:$GOROOT/bin:$PATH
echo ""

echo "Check go version"
go version
exit_if_failed
echo ""

echo "remove existing .steampipe install dir(if any)"
rm -rf ~/.steampipe

echo "Checkout to cloned steampipe repo"
cd steampipe
pwd
echo ""

echo "git reset"
git reset
exit_if_failed
echo ""

echo "git restore all changed files(if any)"
git restore .
exit_if_failed
echo ""

echo "git pull origin main"
git checkout main
git pull origin main
exit_if_failed
echo ""

echo "delete all existing local branches"
git branch | grep -v "main" | xargs git branch -D
exit_if_failed
echo ""

echo "git fetch"
git fetch
exit_if_failed
echo ""

echo "git checkout <branch>"
input=$1
echo $input
git checkout $input
git branch --list
exit_if_failed
echo ""

echo "build steampipe and set PATH"
go build -o ~/bin/steampipe
exit_if_failed
export PATH=$PATH:/home/ubuntu/bin
steampipe -v
exit_if_failed
echo ""

echo "install steampipe and test pre-requisites"
steampipe service start
steampipe plugin install chaos chaosdynamic --progress=false
steampipe service stop
exit_if_failed
echo ""

echo "run acceptance tests"
./tests/acceptance/run.sh
exit_if_failed
echo ""

echo "Hallelujah!"
exit 0
