import { Api } from "./Api"

// Создаем экземпляр API клиента
export const apiClient = new Api<string>({
  baseURL: process.env.NEXT_PUBLIC_API_URL,
  securityWorker: (token) => {
    return {
      headers: {
        Authorization: token ? `Bearer ${token}` : "",
      },
    }
  },
})
