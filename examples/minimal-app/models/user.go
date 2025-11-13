/*
# Module: models/user.go
User data model and validation.

Defines the User entity with fields and validation logic.
This model is used by the user service for CRUD operations.

## Linked Modules
- [utils/validator](../utils/validator.go) - Data validation utilities

## Tags
model, data, user, domain

## Exports
User, UserRole, ValidateUser

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .
@prefix rdfs: <http://www.w3.org/2000/01/rdf-schema#> .

<#user.go> a code:Module ;
    code:name "models/user.go" ;
    code:description "User data model and validation" ;
    code:language "go" ;
    code:layer "model" ;
    code:linksTo <../utils/validator.go> ;
    code:exports <#User>, <#UserRole>, <#ValidateUser> ;
    code:tags "model", "data", "user", "domain" .

<#User> a code:Type ;
    code:name "User" ;
    code:kind "struct" ;
    code:description "User entity with authentication fields" ;
    code:hasField <#User.ID>, <#User.Email>, <#User.Role> .

<#User.ID> a code:Field ;
    code:name "ID" ;
    code:type "string" .

<#User.Email> a code:Field ;
    code:name "Email" ;
    code:type "string" .

<#User.Role> a code:Field ;
    code:name "Role" ;
    code:type "UserRole" .

<#UserRole> a code:Type ;
    code:name "UserRole" ;
    code:kind "enum" ;
    code:description "User role enumeration" ;
    code:hasValue "Admin", "User", "Guest" .

<#ValidateUser> a code:Function ;
    code:name "ValidateUser" ;
    code:description "Validates user data" ;
    code:calls <../utils/validator.go#ValidateEmail> .
<!-- End LinkedDoc RDF -->
*/

package models

import "minimal-app/utils"

// UserRole represents user permission levels
type UserRole string

const (
	RoleAdmin UserRole = "Admin"
	RoleUser  UserRole = "User"
	RoleGuest UserRole = "Guest"
)

// User represents a user entity
type User struct {
	ID    string
	Email string
	Role  UserRole
}

// ValidateUser validates user data
func ValidateUser(user *User) error {
	// Stubbed validation logic
	if user.Email == "" {
		return utils.NewValidationError("email is required")
	}

	if !utils.ValidateEmail(user.Email) {
		return utils.NewValidationError("invalid email format")
	}

	return nil
}
