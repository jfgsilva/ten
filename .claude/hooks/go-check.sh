#!/bin/sh
# Runs gofmt + go test when a .go file is edited/written.
# Called by Claude Code PostToolUse hooks; file path is in CLAUDE_TOOL_INPUT JSON.

FILE=$(echo "$CLAUDE_TOOL_INPUT" | python3 -c "import sys,json; print(json.load(sys.stdin).get('file_path',''))" 2>/dev/null)

case "$FILE" in
  *.go)
    echo "go-check: fmt $FILE"
    gofmt -w "$FILE"
    echo "go-check: tidy"
    cd /Users/jsilva/ten && go mod tidy
    echo "go-check: test ./..."
    go test ./...
    echo "go-check: gosec ./..."
    gosec ./... 2>/dev/null || true
    echo "go-check: govulncheck ./..."
    govulncheck ./... || true
    ;;
  */go.mod|*/go.sum)
    echo "go-check: tidy"
    cd /Users/jsilva/ten && go mod tidy
    ;;
esac
