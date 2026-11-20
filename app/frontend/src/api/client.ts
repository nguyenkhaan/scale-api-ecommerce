import axios from 'axios'
import { API_BASE_URL } from './base-url'

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
