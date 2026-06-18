#!/bin/bash

VERSION=$(git describe --tags)

for file in api/*/openapi.yaml; do
  if [[ -f "$file" ]]; then
    if [[ "$OSTYPE" == "darwin"* ]]; then
      sed -i '' "s|version: .*|version: $VERSION|" "$file"
    else
      sed -i "s|version: .*|version: $VERSION|" "$file"
    fi
    echo "Version in $file auf $VERSION gesetzt."
  fi
done