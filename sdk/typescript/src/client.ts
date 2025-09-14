import axios, { AxiosInstance } from 'axios';
import * as YAML from 'yaml';
import { Workflow, WorkflowStatus, AgentExecution } from './models';
import { AgentFlowError, WorkflowNotFoundError } from './errors';

/**
 * AgentFlow client for managing AI agent workflows.
 * 
 * @example
 * ```typescript
 * const client = new AgentFlowClient('http://localhost:8080');
 * const workflow = await client.deployWorkflow('./my-workflow.yaml');
 * const status = await client.getWorkflowStatus(workflow.id);
 * ```
 */
export class AgentFlowClient {
  private client: AxiosInstance;

  constructor(baseURL: string = 'http://localhost:8080', apiKey?: string) {
    this.client = axios.create({
      baseURL: baseURL.replace(/\/$/, ''),
      headers: {
        'Content-Type': 'application/json',
        ...(apiKey && { Authorization: `Bearer ${apiKey}` }),
      },
    });
  }

  /**
   * Deploy a workflow from YAML file content.
   */
  async deployWorkflow(workflowYaml: string): Promise<Workflow> {
    try {
      const workflowData = YAML.parse(workflowYaml);
      const response = await this.client.post('/api/v1/workflows', workflowData);
      return new Workflow(response.data);
    } catch (error) {
      if (axios.isAxiosError(error)) {
        throw new AgentFlowError(`Failed to deploy workflow: ${error.response?.data || error.message}`);
      }
      throw error;
    }
  }

  /**
   * Get workflow by ID.
   */
  async getWorkflow(workflowId: string): Promise<Workflow> {
    try {
      const response = await this.client.get(`/api/v1/workflows/${workflowId}`);
      return new Workflow(response.data);
    } catch (error) {
      if (axios.isAxiosError(error)) {
        if (error.response?.status === 404) {
          throw new WorkflowNotFoundError(`Workflow ${workflowId} not found`);
        }
        throw new AgentFlowError(`Failed to get workflow: ${error.response?.data || error.message}`);
      }
      throw error;
    }
  }

  /**
   * Get workflow execution status.
   */
  async getWorkflowStatus(workflowId: string): Promise<WorkflowStatus> {
    try {
      const response = await this.client.get(`/api/v1/workflows/${workflowId}/status`);
      return new WorkflowStatus(response.data);
    } catch (error) {
      if (axios.isAxiosError(error)) {
        throw new AgentFlowError(`Failed to get workflow status: ${error.response?.data || error.message}`);
      }
      throw error;
    }
  }

  /**
   * List all workflows.
   */
  async listWorkflows(): Promise<Workflow[]> {
    try {
      const response = await this.client.get('/api/v1/workflows');
      return response.data.workflows.map((w: any) => new Workflow(w));
    } catch (error) {
      if (axios.isAxiosError(error)) {
        throw new AgentFlowError(`Failed to list workflows: ${error.response?.data || error.message}`);
      }
      throw error;
    }
  }

  /**
   * Delete a workflow.
   */
  async deleteWorkflow(workflowId: string): Promise<void> {
    try {
      await this.client.delete(`/api/v1/workflows/${workflowId}`);
    } catch (error) {
      if (axios.isAxiosError(error)) {
        if (error.response?.status === 404) {
          throw new WorkflowNotFoundError(`Workflow ${workflowId} not found`);
        }
        throw new AgentFlowError(`Failed to delete workflow: ${error.response?.data || error.message}`);
      }
      throw error;
    }
  }

  /**
   * Get workflow logs.
   */
  async getWorkflowLogs(workflowId: string, agentName?: string): Promise<any[]> {
    try {
      const params = agentName ? { agent: agentName } : {};
      const response = await this.client.get(`/api/v1/workflows/${workflowId}/logs`, { params });
      return response.data.logs || [];
    } catch (error) {
      if (axios.isAxiosError(error)) {
        throw new AgentFlowError(`Failed to get workflow logs: ${error.response?.data || error.message}`);
      }
      throw error;
    }
  }

  /**
   * Trigger a workflow via webhook.
   */
  async triggerWorkflow(webhookPath: string, payload: any = {}): Promise<void> {
    try {
      await this.client.post(`/api/v1/trigger/${webhookPath}`, payload);
    } catch (error) {
      if (axios.isAxiosError(error)) {
        throw new AgentFlowError(`Failed to trigger workflow: ${error.response?.data || error.message}`);
      }
      throw error;
    }
  }

  /**
   * Stream workflow events via WebSocket.
   */
  async *streamWorkflowEvents(workflowId: string): AsyncGenerator<any, void, unknown> {
    const WebSocket = (await import('ws')).default;
    const wsUrl = this.client.defaults.baseURL!.replace(/^http/, 'ws') + `/api/v1/workflows/${workflowId}/live`;
    
    const ws = new WebSocket(wsUrl);
    
    try {
      for await (const event of this.wsEventGenerator(ws)) {
        yield event;
      }
    } finally {
      ws.close();
    }
  }

  private async *wsEventGenerator(ws: WebSocket): AsyncGenerator<any, void, unknown> {
    const messages: string[] = [];
    let resolve: ((value: string) => void) | null = null;
    let ended = false;

    ws.on('message', (data: string) => {
      if (resolve) {
        resolve(data);
        resolve = null;
      } else {
        messages.push(data);
      }
    });

    ws.on('close', () => {
      ended = true;
      if (resolve) {
        resolve('');
        resolve = null;
      }
    });

    ws.on('error', () => {
      ended = true;
      if (resolve) {
        resolve('');
        resolve = null;
      }
    });

    while (!ended) {
      let message: string;
      
      if (messages.length > 0) {
        message = messages.shift()!;
      } else {
        message = await new Promise<string>((res) => {
          resolve = res;
        });
      }
      
      if (message && !ended) {
        try {
          yield YAML.parse(message);
        } catch {
          yield { raw: message };
        }
      }
    }
  }
}