#!/bin/bash

# AgentFlow Demo Examples
echo "🎯 AgentFlow Demo Examples"
echo "=========================="

# Set API endpoint
API_ENDPOINT="http://localhost:8080"

echo ""
echo "1. 📊 System Status Check"
echo "------------------------"
curl -s "$API_ENDPOINT/api/v1/status" | jq '.'

echo ""
echo "2. 💰 Cost Analytics"
echo "-------------------"
curl -s "$API_ENDPOINT/api/v1/costs/analytics" | jq '.'

echo ""
echo "3. 🔒 PII Redaction Demo"
echo "-----------------------"
curl -s -X POST "$API_ENDPOINT/api/v1/scl/redact" \
  -H "Content-Type: application/json" \
  -d '{
    "content": "Contact John Doe at john.doe@example.com or call +1-555-123-4567. SSN: 123-45-6789"
  }' | jq '.'

echo ""
echo "4. 🚀 Submit Workflow"
echo "-------------------"
WORKFLOW_RESPONSE=$(curl -s -X POST "$API_ENDPOINT/api/v1/workflows/runs" \
  -H "Content-Type: application/json" \
  -d '{
    "workflow_name": "document_analysis",
    "workflow_version": 1,
    "inputs": {
      "document_path": "sample.pdf",
      "analysis_type": "comprehensive"
    },
    "budget_cents": 1000,
    "tags": ["demo", "document"]
  }')

echo "$WORKFLOW_RESPONSE" | jq '.'

# Extract workflow ID for next command
WORKFLOW_ID=$(echo "$WORKFLOW_RESPONSE" | jq -r '.id')

echo ""
echo "5. 📋 List Workflow Runs"
echo "----------------------"
curl -s "$API_ENDPOINT/api/v1/workflows/runs" | jq '.'

echo ""
echo "6. 🔍 Get Specific Workflow Run"
echo "-----------------------------"
if [ "$WORKFLOW_ID" != "null" ] && [ "$WORKFLOW_ID" != "" ]; then
    curl -s "$API_ENDPOINT/api/v1/workflows/runs/$WORKFLOW_ID" | jq '.'
else
    echo "No workflow ID available"
fi

echo ""
echo "7. 🎯 Submit Multiple Workflows (Load Test)"
echo "-----------------------------------------"
for i in {1..3}; do
    echo "Submitting workflow $i..."
    curl -s -X POST "$API_ENDPOINT/api/v1/workflows/runs" \
      -H "Content-Type: application/json" \
      -d "{
        \"workflow_name\": \"data_processing_$i\",
        \"workflow_version\": 1,
        \"inputs\": {
          \"dataset\": \"dataset_$i.csv\",
          \"processing_type\": \"batch\"
        },
        \"budget_cents\": $((500 + i * 100)),
        \"tags\": [\"demo\", \"batch\", \"load_test\"]
      }" > /dev/null
done

echo "Load test completed!"

echo ""
echo "8. 📊 Final Status Check"
echo "----------------------"
curl -s "$API_ENDPOINT/api/v1/status" | jq '.'

echo ""
echo "✅ Demo completed! Check the web dashboard at: http://localhost:3000"
echo "   (If you have Grafana running) or the HTML dashboard at: web/dashboard/index.html"
