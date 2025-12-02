import axios from "@/core/api/axios";
import type { LoginRequest, RegisterRequest, AuthResponse } from "@/core/types";

export const authApi = {
  register: async (data: RegisterRequest): Promise<AuthResponse> => {
    const response = await axios.post<{ data: AuthResponse }>(
      "/users/register",
      data
    );
    return response.data.data;
  },

  login: async (data: LoginRequest): Promise<AuthResponse> => {
    const response = await axios.post<{ data: AuthResponse }>(
      "/users/login",
      data
    );
    return response.data.data;
  },
};
