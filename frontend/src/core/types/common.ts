export interface ApiResponse<T = unknown> {
  data?: T;
  message?: string;
  error?: string;
  status?: number;
}

export interface PaginationParams {
  limit?: number;
  offset?: number;
}

export interface PaginatedResponse<T> {
  items: T[];
  total: number;
  limit: number;
  offset: number;
}

export interface TimeRange {
  start: string;
  end: string;
}

export type LoadingState = 'idle' | 'loading' | 'success' | 'error';

export interface SelectOption {
  value: string;
  label: string;
}