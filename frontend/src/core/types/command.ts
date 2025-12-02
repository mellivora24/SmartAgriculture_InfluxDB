import type { CommandStatus } from './common';

export interface Command {
    id: string;
    user_id: string;
    device_name: string;
    command: string;
    status: CommandStatus;
    executed_at: string;
    created_at: string;
}

export interface CreateCommandRequest {
    device_name: string;
    command: string;
    value?: number;
}

export interface CreateCommandResponse {
    success: boolean;
    command_id: string;
    message: string;
}

export interface UpdateCommandStatusRequest {
    status: CommandStatus;
}

export interface CommandHistoryItem {
    command_id: string;
    user_id: string;
    username: string;
    device_name: string;
    command: string;
    status: CommandStatus;
    executed_at: string;
    created_at: string;
}

export interface QueryCommandHistoryParams {
    user_id?: string;
    device_name?: string;
    limit?: number;
}
