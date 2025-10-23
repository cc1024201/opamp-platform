import apiClient from './api';
import { Agent, AgentListResponse, PaginationParams } from '@/types/api';

export const agentService = {
  // 获取 Agent 列表
  async getAgents(params?: PaginationParams): Promise<AgentListResponse> {
    const response = await apiClient.get<AgentListResponse>('/agents', {
      params,
    });
    return response.data;
  },

  // 获取单个 Agent
  async getAgent(id: string): Promise<Agent> {
    const response = await apiClient.get<Agent>(`/agents/${id}`);
    return response.data;
  },

  // 删除 Agent
  async deleteAgent(id: string): Promise<void> {
    await apiClient.delete(`/agents/${id}`);
  },
};
