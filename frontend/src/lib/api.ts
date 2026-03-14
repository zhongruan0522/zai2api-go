const API_BASE = process.env.NEXT_PUBLIC_API_URL || '';

class ApiClient {
  private getToken(): string | null {
    if (typeof window === 'undefined') return null;
    return localStorage.getItem('token');
  }

  private setToken(token: string | null) {
    if (typeof window === 'undefined') return;
    if (token) {
      localStorage.setItem('token', token);
    } else {
      localStorage.removeItem('token');
    }
  }

  private async request<T>(path: string, options: RequestInit = {}, skipAuthRedirect = false): Promise<T> {
    const token = this.getToken();
    const headers: HeadersInit = {
      'Content-Type': 'application/json',
      ...options.headers,
    };
    if (token) {
      (headers as Record<string, string>)['Authorization'] = `Bearer ${token}`;
    }

    const res = await fetch(`${API_BASE}${path}`, {
      ...options,
      headers,
    });

    if (res.status === 401 && !skipAuthRedirect) {
      this.setToken(null);
      if (typeof window !== 'undefined' && !window.location.pathname.startsWith('/login')) {
        window.location.href = '/login';
      }
    }

    if (!res.ok) {
      const error = await res.json().catch(() => ({ error: 'Unknown error' }));
      throw new Error(error.error || `HTTP ${res.status}`);
    }

    return res.json();
  }

  // Auth
  async login(username: string, password: string) {
    const res = await this.request<{ token: string }>('/api/login', {
      method: 'POST',
      body: JSON.stringify({ username, password }),
    }, true);
    this.setToken(res.token);
    return res;
  }

  logout() {
    this.setToken(null);
  }

  isAuthenticated() {
    return !!this.getToken();
  }

  // Audio Tokens
  getAudioTokens() {
    return this.request<TokenItem[]>('/api/tokens/audio');
  }

  createAudioTokens(tokens: string[]) {
    return this.request<{ created: number; duplicates: number }>('/api/tokens/audio', {
      method: 'POST',
      body: JSON.stringify({ tokens }),
    });
  }

  deleteAudioToken(id: number) {
    return this.request<{ message: string }>(`/api/tokens/audio/${id}`, {
      method: 'DELETE',
    });
  }

  toggleAudioToken(id: number) {
    return this.request<TokenItem>(`/api/tokens/audio/${id}/toggle`, {
      method: 'PUT',
    });
  }

  batchDeleteAudioTokens(ids: number[]) {
    return this.request<{ deleted: number }>('/api/tokens/audio/batch-delete', {
      method: 'POST',
      body: JSON.stringify({ ids }),
    });
  }

  batchToggleAudioTokens(ids: number[], enable: boolean) {
    return this.request<{ updated: number }>('/api/tokens/audio/batch-toggle', {
      method: 'POST',
      body: JSON.stringify({ ids, enable }),
    });
  }

  // OCR Tokens
  getOCRTokens() {
    return this.request<TokenItem[]>('/api/tokens/ocr');
  }

  createOCRTokens(tokens: string[]) {
    return this.request<{ created: number; duplicates: number }>('/api/tokens/ocr', {
      method: 'POST',
      body: JSON.stringify({ tokens }),
    });
  }

  deleteOCRToken(id: number) {
    return this.request<{ message: string }>(`/api/tokens/ocr/${id}`, {
      method: 'DELETE',
    });
  }

  toggleOCRToken(id: number) {
    return this.request<TokenItem>(`/api/tokens/ocr/${id}/toggle`, {
      method: 'PUT',
    });
  }

  batchDeleteOCRTokens(ids: number[]) {
    return this.request<{ deleted: number }>('/api/tokens/ocr/batch-delete', {
      method: 'POST',
      body: JSON.stringify({ ids }),
    });
  }

  batchToggleOCRTokens(ids: number[], enable: boolean) {
    return this.request<{ updated: number }>('/api/tokens/ocr/batch-toggle', {
      method: 'POST',
      body: JSON.stringify({ ids, enable }),
    });
  }

  // Chat Tokens
  getChatTokens() {
    return this.request<TokenItem[]>('/api/tokens/chat');
  }

  createChatTokens(tokens: string[]) {
    return this.request<{ created: number; duplicates: number }>('/api/tokens/chat', {
      method: 'POST',
      body: JSON.stringify({ tokens }),
    });
  }

  deleteChatToken(id: number) {
    return this.request<{ message: string }>(`/api/tokens/chat/${id}`, {
      method: 'DELETE',
    });
  }

  toggleChatToken(id: number) {
    return this.request<TokenItem>(`/api/tokens/chat/${id}/toggle`, {
      method: 'PUT',
    });
  }

  batchDeleteChatTokens(ids: number[]) {
    return this.request<{ deleted: number }>('/api/tokens/chat/batch-delete', {
      method: 'POST',
      body: JSON.stringify({ ids }),
    });
  }

  batchToggleChatTokens(ids: number[], enable: boolean) {
    return this.request<{ updated: number }>('/api/tokens/chat/batch-toggle', {
      method: 'POST',
      body: JSON.stringify({ ids, enable }),
    });
  }

  // API Keys
  getAPIKeys() {
    return this.request<APIKeyItem[]>('/api/apikeys');
  }

  createAPIKey(services: string) {
    return this.request<APIKeyItem>('/api/apikeys', {
      method: 'POST',
      body: JSON.stringify({ services }),
    });
  }

  deleteAPIKey(id: number) {
    return this.request<{ message: string }>(`/api/apikeys/${id}`, {
      method: 'DELETE',
    });
  }

  toggleAPIKey(id: number) {
    return this.request<APIKeyItem>(`/api/apikeys/${id}/toggle`, {
      method: 'PUT',
    });
  }

  batchDeleteAPIKeys(ids: number[]) {
    return this.request<{ deleted: number }>('/api/apikeys/batch-delete', {
      method: 'POST',
      body: JSON.stringify({ ids }),
    });
  }

  batchToggleAPIKeys(ids: number[], enable: boolean) {
    return this.request<{ updated: number }>('/api/apikeys/batch-toggle', {
      method: 'POST',
      body: JSON.stringify({ ids, enable }),
    });
  }

  // Request Logs
  getLogs(channel?: string, page?: number) {
    const params = new URLSearchParams();
    if (channel) params.set('channel', channel);
    if (page) params.set('page', String(page));
    return this.request<{ data: LogItem[]; total: number; page: number }>(
      `/api/logs?${params.toString()}`
    );
  }

  getLogStats() {
    return this.request<LogStats>('/api/logs/stats');
  }
}

export interface TokenItem {
  id: number;
  token: string;
  imported_at: string;
  last_used_at: string | null;
  enabled: boolean;
  total_call_count: number;
  daily_call_count: number;
  daily_limit?: number;
}

export interface APIKeyItem {
  id: number;
  key: string;
  services: string;
  enabled: boolean;
  created_at: string;
}

export interface LogItem {
  id: number;
  request_id: string;
  created_at: string;
  channel: string;
  source_ip: string;
  api_key_id: number;
  token_id: number;
  success: boolean;
  error_code: string;
  error_msg: string;
}

export interface LogStats {
  total: number;
  success: number;
  failed: number;
  today: number;
  ocr: number;
  audio: number;
  chat: number;
}

export const api = new ApiClient();
