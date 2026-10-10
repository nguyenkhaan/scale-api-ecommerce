// Standard API Client for SaleScale Backend

export interface ApiResponse<T> {
  success: boolean;
  data?: T;
  error?: {
    code: string;
    message: string;
    details?: unknown;
  };
  meta: {
    request_id: string;
    timestamp: string;
    pagination?: {
      page: number;
      limit: number;
      total_items: number;
      total_pages: number;
    };
  };
}

export interface HealthData {
  service: string;
  status: string;
  environment: string;
  uptime: string;
}

const API_BASE = import.meta.env.VITE_API_BASE_URL || '/api';

export async function fetchHealth(): Promise<ApiResponse<HealthData>> {
  const res = await fetch(`${API_BASE}/health`, {
    headers: {
      Accept: 'application/json',
    },
  });
  return res.json();
}

export async function fetchPing(): Promise<ApiResponse<{ message: string }>> {
  const res = await fetch(`${API_BASE}/ping`, {
    headers: {
      Accept: 'application/json',
    },
  });
  return res.json();
}

