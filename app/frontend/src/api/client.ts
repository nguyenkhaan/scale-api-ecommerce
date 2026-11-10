import axios from 'axios'

const backendUrl = (import.meta.env.VITE_BASE_URL ?? '').replace(/\/+$/, '')
const API_BASE_URL = backendUrl.endsWith('/api')
  ? backendUrl
  : `${backendUrl}/api`

const createClient = () =>
  axios.create({
    baseURL: API_BASE_URL,
    headers: {
      'Content-Type': 'application/json',
    },
    timeout: 30_000,
    withCredentials: true,
  })

export const publicClient = createClient()
export const privateClient = createClient()
