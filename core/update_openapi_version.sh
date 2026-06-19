#!/bin/bash

VERSION=$(git describe --tags)

for file in api/*/openapi.yaml; do
   case "$file" in
      *evidence/openapi.yaml)
        TITLE="Evidence Store API"
        ;;
    esac

  if [[ -f "$file" ]]; then
    if [[ "$OSTYPE" == "darwin"* ]]; then
      sed -i '' "s|version: .*|version: $VERSION|" "$file"
       [[ -n $TITLE ]] && sed -i '' "s|title: .*|title: $TITLE|" "$file"
    else
      sed -i "s|version: .*|version: $VERSION|" "$file"
      [[ -n $TITLE ]] && sed -i "s|title: .*|title: $TITLE|" "$file"
    fi
    echo "Version of $file set to $VERSION."
     [[ -n $TITLE ]] && echo "Title of $file set to $TITLE."
  fi
done