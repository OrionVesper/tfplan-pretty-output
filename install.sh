#!/bin/bash

set -e

echo "🔧 Installing tfplan-pretty-output..."

# Install using Go (this builds + places binary in GOPATH/bin)
echo "📦 Running go install..."
go mod tidy
go install .

INSTALL_DIR="$(go env GOPATH)/bin"
BIN_NAME="tfplan-pretty-output"

echo "📁 Installing to: $INSTALL_DIR"

echo "✅ Installed successfully!"

echo ""
echo "👉 Try:"
echo "   tfplan-pretty-output help"
