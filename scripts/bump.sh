#!/bin/bash

# ==============================================================================
# El Bulk - Version Bump Script
# Usage: ./scripts/bump.sh [frontend|backend|all] [patch|minor|major]
# ==============================================================================

set -e

MANIFEST=".versions"

# Ensure manifesto exists
if [ ! -f "$MANIFEST" ]; then
    echo "FRONTEND_VERSION=0.1.0" > "$MANIFEST"
    echo "BACKEND_VERSION=0.1.0" >> "$MANIFEST"
fi

# Load current versions
source "$MANIFEST"

TARGET=$1
BUMP_TYPE=$2

if [[ -z "$TARGET" || -z "$BUMP_TYPE" ]]; then
    echo "Usage: ./scripts/bump.sh [frontend|backend|all] [patch|minor|major]"
    exit 1
fi

bump_version() {
    local current=$1
    local type=$2
    local parts=( ${current//./ } )
    local major=${parts[0]}
    local minor=${parts[1]}
    local patch=${parts[2]}

    case "$type" in
        major) major=$((major + 1)); minor=0; patch=0 ;;
        minor) minor=$((minor + 1)); patch=0 ;;
        patch) patch=$((patch + 1)) ;;
        *) echo "Invalid bump type: $type"; exit 1 ;;
    esac

    echo "$major.$minor.$patch"
}

if [[ "$TARGET" == "frontend" || "$TARGET" == "all" ]]; then
    NEW_FRONT=$(bump_version "$FRONTEND_VERSION" "$BUMP_TYPE")
    echo "Bumping Frontend: $FRONTEND_VERSION -> $NEW_FRONT"
    sed -i "s/FRONTEND_VERSION=.*/FRONTEND_VERSION=$NEW_FRONT/" "$MANIFEST"
    
    # Sync to frontend/package.json
    if [ -f "frontend/package.json" ]; then
        cd frontend
        npm version "$BUMP_TYPE" --no-git-tag-version
        cd ..
    fi

    # Sync to frontend/.env.local for local dev
    if [ -f "frontend/.env.local" ]; then
        if grep -q "NEXT_PUBLIC_APP_VERSION" "frontend/.env.local"; then
            sed -i "s/NEXT_PUBLIC_APP_VERSION=.*/NEXT_PUBLIC_APP_VERSION=$NEW_FRONT/" "frontend/.env.local"
        else
            echo "NEXT_PUBLIC_APP_VERSION=$NEW_FRONT" >> "frontend/.env.local"
        fi
    fi
fi

if [[ "$TARGET" == "backend" || "$TARGET" == "all" ]]; then
    NEW_BACK=$(bump_version "$BACKEND_VERSION" "$BUMP_TYPE")
    echo "Bumping Backend: $BACKEND_VERSION -> $NEW_BACK"
    sed -i "s/BACKEND_VERSION=.*/BACKEND_VERSION=$NEW_BACK/" "$MANIFEST"

    # Sync to root .env for local backend dev
    if [ -f ".env" ]; then
        if grep -q "APP_VERSION" ".env"; then
            sed -i "s/APP_VERSION=.*/APP_VERSION=$NEW_BACK/" ".env"
        else
            echo "APP_VERSION=$NEW_BACK" >> ".env"
        fi
    fi
fi

echo "✅ Successfully updated $MANIFEST"
echo "Current state:"
cat "$MANIFEST"
