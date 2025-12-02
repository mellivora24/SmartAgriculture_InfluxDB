export interface User {
    id: string;
    username: string;
    email: string;
    phone: string;
    full_name: string;
    created_at: string;
}

export interface RegisterRequest {
    username: string;
    email: string;
    phone: string;
    password: string;
    full_name: string;
}

export interface LoginRequest {
    username: string;
    password: string;
}

export interface AuthResponse extends User {
    token: string;
}

export interface UpdateUserRequest {
    email?: string;
    phone?: string;
    full_name?: string;
}
