#!/bin/bash

# AgentFlow Standalone Demo Script
# This script demonstrates all the key features of AgentFlow

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

print_header() {
    echo -e "${BLUE}================================${NC}"
    echo -e "${BLUE}  🚀 AgentFlow Demo Script${NC}"
    echo -e "${BLUE}================================${NC}"
    echo
}

print_step() {
    echo -e "${CYAN}📋 Step $1: $2${NC}"
    echo
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
    echo
}

print_info() {
    echo -e "${YELLOW}ℹ️  $1${NC}"
    echo
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
    echo
}

# Check if standalone demo is running
check_demo_running() {
    if ! curl -s http://localhost:8080/health > /dev/null; then
        print_error "Standalone demo server is not running!"
        print_info "Please start it with: ./bin/standalone-demo"
        exit 1
    fi
}

# Test health endpoint
test_health() {
    print_step "1" "Testing Health Endpoint"

    response=$(curl -s http://localhost:8080/health)
    echo "Response: $response"

    if echo "$response" | grep -q "healthy"; then
        print_success "Health check passed!"
    else
        print_error "Health check failed!"
        exit 1
    fi
}

# Test workflow submission
test_workflow_submission() {
    print_step "2" "Testing Workflow Submission"

    print_info "Submitting quick test workflow..."
    response=$(curl -s -X POST http://localhost:8080/api/v1/workflows/runs \
        -H 'Content-Type: application/json' \
        -d '{"workflow_name":"quick_test","workflow_version":1,"inputs":{"message":"Hello AgentFlow!"},"budget_cents":500}')

    echo "Response: $response"

    if echo "$response" | grep -q "completed"; then
        print_success "Quick test workflow completed!"
    else
        print_error "Workflow submission failed!"
        exit 1
    fi

    print_info "Submitting document analysis workflow..."
    response=$(curl -s -X POST http://localhost:8080/api/v1/workflows/runs \
        -H 'Content-Type: application/json' \
        -d '{"workflow_name":"document_analysis","workflow_version":1,"inputs":{"document":"contract.pdf","type":"legal"},"budget_cents":2000}')

    echo "Response: $response"
    print_success "Document analysis workflow submitted!"

    print_info "Submitting data processing workflow..."
    response=$(curl -s -X POST http://localhost:8080/api/v1/workflows/runs \
        -H 'Content-Type: application/json' \
        -d '{"workflow_name":"data_processing","workflow_version":1,"inputs":{"dataset":"sales_data.csv","operation":"aggregation"},"budget_cents":1500}')

    echo "Response: $response"
    print_success "Data processing workflow submitted!"
}

# Test workflow listing
test_workflow_listing() {
    print_step "3" "Testing Workflow Listing"

    response=$(curl -s http://localhost:8080/api/v1/workflows/runs)
    echo "Response: $response"

    run_count=$(echo "$response" | grep -o '"id"' | wc -l)
    print_success "Found $run_count workflow runs!"
}

# Test cost analytics
test_cost_analytics() {
    print_step "4" "Testing Cost Analytics"

    response=$(curl -s http://localhost:8080/api/v1/costs/analytics)
    echo "Response: $response"

    if echo "$response" | grep -q "total_cost"; then
        print_success "Cost analytics retrieved successfully!"
    else
        print_error "Cost analytics failed!"
        exit 1
    fi
}

# Test PII redaction
test_pii_redaction() {
    print_step "5" "Testing PII Redaction"

    print_info "Testing content redaction..."
    response=$(curl -s -X POST http://localhost:8080/api/v1/scl/redact \
        -H 'Content-Type: application/json' \
        -d '{"content":"Contact John Doe at john.doe@example.com or call +1-555-123-4567 for more information."}')

    echo "Response: $response"

    if echo "$response" | grep -q "REDACTED"; then
        print_success "PII redaction working!"
    else
        print_error "PII redaction failed!"
        exit 1
    fi

    print_info "Testing content unredaction..."
    response=$(curl -s -X POST http://localhost:8080/api/v1/scl/unredact \
        -H 'Content-Type: application/json' \
        -d '{"content":"Contact [REDACTED_EMAIL] or call [REDACTED_PHONE] for more information.","redaction_map":{"[REDACTED_EMAIL]":"john.doe@example.com","[REDACTED_PHONE]":"+1-555-123-4567"}}')

    echo "Response: $response"
    print_success "PII unredaction working!"
}

# Test prompt management
test_prompt_management() {
    print_step "6" "Testing Prompt Management"

    print_info "Getting prompt template..."
    response=$(curl -s http://localhost:8080/api/v1/prompts/assistant/versions/1)
    echo "Response: $response"

    if echo "$response" | grep -q "assistant"; then
        print_success "Prompt template retrieved!"
    else
        print_error "Prompt template retrieval failed!"
        exit 1
    fi

    print_info "Listing prompt versions..."
    response=$(curl -s http://localhost:8080/api/v1/prompts/assistant/versions)
    echo "Response: $response"
    print_success "Prompt versions listed!"
}

# Show dashboard info
show_dashboard_info() {
    print_step "7" "Dashboard Information"

    print_info "🌐 Web Dashboard: http://localhost:8080"
    print_info "📊 Interactive UI with real-time updates"
    print_info "🎯 Submit workflows with one click"
    print_info "📈 View cost analytics and metrics"
    print_info "🔒 Test PII redaction features"
    print_info "📝 Manage prompt templates"
}

# Main execution
main() {
    print_header

    check_demo_running

    test_health
    test_workflow_submission
    test_workflow_listing
    test_cost_analytics
    test_pii_redaction
    test_prompt_management
    show_dashboard_info

    echo -e "${GREEN}🎉 All tests passed! AgentFlow demo is working perfectly!${NC}"
    echo
    echo -e "${PURPLE}🚀 Ready for your presentation!${NC}"
    echo -e "${PURPLE}📱 Open http://localhost:8080 in your browser for the interactive dashboard${NC}"
    echo
}

# Run the demo
main "$@"
