import type { SurveyPointStatus } from "./common";

export interface SurveyPoint {
    id: string;
    mcu_id: string;
    name: string;
    description: string;
    status: SurveyPointStatus;
    created_at: string;
    updated_at: string;
}

export interface CreateSurveyPointRequest {
    mcu_id: string;
    name: string;
    description: string;
}

export interface UpdateSurveyPointRequest {
    name?: string;
    description?: string;
}

export interface UpdateSurveyPointStatusRequest {
    status: SurveyPointStatus;
}

export interface SurveyPointListItem {
    survey_point_id: string;
    survey_point_name: string;
    description: string;
    status: SurveyPointStatus;
    created_at: string;
    updated_at: string;
}
