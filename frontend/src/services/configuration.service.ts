import apiClient from './api';
import {
  Configuration,
  ConfigurationListResponse,
  CreateConfigurationRequest,
  UpdateConfigurationRequest,
} from '@/types/api';

export const configurationService = {
  // 获取配置列表
  async getConfigurations(): Promise<ConfigurationListResponse> {
    const response = await apiClient.get<ConfigurationListResponse>('/configurations');
    return response.data;
  },

  // 获取单个配置
  async getConfiguration(name: string): Promise<Configuration> {
    const response = await apiClient.get<Configuration>(`/configurations/${name}`);
    return response.data;
  },

  // 创建配置
  async createConfiguration(data: CreateConfigurationRequest): Promise<Configuration> {
    const response = await apiClient.post<Configuration>('/configurations', data);
    return response.data;
  },

  // 更新配置
  async updateConfiguration(
    name: string,
    data: UpdateConfigurationRequest
  ): Promise<Configuration> {
    const response = await apiClient.put<Configuration>(`/configurations/${name}`, data);
    return response.data;
  },

  // 删除配置
  async deleteConfiguration(name: string): Promise<void> {
    await apiClient.delete(`/configurations/${name}`);
  },
};
