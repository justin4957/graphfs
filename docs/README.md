---
author: GraphFS Team
version: 1.0.0
title: graphfs Documentation
project: graphfs
---

# graphfs Documentation

## Overview

This documentation covers **48 modules** in the graphfs project.

### Statistics

- **Total Modules:** 48
- **Layers:** 17
  - parser: 1 modules
  - query: 3 modules
  - repl: 3 modules
  - schema: 1 modules
  - service: 2 modules
  - storage: 2 modules
  - graph: 3 modules
  - data: 1 modules
  - rules: 5 modules
  - server: 10 modules
  - model: 1 modules
  - analysis: 7 modules
  - visualization: 1 modules
  - unknown: 1 modules
  - utility: 3 modules
  - cache: 1 modules
  - scanner: 3 modules


## Table of Contents

- [examples/minimal-app/main.go](#module-examples-minimal-app-main-go)
- [examples/minimal-app/models/user.go](#module-examples-minimal-app-models-user-go)
- [examples/minimal-app/services/auth.go](#module-examples-minimal-app-services-auth-go)
- [examples/minimal-app/services/user.go](#module-examples-minimal-app-services-user-go)
- [examples/minimal-app/utils/crypto.go](#module-examples-minimal-app-utils-crypto-go)
- [examples/minimal-app/utils/logger.go](#module-examples-minimal-app-utils-logger-go)
- [examples/minimal-app/utils/validator.go](#module-examples-minimal-app-utils-validator-go)
- [internal/store/store.go](#module-internal-store-store-go)
- [internal/store/triple.go](#module-internal-store-triple-go)
- [pkg/analysis/cleanup.go](#module-pkg-analysis-cleanup-go)
- [pkg/analysis/coverage.go](#module-pkg-analysis-coverage-go)
- [pkg/analysis/deadcode.go](#module-pkg-analysis-deadcode-go)
- [pkg/analysis/graph_algorithms.go](#module-pkg-analysis-graph-algorithms-go)
- [pkg/analysis/impact.go](#module-pkg-analysis-impact-go)
- [pkg/analysis/security.go](#module-pkg-analysis-security-go)
- [pkg/analysis/zones.go](#module-pkg-analysis-zones-go)
- [pkg/cache/cache.go](#module-pkg-cache-cache-go)
- [pkg/graph/graph.go](#module-pkg-graph-graph-go)
- [pkg/graph/module.go](#module-pkg-graph-module-go)
- [pkg/graph/validator.go](#module-pkg-graph-validator-go)
- [pkg/parser/parser.go](#module-pkg-parser-parser-go)
- [pkg/parser/triple.go](#module-pkg-parser-triple-go)
- [pkg/query/executor.go](#module-pkg-query-executor-go)
- [pkg/query/parser.go](#module-pkg-query-parser-go)
- [pkg/query/query.go](#module-pkg-query-query-go)
- [pkg/repl/commands.go](#module-pkg-repl-commands-go)
- [pkg/repl/formatter.go](#module-pkg-repl-formatter-go)
- [pkg/repl/repl.go](#module-pkg-repl-repl-go)
- [pkg/rules/engine.go](#module-pkg-rules-engine-go)
- [pkg/rules/evaluator.go](#module-pkg-rules-evaluator-go)
- [pkg/rules/parser.go](#module-pkg-rules-parser-go)
- [pkg/rules/reporter.go](#module-pkg-rules-reporter-go)
- [pkg/rules/rule.go](#module-pkg-rules-rule-go)
- [pkg/scanner/ignore.go](#module-pkg-scanner-ignore-go)
- [pkg/scanner/language.go](#module-pkg-scanner-language-go)
- [pkg/scanner/scanner.go](#module-pkg-scanner-scanner-go)
- [pkg/schema/graphql/generator.go](#module-pkg-schema-graphql-generator-go)
- [pkg/server/cache_middleware.go](#module-pkg-server-cache-middleware-go)
- [pkg/server/graphql/resolvers.go](#module-pkg-server-graphql-resolvers-go)
- [pkg/server/graphql/schema.go](#module-pkg-server-graphql-schema-go)
- [pkg/server/graphql/server.go](#module-pkg-server-graphql-server-go)
- [pkg/server/rest/analysis.go](#module-pkg-server-rest-analysis-go)
- [pkg/server/rest/handler.go](#module-pkg-server-rest-handler-go)
- [pkg/server/rest/modules.go](#module-pkg-server-rest-modules-go)
- [pkg/server/rest/tags.go](#module-pkg-server-rest-tags-go)
- [pkg/server/server.go](#module-pkg-server-server-go)
- [pkg/server/sparql_handler.go](#module-pkg-server-sparql-handler-go)
- [pkg/viz/dot.go](#module-pkg-viz-dot-go)

## Module: examples/minimal-app/main.go

**Tags:** entrypoint, example, demo  
**Language:** go  

### Description
Minimal GraphFS example application entry point

### Dependencies

- [examples/minimal-app/services/user.go](#module-examples-minimal-app-services-user-go) - User management service
- [examples/minimal-app/services/auth.go](#module-examples-minimal-app-services-auth-go) - Authentication and authorization service
- [examples/minimal-app/utils/logger.go](#module-examples-minimal-app-utils-logger-go) - Logging utility for structured logging
- [examples/minimal-app/models/user.go](#module-examples-minimal-app-models-user-go) - User data model and validation

### Exports

- `#main`


## Module: examples/minimal-app/models/user.go

**Layer:** model  
**Tags:** model, data, user, domain  
**Language:** go  

### Description
User data model and validation

### Dependencies

- [examples/minimal-app/utils/validator.go](#module-examples-minimal-app-utils-validator-go) - Data validation utilities

### Dependents

- [examples/minimal-app/services/user.go](#module-examples-minimal-app-services-user-go) - User management service
- [examples/minimal-app/services/auth.go](#module-examples-minimal-app-services-auth-go) - Authentication and authorization service
- [examples/minimal-app/main.go](#module-examples-minimal-app-main-go) - Minimal GraphFS example application entry point

### Exports

- `#User`
- `#UserRole`
- `#ValidateUser`


## Module: examples/minimal-app/services/auth.go

**Layer:** service  
**Tags:** security, authentication, authorization, service  
**Language:** go  

### Description
Authentication and authorization service

### Dependencies

- [examples/minimal-app/utils/crypto.go](#module-examples-minimal-app-utils-crypto-go) - Cryptographic utilities for secure operations
- [examples/minimal-app/utils/logger.go](#module-examples-minimal-app-utils-logger-go) - Logging utility for structured logging
- [examples/minimal-app/models/user.go](#module-examples-minimal-app-models-user-go) - User data model and validation

### Dependents

- [examples/minimal-app/main.go](#module-examples-minimal-app-main-go) - Minimal GraphFS example application entry point
- [examples/minimal-app/services/user.go](#module-examples-minimal-app-services-user-go) - User management service

### Exports

- `#AuthService`
- `#NewAuthService`


## Module: examples/minimal-app/services/user.go

**Layer:** service  
**Tags:** service, user, crud, business-logic  
**Language:** go  

### Description
User management service

### Dependencies

- [examples/minimal-app/services/auth.go](#module-examples-minimal-app-services-auth-go) - Authentication and authorization service
- [examples/minimal-app/models/user.go](#module-examples-minimal-app-models-user-go) - User data model and validation
- [examples/minimal-app/utils/logger.go](#module-examples-minimal-app-utils-logger-go) - Logging utility for structured logging

### Dependents

- [examples/minimal-app/main.go](#module-examples-minimal-app-main-go) - Minimal GraphFS example application entry point

### Exports

- `#UserService`
- `#NewUserService`


## Module: examples/minimal-app/utils/crypto.go

**Layer:** utility  
**Tags:** utility, security, cryptography  
**Language:** go  

### Description
Cryptographic utilities for secure operations

### Dependents

- [examples/minimal-app/services/auth.go](#module-examples-minimal-app-services-auth-go) - Authentication and authorization service

### Exports

- `#CryptoHelper`
- `#NewCryptoHelper`


## Module: examples/minimal-app/utils/logger.go

**Layer:** utility  
**Tags:** utility, logging, observability  
**Language:** go  

### Description
Logging utility for structured logging

### Dependents

- [examples/minimal-app/services/user.go](#module-examples-minimal-app-services-user-go) - User management service
- [examples/minimal-app/services/auth.go](#module-examples-minimal-app-services-auth-go) - Authentication and authorization service
- [examples/minimal-app/main.go](#module-examples-minimal-app-main-go) - Minimal GraphFS example application entry point

### Exports

- `#Logger`
- `#NewLogger`
- `#LogLevel`


## Module: examples/minimal-app/utils/validator.go

**Layer:** utility  
**Tags:** utility, validation, data-quality  
**Language:** go  

### Description
Data validation utilities

### Dependents

- [examples/minimal-app/models/user.go](#module-examples-minimal-app-models-user-go) - User data model and validation

### Exports

- `#ValidateEmail`
- `#ValidationError`
- `#NewValidationError`
- `#PermissionError`
- `#NewPermissionError`


## Module: internal/store/store.go

**Layer:** storage  
**Tags:** store, rdf, triplestore, in-memory  
**Language:** go  

### Description
In-memory RDF triple store with multiple indexes

### Dependencies

- [internal/store/triple.go](#module-internal-store-triple-go) - Triple data structure for RDF storage

### Dependents

- [pkg/graph/graph.go](#module-pkg-graph-graph-go) - Graph data structures for knowledge graph representation
- [pkg/query/executor.go](#module-pkg-query-executor-go) - SPARQL query executor
- [pkg/schema/graphql/generator.go](#module-pkg-schema-graphql-generator-go) - GraphQL schema generator for GraphFS knowledge graphs

### Exports

- `#TripleStore`
- `#NewTripleStore`


## Module: internal/store/triple.go

**Layer:** storage  
**Tags:** store, rdf, data-structure  
**Language:** go  

### Description
Triple data structure for RDF storage

### Dependents

- [internal/store/store.go](#module-internal-store-store-go) - In-memory RDF triple store with multiple indexes

### Exports

- `#Triple`


## Module: pkg/analysis/cleanup.go

**Layer:** analysis  
**Tags:** analysis, cleanup, recommendations  
**Language:** go  

### Description
Cleanup recommendations and script generation for dead code removal

### Dependencies

- [pkg/analysis/deadcode.go](#module-pkg-analysis-deadcode-go) - Dead code detection for identifying unreferenced modules and symbols
- [pkg/graph/graph.go](#module-pkg-graph-graph-go) - Graph data structures for knowledge graph representation

### Exports

- `#CleanupPlan`
- `#CleanupAction`
- `#GenerateCleanupPlan`
- `#GenerateScript`


## Module: pkg/analysis/coverage.go

**Layer:** analysis  
**Tags:** analysis, coverage, usage  
**Language:** go  

### Description
Usage coverage analysis for tracking module and symbol usage

### Dependencies

- [pkg/graph/graph.go](#module-pkg-graph-graph-go) - Graph data structures for knowledge graph representation
- [pkg/analysis/deadcode.go](#module-pkg-analysis-deadcode-go) - Dead code detection for identifying unreferenced modules and symbols

### Exports

- `#CoverageAnalysis`
- `#ModuleCoverage`
- `#AnalyzeCoverage`


## Module: pkg/analysis/deadcode.go

**Layer:** analysis  
**Tags:** analysis, dead-code, unused  
**Language:** go  

### Description
Dead code detection for identifying unreferenced modules and symbols

### Dependencies

- [pkg/graph/graph.go](#module-pkg-graph-graph-go) - Graph data structures for knowledge graph representation

### Dependents

- [pkg/analysis/cleanup.go](#module-pkg-analysis-cleanup-go) - Cleanup recommendations and script generation for dead code removal
- [pkg/analysis/coverage.go](#module-pkg-analysis-coverage-go) - Usage coverage analysis for tracking module and symbol usage

### Exports

- `#DeadCodeAnalysis`
- `#DeadModule`
- `#DetectDeadCode`
- `#DeadCodeOptions`


## Module: pkg/analysis/graph_algorithms.go

**Layer:** analysis  
**Tags:** analysis, graph-algorithms, dependencies, topology  
**Language:** go  

### Description
Dependency graph algorithms for GraphFS analysis

### Dependencies

- [pkg/graph/graph.go](#module-pkg-graph-graph-go) - Graph data structures for knowledge graph representation

### Dependents

- [pkg/analysis/impact.go](#module-pkg-analysis-impact-go) - Impact analysis engine for GraphFS

### Exports

- `#TopologicalSort`
- `#ShortestPath`
- `#StronglyConnectedComponents`
- `#TransitiveDependencies`


## Module: pkg/analysis/impact.go

**Layer:** analysis  
**Tags:** analysis, impact-analysis, refactoring, risk-assessment  
**Language:** go  

### Description
Impact analysis engine for GraphFS

### Dependencies

- [pkg/graph/graph.go](#module-pkg-graph-graph-go) - Graph data structures for knowledge graph representation
- [pkg/analysis/graph_algorithms.go](#module-pkg-analysis-graph-algorithms-go) - Dependency graph algorithms for GraphFS analysis

### Dependents

- [pkg/viz/dot.go](#module-pkg-viz-dot-go) - GraphViz DOT format generation for dependency visualization

### Exports

- `#ImpactAnalysis`
- `#ImpactResult`
- `#RiskLevel`
- `#AnalyzeImpact`


## Module: pkg/analysis/security.go

**Layer:** analysis  
**Tags:** analysis, security, boundaries  
**Language:** go  

### Description
Security boundary analysis for detecting and enforcing security zones

### Dependencies

- [pkg/graph/graph.go](#module-pkg-graph-graph-go) - Graph data structures for knowledge graph representation
- [pkg/analysis/zones.go](#module-pkg-analysis-zones-go) - Security zone classification and detection

### Dependents

- [pkg/viz/dot.go](#module-pkg-viz-dot-go) - GraphViz DOT format generation for dependency visualization

### Exports

- `#SecurityAnalysis`
- `#SecurityBoundary`
- `#SecurityViolation`
- `#AnalyzeSecurity`


## Module: pkg/analysis/zones.go

**Layer:** analysis  
**Tags:** analysis, security, zones  
**Language:** go  

### Description
Security zone classification and detection

### Dependencies

- [pkg/graph/graph.go](#module-pkg-graph-graph-go) - Graph data structures for knowledge graph representation

### Dependents

- [pkg/analysis/security.go](#module-pkg-analysis-security-go) - Security boundary analysis for detecting and enforcing security zones

### Exports

- `#SecurityZone`
- `#ZoneClassifier`
- `#ClassifyZones`


## Module: pkg/cache/cache.go

**Layer:** cache  
**Tags:** cache, performance, lru  
**Language:** go  

### Description
Query result caching with LRU eviction

### Dependents

- [pkg/server/cache_middleware.go](#module-pkg-server-cache-middleware-go) - HTTP response caching middleware

### Exports

- `#Cache`
- `#NewCache`
- `#CacheEntry`
- `#Stats`


## Module: pkg/graph/graph.go

**Layer:** graph  
**Tags:** graph, knowledge-graph, data-structure  
**Language:** go  

### Description
Graph data structures for knowledge graph representation

### Dependencies

- [pkg/graph/module.go](#module-pkg-graph-module-go) - Module data structure for representing code modules
- [internal/store/store.go](#module-internal-store-store-go) - In-memory RDF triple store with multiple indexes

### Dependents

- [pkg/server/rest/handler.go](#module-pkg-server-rest-handler-go) - REST API handler for GraphFS
- [pkg/analysis/security.go](#module-pkg-analysis-security-go) - Security boundary analysis for detecting and enforcing security zones
- [pkg/server/rest/tags.go](#module-pkg-server-rest-tags-go) - Tag and export endpoints for REST API
- [pkg/analysis/graph_algorithms.go](#module-pkg-analysis-graph-algorithms-go) - Dependency graph algorithms for GraphFS analysis
- [pkg/analysis/coverage.go](#module-pkg-analysis-coverage-go) - Usage coverage analysis for tracking module and symbol usage
- [pkg/analysis/zones.go](#module-pkg-analysis-zones-go) - Security zone classification and detection
- [pkg/repl/repl.go](#module-pkg-repl-repl-go) - Interactive REPL for GraphFS queries
- [pkg/rules/rule.go](#module-pkg-rules-rule-go) - Rule data structures and types for architecture validation
- [pkg/schema/graphql/generator.go](#module-pkg-schema-graphql-generator-go) - GraphQL schema generator for GraphFS knowledge graphs
- [pkg/server/graphql/resolvers.go](#module-pkg-server-graphql-resolvers-go) - GraphQL resolvers for GraphFS
- [pkg/server/graphql/schema.go](#module-pkg-server-graphql-schema-go) - GraphQL schema definition for GraphFS
- [pkg/server/rest/modules.go](#module-pkg-server-rest-modules-go) - Module endpoints for REST API
- [pkg/analysis/deadcode.go](#module-pkg-analysis-deadcode-go) - Dead code detection for identifying unreferenced modules and symbols
- [pkg/analysis/impact.go](#module-pkg-analysis-impact-go) - Impact analysis engine for GraphFS
- [pkg/graph/validator.go](#module-pkg-graph-validator-go) - Graph validation implementation
- [pkg/rules/evaluator.go](#module-pkg-rules-evaluator-go) - SPARQL-based rule evaluator for executing rules
- [pkg/server/rest/analysis.go](#module-pkg-server-rest-analysis-go) - Analysis endpoints for REST API
- [pkg/viz/dot.go](#module-pkg-viz-dot-go) - GraphViz DOT format generation for dependency visualization
- [pkg/rules/engine.go](#module-pkg-rules-engine-go) - Rule engine for validating architectural constraints
- [pkg/analysis/cleanup.go](#module-pkg-analysis-cleanup-go) - Cleanup recommendations and script generation for dead code removal
- [pkg/graph/module.go](#module-pkg-graph-module-go) - Module data structure for representing code modules
- [pkg/server/graphql/server.go](#module-pkg-server-graphql-server-go) - GraphQL HTTP server for GraphFS

### Exports

- `#Graph`
- `#GraphStats`


## Module: pkg/graph/module.go

**Layer:** graph  
**Tags:** graph, module, data-structure  
**Language:** go  

### Description
Module data structure for representing code modules

### Dependencies

- [pkg/graph/graph.go](#module-pkg-graph-graph-go) - Graph data structures for knowledge graph representation

### Dependents

- [pkg/graph/graph.go](#module-pkg-graph-graph-go) - Graph data structures for knowledge graph representation
- [pkg/graph/validator.go](#module-pkg-graph-validator-go) - Graph validation implementation

### Exports

- `#Module`


## Module: pkg/graph/validator.go

**Layer:** graph  
**Tags:** graph, validation, consistency  
**Language:** go  

### Description
Graph validation implementation

### Dependencies

- [pkg/graph/graph.go](#module-pkg-graph-graph-go) - Graph data structures for knowledge graph representation
- [pkg/graph/module.go](#module-pkg-graph-module-go) - Module data structure for representing code modules

### Exports

- `#Validator`
- `#ValidationResult`
- `#ValidationError`
- `#ValidationWarning`
- `#NewValidator`


## Module: pkg/parser/parser.go

**Layer:** parser  
**Tags:** parser, rdf, turtle, linkeddoc  
**Language:** go  

### Description
LinkedDoc+RDF parser implementation

### Dependencies

- [pkg/parser/triple.go](#module-pkg-parser-triple-go) - RDF triple data structure

### Dependents

- [pkg/scanner/scanner.go](#module-pkg-scanner-scanner-go) - Filesystem scanner for GraphFS

### Exports

- `#Parser`
- `#NewParser`
- `#ParseError`


## Module: pkg/parser/triple.go

**Layer:** data  
**Tags:** parser, rdf, data-structure  
**Language:** go  

### Description
RDF triple data structure

### Dependents

- [pkg/parser/parser.go](#module-pkg-parser-parser-go) - LinkedDoc+RDF parser implementation

### Exports

- `#Triple`
- `#TripleObject`
- `#LiteralObject`
- `#URIObject`
- `#BlankNodeObject`


## Module: pkg/query/executor.go

**Layer:** query  
**Tags:** query, sparql, executor  
**Language:** go  

### Description
SPARQL query executor

### Dependencies

- [pkg/query/query.go](#module-pkg-query-query-go) - SPARQL query data structures
- [internal/store/store.go](#module-internal-store-store-go) - In-memory RDF triple store with multiple indexes

### Dependents

- [pkg/repl/repl.go](#module-pkg-repl-repl-go) - Interactive REPL for GraphFS queries
- [pkg/repl/formatter.go](#module-pkg-repl-formatter-go) - Output formatters for REPL results
- [pkg/server/sparql_handler.go](#module-pkg-server-sparql-handler-go) - HTTP handler for SPARQL queries
- [pkg/server/server.go](#module-pkg-server-server-go) - HTTP server for GraphFS query endpoints

### Exports

- `#Executor`
- `#NewExecutor`
- `#QueryResult`


## Module: pkg/query/parser.go

**Layer:** query  
**Tags:** query, sparql, parser  
**Language:** go  

### Description
SPARQL query parser

### Dependencies

- [pkg/query/query.go](#module-pkg-query-query-go) - SPARQL query data structures

### Exports

- `#ParseQuery`


## Module: pkg/query/query.go

**Layer:** query  
**Tags:** query, sparql, data-structure  
**Language:** go  

### Description
SPARQL query data structures

### Dependents

- [pkg/query/parser.go](#module-pkg-query-parser-go) - SPARQL query parser
- [pkg/query/executor.go](#module-pkg-query-executor-go) - SPARQL query executor

### Exports

- `#Query`
- `#SelectQuery`
- `#TriplePattern`
- `#Filter`


## Module: pkg/repl/commands.go

**Layer:** repl  
**Tags:** repl, commands, cli  
**Language:** go  

### Description
REPL command handlers

### Dependencies

- [pkg/repl/repl.go](#module-pkg-repl-repl-go) - Interactive REPL for GraphFS queries


## Module: pkg/repl/formatter.go

**Layer:** repl  
**Tags:** repl, formatter, output  
**Language:** go  

### Description
Output formatters for REPL results

### Dependencies

- [pkg/query/executor.go](#module-pkg-query-executor-go) - SPARQL query executor
- [pkg/repl/repl.go](#module-pkg-repl-repl-go) - Interactive REPL for GraphFS queries


## Module: pkg/repl/repl.go

**Layer:** repl  
**Tags:** repl, interactive, cli, sparql  
**Language:** go  

### Description
Interactive REPL for GraphFS queries

### Dependencies

- [pkg/query/executor.go](#module-pkg-query-executor-go) - SPARQL query executor
- [pkg/graph/graph.go](#module-pkg-graph-graph-go) - Graph data structures for knowledge graph representation

### Dependents

- [pkg/repl/formatter.go](#module-pkg-repl-formatter-go) - Output formatters for REPL results
- [pkg/repl/commands.go](#module-pkg-repl-commands-go) - REPL command handlers

### Exports

- `#REPL`
- `#Config`
- `#New`


## Module: pkg/rules/engine.go

**Layer:** rules  
**Tags:** rules, engine, validation  
**Language:** go  

### Description
Rule engine for validating architectural constraints

### Dependencies

- [pkg/rules/rule.go](#module-pkg-rules-rule-go) - Rule data structures and types for architecture validation
- [pkg/rules/parser.go](#module-pkg-rules-parser-go) - YAML rule parser for loading and validating rule definitions
- [pkg/rules/evaluator.go](#module-pkg-rules-evaluator-go) - SPARQL-based rule evaluator for executing rules
- [pkg/rules/reporter.go](#module-pkg-rules-reporter-go) - Violation reporter for formatting and displaying rule violations
- [pkg/graph/graph.go](#module-pkg-graph-graph-go) - Graph data structures for knowledge graph representation

### Exports

- `#Engine`
- `#ValidateRules`


## Module: pkg/rules/evaluator.go

**Layer:** rules  
**Tags:** rules, evaluator, sparql  
**Language:** go  

### Description
SPARQL-based rule evaluator for executing rules

### Dependencies

- [pkg/rules/rule.go](#module-pkg-rules-rule-go) - Rule data structures and types for architecture validation
- [pkg/graph/graph.go](#module-pkg-graph-graph-go) - Graph data structures for knowledge graph representation

### Dependents

- [pkg/rules/engine.go](#module-pkg-rules-engine-go) - Rule engine for validating architectural constraints

### Exports

- `#Evaluator`
- `#EvaluateRule`


## Module: pkg/rules/parser.go

**Layer:** rules  
**Tags:** rules, parser, yaml  
**Language:** go  

### Description
YAML rule parser for loading and validating rule definitions

### Dependencies

- [pkg/rules/rule.go](#module-pkg-rules-rule-go) - Rule data structures and types for architecture validation

### Dependents

- [pkg/rules/engine.go](#module-pkg-rules-engine-go) - Rule engine for validating architectural constraints

### Exports

- `#Parser`
- `#ParseRules`
- `#ParseRuleSet`


## Module: pkg/rules/reporter.go

**Layer:** rules  
**Tags:** rules, reporter, output  
**Language:** go  

### Description
Violation reporter for formatting and displaying rule violations

### Dependencies

- [pkg/rules/rule.go](#module-pkg-rules-rule-go) - Rule data structures and types for architecture validation

### Dependents

- [pkg/rules/engine.go](#module-pkg-rules-engine-go) - Rule engine for validating architectural constraints

### Exports

- `#Reporter`
- `#FormatText`
- `#FormatJSON`


## Module: pkg/rules/rule.go

**Layer:** rules  
**Tags:** rules, validation, architecture  
**Language:** go  

### Description
Rule data structures and types for architecture validation

### Dependencies

- [pkg/graph/graph.go](#module-pkg-graph-graph-go) - Graph data structures for knowledge graph representation

### Dependents

- [pkg/rules/parser.go](#module-pkg-rules-parser-go) - YAML rule parser for loading and validating rule definitions
- [pkg/rules/evaluator.go](#module-pkg-rules-evaluator-go) - SPARQL-based rule evaluator for executing rules
- [pkg/rules/engine.go](#module-pkg-rules-engine-go) - Rule engine for validating architectural constraints
- [pkg/rules/reporter.go](#module-pkg-rules-reporter-go) - Violation reporter for formatting and displaying rule violations

### Exports

- `#Rule`
- `#Severity`
- `#Violation`
- `#ValidationResult`


## Module: pkg/scanner/ignore.go

**Layer:** scanner  
**Tags:** scanner, ignore-patterns, filtering  
**Language:** go  

### Description
Ignore pattern handling for filesystem scanning

### Dependents

- [pkg/scanner/scanner.go](#module-pkg-scanner-scanner-go) - Filesystem scanner for GraphFS

### Exports

- `#IgnoreMatcher`
- `#NewIgnoreMatcher`
- `#DefaultIgnorePatterns`


## Module: pkg/scanner/language.go

**Layer:** scanner  
**Tags:** scanner, language-detection, utility  
**Language:** go  

### Description
Language detection for source files

### Dependents

- [pkg/scanner/scanner.go](#module-pkg-scanner-scanner-go) - Filesystem scanner for GraphFS

### Exports

- `#Language`
- `#DetectLanguage`
- `#RegisterLanguage`
- `#SupportedLanguages`


## Module: pkg/scanner/scanner.go

**Layer:** scanner  
**Tags:** scanner, filesystem, recursive  
**Language:** go  

### Description
Filesystem scanner for GraphFS

### Dependencies

- [pkg/scanner/language.go](#module-pkg-scanner-language-go) - Language detection for source files
- [pkg/scanner/ignore.go](#module-pkg-scanner-ignore-go) - Ignore pattern handling for filesystem scanning
- [pkg/parser/parser.go](#module-pkg-parser-parser-go) - LinkedDoc+RDF parser implementation

### Exports

- `#Scanner`
- `#NewScanner`
- `#ScanOptions`
- `#ScanResult`
- `#FileInfo`


## Module: pkg/schema/graphql/generator.go

**Layer:** schema  
**Tags:** graphql, schema, generator, sdl  
**Language:** go  

### Description
GraphQL schema generator for GraphFS knowledge graphs

### Dependencies

- [pkg/graph/graph.go](#module-pkg-graph-graph-go) - Graph data structures for knowledge graph representation
- [internal/store/store.go](#module-internal-store-store-go) - In-memory RDF triple store with multiple indexes

### Exports

- `#Generator`
- `#GenerateSchema`
- `#GenerateOptions`


## Module: pkg/server/cache_middleware.go

**Layer:** server  
**Tags:** server, cache, middleware, http  
**Language:** go  

### Description
HTTP response caching middleware

### Dependencies

- [pkg/cache/cache.go](#module-pkg-cache-cache-go) - Query result caching with LRU eviction

### Exports

- `#CacheMiddleware`
- `#responseWriter`


## Module: pkg/server/graphql/resolvers.go

**Layer:** server  
**Tags:** graphql, resolvers, server  
**Language:** go  

### Description
GraphQL resolvers for GraphFS

### Dependencies

- [pkg/graph/graph.go](#module-pkg-graph-graph-go) - Graph data structures for knowledge graph representation
- [pkg/server/graphql/schema.go](#module-pkg-server-graphql-schema-go) - GraphQL schema definition for GraphFS

### Dependents

- [pkg/server/graphql/schema.go](#module-pkg-server-graphql-schema-go) - GraphQL schema definition for GraphFS
- [pkg/server/graphql/server.go](#module-pkg-server-graphql-server-go) - GraphQL HTTP server for GraphFS

### Exports

- `#Resolver`
- `#NewResolver`


## Module: pkg/server/graphql/schema.go

**Layer:** server  
**Tags:** graphql, schema, server  
**Language:** go  

### Description
GraphQL schema definition for GraphFS

### Dependencies

- [pkg/graph/graph.go](#module-pkg-graph-graph-go) - Graph data structures for knowledge graph representation
- [pkg/server/graphql/resolvers.go](#module-pkg-server-graphql-resolvers-go) - GraphQL resolvers for GraphFS

### Dependents

- [pkg/server/graphql/resolvers.go](#module-pkg-server-graphql-resolvers-go) - GraphQL resolvers for GraphFS
- [pkg/server/graphql/server.go](#module-pkg-server-graphql-server-go) - GraphQL HTTP server for GraphFS

### Exports

- `#BuildSchema`
- `#ModuleType`
- `#QueryType`


## Module: pkg/server/graphql/server.go

**Layer:** server  
**Tags:** graphql, server, http  
**Language:** go  

### Description
GraphQL HTTP server for GraphFS

### Dependencies

- [pkg/graph/graph.go](#module-pkg-graph-graph-go) - Graph data structures for knowledge graph representation
- [pkg/server/graphql/schema.go](#module-pkg-server-graphql-schema-go) - GraphQL schema definition for GraphFS
- [pkg/server/graphql/resolvers.go](#module-pkg-server-graphql-resolvers-go) - GraphQL resolvers for GraphFS

### Exports

- `#NewHandler`


## Module: pkg/server/rest/analysis.go

**Layer:** server  
**Tags:** rest, api, analysis  
**Language:** go  

### Description
Analysis endpoints for REST API

### Dependencies

- [pkg/graph/graph.go](#module-pkg-graph-graph-go) - Graph data structures for knowledge graph representation
- [pkg/server/rest/handler.go](#module-pkg-server-rest-handler-go) - REST API handler for GraphFS


## Module: pkg/server/rest/handler.go

**Layer:** server  
**Tags:** rest, api, http, server  
**Language:** go  

### Description
REST API handler for GraphFS

### Dependencies

- [pkg/graph/graph.go](#module-pkg-graph-graph-go) - Graph data structures for knowledge graph representation

### Dependents

- [pkg/server/rest/tags.go](#module-pkg-server-rest-tags-go) - Tag and export endpoints for REST API
- [pkg/server/rest/modules.go](#module-pkg-server-rest-modules-go) - Module endpoints for REST API
- [pkg/server/rest/analysis.go](#module-pkg-server-rest-analysis-go) - Analysis endpoints for REST API

### Exports

- `#Handler`
- `#NewHandler`


## Module: pkg/server/rest/modules.go

**Layer:** server  
**Tags:** rest, api, modules  
**Language:** go  

### Description
Module endpoints for REST API

### Dependencies

- [pkg/graph/graph.go](#module-pkg-graph-graph-go) - Graph data structures for knowledge graph representation
- [pkg/server/rest/handler.go](#module-pkg-server-rest-handler-go) - REST API handler for GraphFS


## Module: pkg/server/rest/tags.go

**Layer:** server  
**Tags:** rest, api, tags, exports  
**Language:** go  

### Description
Tag and export endpoints for REST API

### Dependencies

- [pkg/graph/graph.go](#module-pkg-graph-graph-go) - Graph data structures for knowledge graph representation
- [pkg/server/rest/handler.go](#module-pkg-server-rest-handler-go) - REST API handler for GraphFS


## Module: pkg/server/server.go

**Layer:** server  
**Tags:** server, http, api  
**Language:** go  

### Description
HTTP server for GraphFS query endpoints

### Dependencies

- [pkg/server/sparql_handler.go](#module-pkg-server-sparql-handler-go) - HTTP handler for SPARQL queries
- [pkg/query/executor.go](#module-pkg-query-executor-go) - SPARQL query executor

### Exports

- `#Server`
- `#Config`
- `#NewServer`


## Module: pkg/server/sparql_handler.go

**Layer:** server  
**Tags:** server, sparql, http, handler  
**Language:** go  

### Description
HTTP handler for SPARQL queries

### Dependencies

- [pkg/query/executor.go](#module-pkg-query-executor-go) - SPARQL query executor

### Dependents

- [pkg/server/server.go](#module-pkg-server-server-go) - HTTP server for GraphFS query endpoints

### Exports

- `#SPARQLHandler`
- `#NewSPARQLHandler`


## Module: pkg/viz/dot.go

**Layer:** visualization  
**Tags:** visualization, graphviz, dot, export  
**Language:** go  

### Description
GraphViz DOT format generation for dependency visualization

### Dependencies

- [pkg/graph/graph.go](#module-pkg-graph-graph-go) - Graph data structures for knowledge graph representation
- [pkg/analysis/impact.go](#module-pkg-analysis-impact-go) - Impact analysis engine for GraphFS
- [pkg/analysis/security.go](#module-pkg-analysis-security-go) - Security boundary analysis for detecting and enforcing security zones

### Exports

- `#GenerateDOT`
- `#VizOptions`
- `#VizType`
- `#RenderToFile`



---

*Generated by GraphFS on 2025-11-17 16:54:05*
