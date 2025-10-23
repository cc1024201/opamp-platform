// API 响应通用类型
export interface ApiResponse<T = any> {
  data?: T;
  error?: string;
  message?: string;
}

// 用户相关类型
export interface User {
  id: number;
  username: string;
  email: string;
  role: 'admin' | 'user';
  created_at: string;
  updated_at: string;
}

export interface LoginRequest {
  username: string;
  password: string;
}

export interface RegisterRequest {
  username: string;
  email: string;
  password: string;
  role?: 'admin' | 'user';
}

export interface AuthResponse {
  token: string;
  user: User;
}

// Agent 相关类型
export interface Agent {
  id: string;
  name: string;
  version: string;
  hostname: string;
  architecture: string;
  os_type: string;
  status: 'connected' | 'disconnected' | 'configuring' | 'error';
  last_seen: string;
  labels?: Record<string, string>;
  effective_config?: string;
  created_at: string;
  updated_at: string;
}

export interface AgentListResponse {
  agents: Agent[];
  total: number;
  page: number;
  page_size: number;
}

// Configuration 相关类型
export interface Configuration {
  id: number;
  name: string;
  display_name: string;
  content_type: 'yaml' | 'json';
  raw_config: string;
  config_hash: string;
  selector?: Record<string, string>;
  created_at: string;
  updated_at: string;
}

export interface ConfigurationListResponse {
  configurations: Configuration[];
  total: number;
}

export interface CreateConfigurationRequest {
  name: string;
  display_name: string;
  content_type: 'yaml' | 'json';
  raw_config: string;
  selector?: Record<string, string>;
}

export interface UpdateConfigurationRequest {
  display_name?: string;
  raw_config?: string;
  selector?: Record<string, string>;
}

// 健康检查类型
export interface HealthCheck {
  status: 'healthy' | 'degraded' | 'unhealthy';
  timestamp: string;
  checks: {
    database?: {
      status: string;
      latency_ms?: number;
    };
    redis?: {
      status: string;
      latency_ms?: number;
    };
  };
}

// 分页参数
export interface PaginationParams {
  page?: number;
  page_size?: number;
}

// 配置历史相关类型
export interface ConfigurationHistory {
  id: number;
  configuration_name: string;
  version: number;
  content_type: 'yaml' | 'json';
  raw_config: string;
  config_hash: string;
  selector?: Record<string, string>;
  change_description?: string;
  created_by?: string;
  created_at: string;
}

export interface ConfigurationHistoryListResponse {
  histories: ConfigurationHistory[];
  total: number;
  limit: number;
  offset: number;
}

// 配置应用历史类型
export interface ConfigurationApplyHistory {
  id: number;
  agent_id: string;
  configuration_name: string;
  version: number;
  config_hash: string;
  status: 'pending' | 'applying' | 'applied' | 'failed';
  error_message?: string;
  created_at: string;
  updated_at: string;
}

export interface ApplyHistoryListResponse {
  histories: ConfigurationApplyHistory[];
  total: number;
  limit: number;
  offset: number;
}
