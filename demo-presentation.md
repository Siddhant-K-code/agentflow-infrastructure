# 🚀 AgentFlow Demo Presentation

## Demo Script for Tonight

### **Opening Hook (30 seconds)**
*"While other tools help you route conversations between agents, AgentFlow orchestrates complex business processes with AI. Think of it as the Kubernetes for AI workflows."*

### **Problem Statement (1 minute)**
*"Enterprises today need to orchestrate complex AI-powered business processes:*
- *Multi-step document processing pipelines*
- *Customer support automation with multiple AI models*
- *Content moderation workflows with compliance requirements*
- *Cost management and budget enforcement*
- *Audit trails and observability*

*Current solutions are either too simple (just agent routing) or too complex (full workflow engines without AI focus)."*

### **AgentFlow Solution (2 minutes)**

#### **The 5-Pillar Architecture**
1. **🎯 AOR (Agent Orchestration Runtime)** - Fan-out/fan-in, retries, backpressure
2. **📝 POP (PromptOps Platform)** - Versioned prompts, evaluation, canary rollouts
3. **🔍 SCL (Secure Context Layer)** - PII redaction, compliance, audit trails
4. **📊 AOS (Agent Observability Stack)** - Semantic tracing, replay, root-cause analysis
5. **💰 CAS (Cost-Aware Scheduler)** - Budget enforcement, cost optimization

#### **Live Demo Scenarios**

**Scenario 1: Document Analysis Pipeline**
```bash
curl -X POST http://localhost:8080/api/v1/workflows/submit \
  -H 'Content-Type: application/json' \
  -d '{
    "workflow_name": "document_analysis",
    "workflow_version": 1,
    "inputs": {
      "document_url": "https://example.com/sample.pdf",
      "analysis_type": "comprehensive"
    },
    "budget_cents": 1000
  }'
```

**Scenario 2: Customer Support Automation**
```bash
curl -X POST http://localhost:8080/api/v1/workflows/submit \
  -H 'Content-Type: application/json' \
  -d '{
    "workflow_name": "customer_support",
    "workflow_version": 1,
    "inputs": {
      "customer_query": "I'm having trouble with my account",
      "priority": "high"
    },
    "budget_cents": 500
  }'
```

**Scenario 3: Content Moderation Workflow**
```bash
curl -X POST http://localhost:8080/api/v1/workflows/submit \
  -H 'Content-Type: application/json' \
  -d '{
    "workflow_name": "content_moderation",
    "workflow_version": 1,
    "inputs": {
      "content": "Sample user-generated content",
      "platform": "social_media"
    },
    "budget_cents": 300
  }'
```

### **Key Differentiators (1 minute)**

#### **vs Agent Squad (Conversational AI)**
- **Agent Squad**: Routes conversations between agents
- **AgentFlow**: Orchestrates complex business processes with DAGs

#### **vs Temporal/Airflow (Workflow Engines)**
- **Temporal/Airflow**: General-purpose workflow orchestration
- **AgentFlow**: AI-native with cost management, prompt versioning, compliance

#### **vs LangChain/LlamaIndex (AI Frameworks)**
- **LangChain/LlamaIndex**: AI application frameworks
- **AgentFlow**: Enterprise platform for AI workflow orchestration

### **Technical Architecture Demo (2 minutes)**

#### **Show the Web UI**
- Open http://localhost:8080/
- Demonstrate the 3 demo scenarios
- Show real-time workflow execution
- Display cost tracking and status updates

#### **Show the API**
- Demonstrate REST API endpoints
- Show workflow submission and status checking
- Display monitoring and observability

#### **Show the Infrastructure**
- Docker Compose setup
- Multiple databases (PostgreSQL, Redis, ClickHouse)
- NATS messaging
- Monitoring with Prometheus/Grafana

### **Business Value Proposition (1 minute)**

#### **For Enterprises**
- **Compliance**: Built-in PII redaction and audit trails
- **Cost Control**: Budget enforcement and cost optimization
- **Reliability**: Retries, backpressure, idempotency
- **Observability**: Complete traceability and replay capabilities

#### **For Developers**
- **Simple API**: Easy workflow submission and management
- **Flexible**: Support for multiple AI providers
- **Scalable**: Horizontal scaling with worker nodes
- **Observable**: Comprehensive monitoring and debugging

### **Market Opportunity (30 seconds)**
*"The AI workflow orchestration market is growing rapidly. AgentFlow addresses the gap between simple agent routing and complex workflow engines, specifically for AI-native business processes."*

### **Next Steps (30 seconds)**
*"We're looking for early enterprise customers and strategic partners to help us build the future of AI workflow orchestration."*

---

## 🎯 Demo Setup Checklist

### **Before the Demo (15 minutes)**
- [ ] Run `./demo-setup.sh` to start the environment
- [ ] Verify all services are running: http://localhost:8080/health
- [ ] Test the demo UI: http://localhost:8080/
- [ ] Prepare the 3 demo scenarios
- [ ] Have backup slides ready

### **Demo Environment**
- [ ] Control Plane: http://localhost:8080
- [ ] Demo UI: http://localhost:8080/
- [ ] Health Check: http://localhost:8080/health
- [ ] API Scenarios: http://localhost:8080/api/v1/demo/scenarios

### **Backup Plans**
- [ ] If live demo fails, show recorded video
- [ ] If API fails, show the architecture diagrams
- [ ] If UI fails, show curl commands
- [ ] Have screenshots ready as backup

---

## 🎤 Talking Points

### **Opening**
*"Good evening! I'm excited to show you AgentFlow, the enterprise platform for AI workflow orchestration. While other tools help you route conversations between agents, AgentFlow orchestrates complex business processes with AI."*

### **Problem**
*"Enterprises today need to orchestrate complex AI-powered business processes, but current solutions are either too simple or too complex. We need something in between - a platform that's AI-native but enterprise-ready."*

### **Solution**
*"AgentFlow provides a 5-pillar architecture that addresses all aspects of AI workflow orchestration, from execution to compliance to cost management."*

### **Demo**
*"Let me show you this in action with three real-world scenarios..."*

### **Differentiation**
*"What makes AgentFlow unique is that we're not just routing conversations - we're orchestrating complex business processes with enterprise-grade features."*

### **Closing**
*"AgentFlow is the Kubernetes for AI workflows. We're looking for early customers and partners to help us build the future of AI workflow orchestration."*

---

## 📊 Demo Metrics to Track

- **Engagement**: How many people ask questions?
- **Interest**: How many people want to try it?
- **Feedback**: What are the main concerns or suggestions?
- **Follow-up**: How many people want to schedule a deeper demo?

---

## 🎯 Success Criteria

- [ ] Audience understands the problem we're solving
- [ ] Audience sees the value proposition
- [ ] Audience understands the differentiation
- [ ] Audience is interested in learning more
- [ ] We get at least 3 follow-up requests
