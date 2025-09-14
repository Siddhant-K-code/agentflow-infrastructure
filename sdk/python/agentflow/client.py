"""
AgentFlow Python Client

Main client class for interacting with AgentFlow orchestrator.
"""

import asyncio
from typing import List, Optional, Dict, Any
import requests
import aiohttp
import yaml
from .models import Workflow, WorkflowStatus, AgentExecution
from .exceptions import AgentFlowError, WorkflowNotFoundError


class AgentFlowClient:
    """
    AgentFlow client for managing AI agent workflows.
    
    Example:
        >>> client = AgentFlowClient("http://localhost:8080")
        >>> workflow = client.deploy_workflow("my-workflow.yaml")
        >>> status = client.get_workflow_status(workflow.id)
    """
    
    def __init__(self, base_url: str = "http://localhost:8080", api_key: Optional[str] = None):
        """
        Initialize AgentFlow client.
        
        Args:
            base_url: AgentFlow orchestrator URL
            api_key: Optional API key for authentication
        """
        self.base_url = base_url.rstrip('/')
        self.api_key = api_key
        self._session = requests.Session()
        
        if api_key:
            self._session.headers.update({"Authorization": f"Bearer {api_key}"})
    
    def deploy_workflow(self, workflow_file: str) -> Workflow:
        """
        Deploy a workflow from YAML file.
        
        Args:
            workflow_file: Path to workflow YAML file
            
        Returns:
            Deployed workflow instance
            
        Raises:
            AgentFlowError: If deployment fails
        """
        with open(workflow_file, 'r') as f:
            workflow_data = yaml.safe_load(f)
        
        response = self._session.post(
            f"{self.base_url}/api/v1/workflows",
            json=workflow_data
        )
        
        if response.status_code != 201:
            raise AgentFlowError(f"Failed to deploy workflow: {response.text}")
        
        return Workflow.from_dict(response.json())
    
    def get_workflow(self, workflow_id: str) -> Workflow:
        """
        Get workflow by ID.
        
        Args:
            workflow_id: Workflow identifier
            
        Returns:
            Workflow instance
            
        Raises:
            WorkflowNotFoundError: If workflow doesn't exist
        """
        response = self._session.get(f"{self.base_url}/api/v1/workflows/{workflow_id}")
        
        if response.status_code == 404:
            raise WorkflowNotFoundError(f"Workflow {workflow_id} not found")
        elif response.status_code != 200:
            raise AgentFlowError(f"Failed to get workflow: {response.text}")
        
        return Workflow.from_dict(response.json())
    
    def get_workflow_status(self, workflow_id: str) -> WorkflowStatus:
        """
        Get workflow execution status.
        
        Args:
            workflow_id: Workflow identifier
            
        Returns:
            Workflow status
        """
        response = self._session.get(f"{self.base_url}/api/v1/workflows/{workflow_id}/status")
        
        if response.status_code != 200:
            raise AgentFlowError(f"Failed to get workflow status: {response.text}")
        
        return WorkflowStatus.from_dict(response.json())
    
    def list_workflows(self) -> List[Workflow]:
        """
        List all workflows.
        
        Returns:
            List of workflows
        """
        response = self._session.get(f"{self.base_url}/api/v1/workflows")
        
        if response.status_code != 200:
            raise AgentFlowError(f"Failed to list workflows: {response.text}")
        
        data = response.json()
        return [Workflow.from_dict(w) for w in data.get('workflows', [])]
    
    def delete_workflow(self, workflow_id: str) -> bool:
        """
        Delete a workflow.
        
        Args:
            workflow_id: Workflow identifier
            
        Returns:
            True if successful
        """
        response = self._session.delete(f"{self.base_url}/api/v1/workflows/{workflow_id}")
        
        if response.status_code == 404:
            raise WorkflowNotFoundError(f"Workflow {workflow_id} not found")
        elif response.status_code != 200:
            raise AgentFlowError(f"Failed to delete workflow: {response.text}")
        
        return True
    
    def get_workflow_logs(self, workflow_id: str, agent_name: Optional[str] = None) -> List[Dict[str, Any]]:
        """
        Get workflow logs.
        
        Args:
            workflow_id: Workflow identifier
            agent_name: Optional agent name to filter logs
            
        Returns:
            List of log entries
        """
        params = {}
        if agent_name:
            params['agent'] = agent_name
        
        response = self._session.get(
            f"{self.base_url}/api/v1/workflows/{workflow_id}/logs",
            params=params
        )
        
        if response.status_code != 200:
            raise AgentFlowError(f"Failed to get workflow logs: {response.text}")
        
        return response.json().get('logs', [])
    
    async def stream_workflow_events(self, workflow_id: str):
        """
        Stream real-time workflow events via WebSocket.
        
        Args:
            workflow_id: Workflow identifier
            
        Yields:
            Workflow events as they occur
        """
        import websockets
        
        ws_url = self.base_url.replace('http', 'ws') + f"/api/v1/workflows/{workflow_id}/live"
        
        async with websockets.connect(ws_url) as websocket:
            while True:
                try:
                    message = await websocket.recv()
                    yield yaml.safe_load(message)
                except websockets.exceptions.ConnectionClosed:
                    break


class AsyncAgentFlowClient:
    """
    Async version of AgentFlow client for high-performance applications.
    """
    
    def __init__(self, base_url: str = "http://localhost:8080", api_key: Optional[str] = None):
        self.base_url = base_url.rstrip('/')
        self.api_key = api_key
        self._session = None
    
    async def __aenter__(self):
        headers = {}
        if self.api_key:
            headers["Authorization"] = f"Bearer {self.api_key}"
        
        self._session = aiohttp.ClientSession(headers=headers)
        return self
    
    async def __aexit__(self, exc_type, exc_val, exc_tb):
        if self._session:
            await self._session.close()
    
    async def deploy_workflow(self, workflow_file: str) -> Workflow:
        """Async version of deploy_workflow"""
        with open(workflow_file, 'r') as f:
            workflow_data = yaml.safe_load(f)
        
        async with self._session.post(
            f"{self.base_url}/api/v1/workflows",
            json=workflow_data
        ) as response:
            if response.status != 201:
                text = await response.text()
                raise AgentFlowError(f"Failed to deploy workflow: {text}")
            
            data = await response.json()
            return Workflow.from_dict(data)
    
    async def get_workflow_status(self, workflow_id: str) -> WorkflowStatus:
        """Async version of get_workflow_status"""
        async with self._session.get(
            f"{self.base_url}/api/v1/workflows/{workflow_id}/status"
        ) as response:
            if response.status != 200:
                text = await response.text()
                raise AgentFlowError(f"Failed to get workflow status: {text}")
            
            data = await response.json()
            return WorkflowStatus.from_dict(data)