#!/usr/bin/bash

modulePath=$1
if [[ "$modulePath" == "" ]]; then
    echo "You must provide a module path as argument"
    printf "\nExample:\n"
    echo "./rename-module.sh github.com/zapling/some-go-service"
    exit 1
fi

find . -type f -not -path '*/\.git/*' -not -name 'rename-module.sh' -exec sed -i 's+github.com/zapling/go-service+github.com/zapling/something-else+g' '{}' +
