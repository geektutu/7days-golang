#!/bin/bash
set -eou pipefail

cur=$PWD
for item in "$cur"/day*/
do
    echo "$item"
    cd "$item"
    go test geeorm/... 2>&1 | grep -v warning
done