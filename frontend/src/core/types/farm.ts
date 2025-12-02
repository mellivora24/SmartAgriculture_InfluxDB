import type { UserRole } from './common';

export interface Farm {
    id: string;
    name: string;
    description: string;
    location: string;
    created_at: string;
    updated_at: string;
}

export interface CreateFarmRequest {
    name: string;
    description: string;
    location: string;
}

export interface UpdateFarmRequest {
    name?: string;
    description?: string;
    location?: string;
}

export interface CreateFarmResponse {
    success: boolean;
    farm_id: string;
    message: string;
}

export interface MyFarm {
    farm_id: string;
    farm_name: string;
    farm_description: string;
    farm_location: string;
    user_role: UserRole;
    mcu_count: number;
    survey_point_count: number;
    online_mcu_count: number;
    created_at: string;
}

export interface FarmOverview {
    farm_id: string;
    farm_name: string;
    farm_description: string;
    farm_location: string;
    total_mcus: number;
    online_mcus: number;
    offline_mcus: number;
    total_survey_points: number;
    connecting_points: number;
    connected_points: number;
    disconnected_points: number;
    created_at: string;
}

export interface FarmStructure {
    farm_id: string;
    farm_name: string;
    mcu_id: string | null;
    mcu_code: string | null;
    mcu_status: string | null;
    survey_point_id: string | null;
    survey_point_name: string | null;
    survey_point_status: string | null;
}

export interface AddUserToFarmRequest {
    user_id: string;
    role: UserRole;
}
