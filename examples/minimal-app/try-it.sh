#!/bin/bash
# Quick demo script for GraphFS CLI
# This script demonstrates all core functionality

set -e

echo "🚀 GraphFS CLI Demo"
echo "==================="
echo ""

# Check if graphfs is installed
if ! command -v graphfs &> /dev/null; then
    echo "❌ graphfs command not found"
    echo "Installing graphfs..."
    cd ../..
    go install ./cmd/graphfs
    cd examples/minimal-app
    echo "✅ graphfs installed"
fi

echo "📍 Current directory: $(pwd)"
echo ""

# Initialize GraphFS
echo "1️⃣  Initializing GraphFS..."
graphfs init
echo ""

# Scan with validation and stats
echo "2️⃣  Scanning codebase..."
graphfs scan --validate --stats
echo ""

# Query 1: List all modules
echo "3️⃣  Query: List all modules"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━"
graphfs query 'SELECT ?module ?description WHERE {
  ?module <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <https://schema.codedoc.org/Module> .
  ?module <https://schema.codedoc.org/description> ?description .
}'
echo ""

# Query 2: Service modules
echo "4️⃣  Query: List service layer modules"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
graphfs query --file queries/list-service-modules.sparql
echo ""

# Query 3: Dependencies
echo "5️⃣  Query: Find module dependencies"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
graphfs query --file queries/find-dependencies.sparql | head -20
echo ""

# Query 4: JSON output
echo "6️⃣  Query: JSON output format"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
graphfs query --file queries/list-service-modules.sparql --format json
echo ""

# Export graph
echo "7️⃣  Exporting graph to JSON..."
graphfs scan --output /tmp/minimal-app-graph.json
echo "✅ Graph exported to /tmp/minimal-app-graph.json"
echo ""

echo "✨ Demo complete!"
echo ""
echo "📚 Learn more:"
echo "   - Query examples: examples/query-examples.md"
echo "   - Pre-built queries: queries/"
echo "   - Full documentation: ../../README.md"
