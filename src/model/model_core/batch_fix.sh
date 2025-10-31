#!/bin/bash

# Fix common exported struct patterns
sed -i 's/^OrganizationRequest struct {$/\/\/ OrganizationRequest represents the request payload for organization operations\
OrganizationRequest struct {/' organization.go

sed -i 's/^OrganizationSubscriptionRequest struct {$/\/\/ OrganizationSubscriptionRequest represents the request payload for organization subscription operations\
OrganizationSubscriptionRequest struct {/' organization.go

sed -i 's/^CreateOrganizationResponse struct {$/\/\/ CreateOrganizationResponse represents the response structure for organization creation\
CreateOrganizationResponse struct {/' organization.go

sed -i 's/^OrganizationPerCategoryResponse struct {$/\/\/ OrganizationPerCategoryResponse represents organization data grouped by category\
OrganizationPerCategoryResponse struct {/' organization.go

echo "Batch fixes applied!"
