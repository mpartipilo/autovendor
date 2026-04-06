#!/bin/sh
# autovendor:begin — do not edit this block
# Installed by autovendor — https://github.com/mpartipilo/autovendor
if command -v autovendor >/dev/null 2>&1; then
  autovendor run post-rewrite "$@"
else
  go run github.com/mpartipilo/autovendor@latest run post-rewrite "$@"
fi
# autovendor:end
