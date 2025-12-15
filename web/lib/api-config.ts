import { useAuthStore } from "@/store/auth-store";
import type { CreateClientConfig } from "./api/client.gen";

export const createClientConfig: CreateClientConfig = (config) => ({
  ...config,
  baseURL: process.env.NEXT_PUBLIC_API_URL,
  withCredentials: true,
  auth: () => {
    const token = useAuthStore.getState().token;
    return token ? `Bearer ${token}` : "";
  },
});
