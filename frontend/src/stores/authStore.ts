import { create } from 'zustand';
import { User } from '@/types/api';
import { authService } from '@/services/auth.service';

interface AuthState {
  user: User | null;
  token: string | null;
  isLoading: boolean;
  error: string | null;

  // Actions
  login: (username: string, password: string) => Promise<void>;
  register: (username: string, email: string, password: string) => Promise<void>;
  logout: () => void;
  loadUser: () => Promise<void>;
  clearError: () => void;
}

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  token: authService.getToken(),
  isLoading: false,
  error: null,

  login: async (username: string, password: string) => {
    set({ isLoading: true, error: null });
    try {
      const response = await authService.login({ username, password });
      authService.saveToken(response.token);
      localStorage.setItem('user', JSON.stringify(response.user));
      set({ user: response.user, token: response.token, isLoading: false });
    } catch (error: any) {
      const errorMessage = error.response?.data?.error || '登录失败';
      set({ error: errorMessage, isLoading: false });
      throw error;
    }
  },

  register: async (username: string, email: string, password: string) => {
    set({ isLoading: true, error: null });
    try {
      const response = await authService.register({ username, email, password });
      authService.saveToken(response.token);
      localStorage.setItem('user', JSON.stringify(response.user));
      set({ user: response.user, token: response.token, isLoading: false });
    } catch (error: any) {
      const errorMessage = error.response?.data?.error || '注册失败';
      set({ error: errorMessage, isLoading: false });
      throw error;
    }
  },

  logout: () => {
    authService.logout();
    set({ user: null, token: null });
  },

  loadUser: async () => {
    const token = authService.getToken();
    if (!token) {
      set({ user: null, token: null });
      return;
    }

    try {
      const user = await authService.getCurrentUser();
      localStorage.setItem('user', JSON.stringify(user));
      set({ user, token });
    } catch (error) {
      // Token 无效,清除
      authService.logout();
      set({ user: null, token: null });
    }
  },

  clearError: () => {
    set({ error: null });
  },
}));
