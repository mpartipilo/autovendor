#!/bin/sh
# autovendor:begin — do not edit this block
# Installed by autovendor — https://github.com/mpartipilo/autovendor
if command -v autovendor >/dev/null 2>&1; then
  autovendor run post-checkout "$@"
else
  go run github.com/mpartipilo/autovendor@latest run post-checkout "$@"
fi
# autovendor:end
