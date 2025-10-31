#!/bin/bash

# Function to add comments to exported structs
fix_exported_structs() {
    local file=$1
    # TimeDepositComputation
    sed -i 's/^TimeDepositComputation struct {$/\/\/ TimeDepositComputation represents time deposit calculation data\
TimeDepositComputation struct {/' "$file" 2>/dev/null
    
    # TimeDepositComputationResponse  
    sed -i 's/^TimeDepositComputationResponse struct {$/\/\/ TimeDepositComputationResponse represents the JSON response for time deposit computation\
TimeDepositComputationResponse struct {/' "$file" 2>/dev/null
    
    # TimeDepositComputationRequest
    sed -i 's/^TimeDepositComputationRequest struct {$/\/\/ TimeDepositComputationRequest represents the request payload for time deposit computation\
TimeDepositComputationRequest struct {/' "$file" 2>/dev/null
    
    # PostDatedCheck
    sed -i 's/^PostDatedCheck struct {$/\/\/ PostDatedCheck represents a post-dated check in the system\
PostDatedCheck struct {/' "$file" 2>/dev/null
    
    # PostDatedCheckResponse
    sed -i 's/^PostDatedCheckResponse struct {$/\/\/ PostDatedCheckResponse represents the JSON response for post-dated check data\
PostDatedCheckResponse struct {/' "$file" 2>/dev/null
    
    # PostDatedCheckRequest
    sed -i 's/^PostDatedCheckRequest struct {$/\/\/ PostDatedCheckRequest represents the request payload for post-dated check operations\
PostDatedCheckRequest struct {/' "$file" 2>/dev/null
}

# Function to add method comments
fix_exported_methods() {
    local file=$1
    # CurrentBranch methods
    sed -i 's/^func (m \*ModelCore) \([A-Za-z]*\)CurrentBranch(context context\.Context, orgId uuid\.UUID, branchId uuid\.UUID)/\/\/ \1CurrentBranch retrieves all \L\1\E records for the specified organization and branch\
func (m *ModelCore) \1CurrentBranch(context context.Context, orgID uuid.UUID, branchID uuid.UUID)/' "$file" 2>/dev/null
    
    # Model initialization methods
    sed -i 's/^func (m \*ModelCore) \([A-Za-z]*\)() {$/\/\/ \1 initializes the \1 model and its repository manager\
func (m *ModelCore) \1() {/' "$file" 2>/dev/null
}

# Apply fixes to specific files
echo "Applying fixes to time_deposit_computation.go..."
fix_exported_structs "time_deposit_computation.go"
fix_exported_methods "time_deposit_computation.go"

echo "Applying fixes to post_dated_check.go..."
fix_exported_structs "post_dated_check.go"
fix_exported_methods "post_dated_check.go"

echo "Fixes applied successfully!"
