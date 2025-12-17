import { getToken } from "@/lib/auth-storage";
import type { CreateClientConfig } from "./api/client.gen";

export const createClientConfig: CreateClientConfig = (config) => ({
  ...config,
  baseURL: process.env.NEXT_PUBLIC_API_URL,
  withCredentials: true,
  auth: () => {
    const token = getToken();
    return token ? `Bearer ${token}` : "";
  },
});
