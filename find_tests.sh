#!/bin/bash
cd /home/user/teesql/parser/testdata
for dir in */; do
    d="${dir%/}"
    if [ -f "$d/ast.json" ] && grep -q '"skip": true' "$d/metadata.json" 2>/dev/null; then
        echo "$d"
    fi
done
