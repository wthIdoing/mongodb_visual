import axios from 'axios'
import { copy } from '../copy'
import { buildMongoSessionHeaders, getMongoSessionConnection } from '../session/mongoConnection'

export const http = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL ?? '',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
})

http.interceptors.request.use((config) => {
  const connection = getMongoSessionConnection()
  Object.entries(buildMongoSessionHeaders(connection)).forEach(([key, value]) => {
    if (value) {
      config.headers.set(key, value)
    }
  })
  return config
})

http.interceptors.response.use(
  (response) => response,
  (error) => {
    const message = error.response?.data?.message ?? error.message ?? copy.http.requestFailed
    return Promise.reject(new Error(message))
  },
)
