#!/bin/bash

set -e

THIS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

cd "${THIS_DIR}"

rm -rf _test
mkdir _test
pushd _test || exit

git init src
echo "creating dummy data..."

pushd src || exit
"${THIS_DIR}/tools/create_dummy_data.py" \
    --user-name="Mr. Private" \
    --email=private@example.com \
    --include-other-users \
    > "${THIS_DIR}/_test/create_dummy_data.log"
popd

git init dest

popd

go run ./cmd/gh-activity-importer \
    --source-repo=_test/src \
    --dest-repo=_test/dest \
    --source-author-email=private@example.com \
    --dest-author-name="Mrs. Public" \
    --dest-author-email=public@example.com

echo "all done. Take a look at _test/src and _test/dest and see what your think"
