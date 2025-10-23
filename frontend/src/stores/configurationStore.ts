import { create } from 'zustand';
import { Configuration, CreateConfigurationRequest, UpdateConfigurationRequest } from '@/types/api';
import { configurationService } from '@/services/configuration.service';

interface ConfigurationState {
  configurations: Configuration[];
  selectedConfiguration: Configuration | null;
  total: number;
  isLoading: boolean;
  error: string | null;

  // Actions
  fetchConfigurations: () => Promise<void>;
  fetchConfiguration: (name: string) => Promise<void>;
  createConfiguration: (data: CreateConfigurationRequest) => Promise<void>;
  updateConfiguration: (name: string, data: UpdateConfigurationRequest) => Promise<void>;
  deleteConfiguration: (name: string) => Promise<void>;
  clearError: () => void;
}

export const useConfigurationStore = create<ConfigurationState>((set, get) => ({
  configurations: [],
  selectedConfiguration: null,
  total: 0,
  isLoading: false,
  error: null,

  fetchConfigurations: async () => {
    set({ isLoading: true, error: null });
    try {
      const response = await configurationService.getConfigurations();
      set({
        configurations: response.configurations || [],
        total: response.total || 0,
        isLoading: false,
      });
    } catch (error: any) {
      const errorMessage = error.response?.data?.error || '获取配置列表失败';
      set({ error: errorMessage, isLoading: false });
    }
  },

  fetchConfiguration: async (name: string) => {
    set({ isLoading: true, error: null });
    try {
      const configuration = await configurationService.getConfiguration(name);
      set({ selectedConfiguration: configuration, isLoading: false });
    } catch (error: any) {
      const errorMessage = error.response?.data?.error || '获取配置详情失败';
      set({ error: errorMessage, isLoading: false });
    }
  },

  createConfiguration: async (data: CreateConfigurationRequest) => {
    set({ isLoading: true, error: null });
    try {
      await configurationService.createConfiguration(data);
      // 创建成功后重新加载列表
      await get().fetchConfigurations();
    } catch (error: any) {
      const errorMessage = error.response?.data?.error || '创建配置失败';
      set({ error: errorMessage, isLoading: false });
      throw error;
    }
  },

  updateConfiguration: async (name: string, data: UpdateConfigurationRequest) => {
    set({ isLoading: true, error: null });
    try {
      await configurationService.updateConfiguration(name, data);
      // 更新成功后重新加载列表
      await get().fetchConfigurations();
    } catch (error: any) {
      const errorMessage = error.response?.data?.error || '更新配置失败';
      set({ error: errorMessage, isLoading: false });
      throw error;
    }
  },

  deleteConfiguration: async (name: string) => {
    set({ isLoading: true, error: null });
    try {
      await configurationService.deleteConfiguration(name);
      // 删除成功后重新加载列表
      await get().fetchConfigurations();
    } catch (error: any) {
      const errorMessage = error.response?.data?.error || '删除配置失败';
      set({ error: errorMessage, isLoading: false });
      throw error;
    }
  },

  clearError: () => {
    set({ error: null });
  },
}));
