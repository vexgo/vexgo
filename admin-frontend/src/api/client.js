import { getToken, removeToken } from '../utils/auth'

const API_BASE = '/api'

function getHeaders() {
  const headers = { 'Content-Type': 'application/json' }
  const token = getToken()
  if (token) headers['Authorization'] = 'Bearer ' + token
  return headers
}

export async function request(path, options = {}) {
  const res = await fetch(API_BASE + path, {
    ...options,
    headers: { ...getHeaders(), ...(options.headers || {}) },
  })
  if (res.status === 401) {
    removeToken()
    window.location.href = '/admin/login'
    return
  }
  const data = await res.json()
  if (!res.ok) throw new Error(data.error || 'Request failed')
  return data
}

export const api = {
  get: (path) => request(path),
  post: (path, body) => request(path, { method: 'POST', body: JSON.stringify(body) }),
  put: (path, body) => request(path, { method: 'PUT', body: JSON.stringify(body) }),
  del: (path) => request(path, { method: 'DELETE' }),
}
