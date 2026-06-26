#!/bin/bash
# scripts/validate_migrations.sh
# Checks that every .up.sql has a corresponding .down.sql and versions are sequential.
set -e

MIGRATION_DIR="migrations"
errors=0

if [ ! -d "$MIGRATION_DIR" ]; then
    echo "ERROR: Directory $MIGRATION_DIR does not exist."
    exit 1
fi

# Check that every up file has a down file
for up_file in "$MIGRATION_DIR"/*.up.sql; do
    [ -e "$up_file" ] || continue
    down_file="${up_file/.up.sql/.down.sql}"
    if [ ! -f "$down_file" ]; then
        echo "ERROR: Missing down file for $up_file"
        errors=$((errors + 1))
    fi
done

# Check that every down file has an up file
for down_file in "$MIGRATION_DIR"/*.down.sql; do
    [ -e "$down_file" ] || continue
    up_file="${down_file/.down.sql/.up.sql}"
    if [ ! -f "$up_file" ]; then
        echo "ERROR: Missing up file for $down_file"
        errors=$((errors + 1))
    fi
done

# Check sequential numbering
expected=1
for up_file in $(ls "$MIGRATION_DIR"/*.up.sql 2>/dev/null | sort); do
    filename=$(basename "$up_file")
    version_str=$(echo "$filename" | cut -d'_' -f1)
    # Remove leading zeros (base 10)
    version=$((10#$version_str))
    if [ "$version" -ne "$expected" ]; then
        echo "ERROR: Expected version $(printf "%06d" $expected), but found $version_str in $filename"
        errors=$((errors + 1))
    fi
    expected=$((expected + 1))
done

if [ $errors -gt 0 ]; then
    echo "Migration validation failed ($errors error(s)). Blocking deployment."
    exit 1
fi

echo "Migration validation passed."
