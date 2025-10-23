import { create } from 'zustand';
import { Agent } from '@/types/api';
import { agentService } from '@/services/agent.service';

interface AgentState {
  agents: Agent[];
  selectedAgent: Agent | null;
  total: number;
  page: number;
  pageSize: number;
  isLoading: boolean;
  error: string | null;

  // Actions
  fetchAgents: (page?: number, pageSize?: number) => Promise<void>;
  fetchAgent: (id: string) => Promise<void>;
  deleteAgent: (id: string) => Promise<void>;
  clearError: () => void;
}

export const useAgentStore = create<AgentState>((set, get) => ({
  agents: [],
  selectedAgent: null,
  total: 0,
  page: 1,
  pageSize: 10,
  isLoading: false,
  error: null,

  fetchAgents: async (page = 1, pageSize = 10) => {
    set({ isLoading: true, error: null });
    try {
      const response = await agentService.getAgents({ page, page_size: pageSize });
      set({
        agents: response.agents || [],
        total: response.total || 0,
        page,
        pageSize,
        isLoading: false,
      });
    } catch (error: any) {
      const errorMessage = error.response?.data?.error || '获取 Agent 列表失败';
      set({ error: errorMessage, isLoading: false });
    }
  },

  fetchAgent: async (id: string) => {
    set({ isLoading: true, error: null });
    try {
      const agent = await agentService.getAgent(id);
      set({ selectedAgent: agent, isLoading: false });
    } catch (error: any) {
      const errorMessage = error.response?.data?.error || '获取 Agent 详情失败';
      set({ error: errorMessage, isLoading: false });
    }
  },

  deleteAgent: async (id: string) => {
    set({ isLoading: true, error: null });
    try {
      await agentService.deleteAgent(id);
      // 删除成功后重新加载列表
      await get().fetchAgents(get().page, get().pageSize);
    } catch (error: any) {
      const errorMessage = error.response?.data?.error || '删除 Agent 失败';
      set({ error: errorMessage, isLoading: false });
      throw error;
    }
  },

  clearError: () => {
    set({ error: null });
  },
}));
