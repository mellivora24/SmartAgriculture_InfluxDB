import type { MCUStatus } from './common';

export interface MCU {
    id: string;
    farm_id: string;
    mcu_code: string;
    status: MCUStatus;
    created_at: string;
    updated_at: string;
}

export interface CreateMCURequest {
    farm_id: string;
    mcu_code: string;
}

export interface UpdateMCUStatusRequest {
    status: MCUStatus;
}

export interface MCUWithStats {
    mcu_id: string;
    mcu_code: string;
    status: MCUStatus;
    survey_point_count: number;
    created_at: string;
    updated_at: string;
}
