import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { setToken } from '../utils/auth'
import { api } from '../api/client'

export default function Login() {
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const navigate = useNavigate()

  async function handleSubmit(e) {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      const data = await api.post('/login', { username, password })
      setToken(data.token)
      navigate('/posts/', { replace: true })
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div style={{ maxWidth: 400, margin: '100px auto', padding: 20 }}>
      <h2>go-cms Admin Login</h2>
      {error && <p style={{ color: 'red' }}>{error}</p>}
      <form onSubmit={handleSubmit}>
        <div style={{ marginBottom: 12 }}>
          <label>Username<br/><input value={username} onChange={e => setUsername(e.target.value)} style={{ width: '100%', padding: 8 }} /></label>
        </div>
        <div style={{ marginBottom: 12 }}>
          <label>Password<br/><input type="password" value={password} onChange={e => setPassword(e.target.value)} style={{ width: '100%', padding: 8 }} /></label>
        </div>
        <button type="submit" disabled={loading} style={{ width: '100%', padding: 10 }}>{loading ? 'Logging in...' : 'Login'}</button>
      </form>
    </div>
  )
}
