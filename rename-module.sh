#!/usr/bin/bash

find . -type f -not -path '*/\.git/*' -not -name 'rename-module.sh' -exec sed -i 's+github.com/zapling/go-service+github.com/zapling/something-else+g' '{}' +
