export interface ApiResponse<T> {
    data: T;
    message?: string;
}

export interface SuccessResponse {
    success: boolean;
    message: string;
}

export interface ErrorResponse {
    error: string;
    message: string;
    status?: number;
}

export interface PaginationParams {
    limit?: number;
    offset?: number;
}

export interface TimeRange {
    start_time: string;
    end_time: string;
}

export type MCUStatus = 'online' | 'offline';
export type SurveyPointStatus = 'connecting' | 'connected' | 'disconnected';
export type CommandStatus = 'pending' | 'success' | 'failed';
export type UserRole = 'owner' | 'admin' | 'editor' | 'viewer';
export type AlertSeverity = 'info' | 'warning' | 'critical';
