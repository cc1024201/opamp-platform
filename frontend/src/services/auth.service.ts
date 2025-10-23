import apiClient from './api';
import {
  LoginRequest,
  RegisterRequest,
  AuthResponse,
  User,
} from '@/types/api';

export const authService = {
  // 用户登录
  async login(data: LoginRequest): Promise<AuthResponse> {
    const response = await apiClient.post<AuthResponse>('/auth/login', data);
    return response.data;
  },

  // 用户注册
  async register(data: RegisterRequest): Promise<AuthResponse> {
    const response = await apiClient.post<AuthResponse>('/auth/register', data);
    return response.data;
  },

  // 获取当前用户信息
  async getCurrentUser(): Promise<User> {
    const response = await apiClient.get<User>('/auth/me');
    return response.data;
  },

  // 登出
  logout(): void {
    localStorage.removeItem('token');
    localStorage.removeItem('user');
  },

  // 保存 token
  saveToken(token: string): void {
    localStorage.setItem('token', token);
  },

  // 获取 token
  getToken(): string | null {
    return localStorage.getItem('token');
  },

  // 检查是否已登录
  isAuthenticated(): boolean {
    return !!this.getToken();
  },
};
