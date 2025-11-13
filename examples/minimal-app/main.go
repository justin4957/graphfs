/*
# Module: main.go
Minimal GraphFS example application entry point.

This is a minimal example demonstrating LinkedDoc+RDF metadata format.
All code is stubbed but contains working LinkedDoc headers that can be
parsed by GraphFS to build a semantic knowledge graph.

## Linked Modules
- [services/user](./services/user.go) - User management service
- [services/auth](./services/auth.go) - Authentication service
- [utils/logger](./utils/logger.go) - Logging utilities
- [models/user](./models/user.go) - User data model

## Tags
entrypoint, example, demo

## Exports
main

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .
@prefix rdfs: <http://www.w3.org/2000/01/rdf-schema#> .

<#main.go> a code:Module ;
    code:name "main.go" ;
    code:description "Minimal GraphFS example application entry point" ;
    code:language "go" ;
    code:linksTo <./services/user.go>, <./services/auth.go>, <./utils/logger.go>, <./models/user.go> ;
    code:exports <#main> ;
    code:tags "entrypoint", "example", "demo" ;
    code:isEntryPoint true .

<#main> a code:Function ;
    code:name "main" ;
    code:description "Application entry point" ;
    code:calls <./services/user.go#NewUserService>, <./services/auth.go#NewAuthService>, <./utils/logger.go#NewLogger> .
<!-- End LinkedDoc RDF -->
*/

package main

import (
	"fmt"
	"minimal-app/services"
	"minimal-app/utils"
)

// main is the application entry point
func main() {
	// Initialize logger
	logger := utils.NewLogger()
	logger.Info("Starting minimal example application")

	// Initialize services
	authService := services.NewAuthService()
	userService := services.NewUserService(authService)

	// Stubbed application logic
	fmt.Println("Minimal GraphFS Example App")
	fmt.Printf("Auth Service: %v\n", authService)
	fmt.Printf("User Service: %v\n", userService)

	logger.Info("Application initialized successfully")
}
