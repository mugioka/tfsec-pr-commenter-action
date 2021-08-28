#!/bin/bash

set -x

if [ -n "${GITHUB_WORKSPACE}" ]; then
  cd "${GITHUB_WORKSPACE}" || exit
fi
tfsec --format=json "${INPUT_WORKING_DIRECTORY}" 2>/dev/null >results.json; then
echo "tfsec violations were identified, running commenter..."
commenter
