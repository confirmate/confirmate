#!/bin/bash

VERSION=$(git describe --tags)

for file in api/*/openapi.yaml; do
  if [[ -f "$file" ]]; then
    if [[ "$OSTYPE" == "darwin"* ]]; then
      sed -i '' "s|version: .*|version: $VERSION|" "$file"
      if [[ "$file" == *"evidence/openapi.yaml" ]]; then
        TITLE="Evidence Store API"
        sed -i '' "s|title: .*|title: $TITLE|" "$file"
        echo "Title of $file set to $TITLE."
      fi
    else
      sed -i "s|version: .*|version: $VERSION|" "$file"
      if [[ "$file" == *"evidence/openapi.yaml" ]]; then
        TITLE="Evidence Store API"
        sed -i "s|title: .*|title: $TITLE|" "$file"
        echo "Title of $file set to $TITLE."
      fi
    fi
    echo "Version of $file set to $VERSION."
  fi
done