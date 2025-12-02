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

// Thêm type cho Axios Error
export interface ApiErrorResponse {
    response?: {
        data?: {
            message?: string;
            error?: string;
        };
        status?: number;
    };
    message?: string;
}

export type MCUStatus = 'online' | 'offline';
export type SurveyPointStatus = 'connecting' | 'connected' | 'disconnected';
export type CommandStatus = 'pending' | 'success' | 'failed';
export type UserRole = 'owner' | 'admin' | 'editor' | 'viewer';
export type AlertSeverity = 'info' | 'warning' | 'critical';