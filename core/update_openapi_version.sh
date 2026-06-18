#!/bin/bash

VERSION=$(git describe --tags)

for file in api/*/openapi.yaml; do
  if [[ -f "$file" ]]; then
    # empty title if no version is available
    TITLE=""

    case "$file" in
      *evidence/openapi.yaml)
        TITLE="Evidence Store API"
        ;;
      *assessment/openapi.yaml)
        TITLE="Assessment API"
        ;;
      *orchestrator/openapi.yaml)
        TITLE="Orchestrator API"
        ;;
      *evaluation/openapi.yaml)
        TITLE="Evaluation API"
        ;;
    esac

    if [[ "$OSTYPE" == "darwin"* ]]; then
      sed -i '' "s|version: .*|version: $VERSION|" "$file"
      [[ -n "$TITLE" ]] && sed -i '' "s|title: .*|title: \"$TITLE\"|" "$file"
    else
      sed -i "s|version: .*|version: $VERSION|" "$file"
      [[ -n "$TITLE" ]] && sed -i "s|title: .*|title: \"$TITLE\"|" "$file"
    fi
    echo "Version of $file set to $VERSION."
    [[ -n "$TITLE" ]] && echo "Title of $file set to $TITLE."
  fi
done