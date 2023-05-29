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

echo "git fetch"
git fetch
exit_if_failed
echo ""

echo "git pull origin main"
git checkout main
git pull origin main
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

echo "run acceptance tests"
declare -a arr=("migration" "service_and_plugin" "search_path" "chaos_and_query" "dynamic_schema" "dynamic_aggregators" "cache" "mod_install" "mod" "mod_require" "check" "performance" "workspace" introspection "cloud" "snapshot" "dashboard" "dashboard_parsing_validation" "schema_cloning" "exit_codes")
declare -i failure_count=0
# run test suite
for i in "${arr[@]}"
do
  echo ""
  echo ">>>>> running $i.bats"
  ./tests/acceptance/run-local.sh $i.bats
  failure_count+=$?
done
echo ""

# check if all tests passed
echo $failure_count
if [[ $failure_count -eq 0 ]]; then
  echo "test run successful"
  exit 0
else
  echo "test run failed"
  exit 1
fi

echo "Hallelujah!"
exit 0
