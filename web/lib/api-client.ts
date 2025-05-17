import { useAuthStore } from "@/store/auth-store"
import { Api } from "./Api"
import axios from "axios"
import { parseApiError } from "./error-handler"

// Configure axios interceptors for global error handling
const axiosInstance = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_URL,
})

// Response interceptor to standardize error handling
axiosInstance.interceptors.response.use(
  response => response, 
  error => {
    // Parse the error into our standard format
    const standardError = parseApiError(error);
    
    // Handle 401 errors which may require logout
    if (standardError.statusCode === 401) {
      // Check if this is not a login or token validation endpoint
      const url = error.config?.url;
      if (url && !url.includes('/auth/login') && !url.includes('/auth/validate')) {
        // Clear auth if the token is invalid
        useAuthStore.getState().clearAuth();
      }
    }
    
    // Add the standardized error to the original error object
    error.standardError = standardError;
    
    return Promise.reject(error);
  }
);

// Create API client
export const apiClient = new Api<string>({
  baseURL: process.env.NEXT_PUBLIC_API_URL,
  securityWorker: () => {
    const token = useAuthStore.getState().token;
    return {
      headers: {
        Authorization: token ? `Bearer ${token}` : "",
      },
    }
  },
})

// Replace the default axios instance with our custom one
apiClient.instance = axiosInstance;
