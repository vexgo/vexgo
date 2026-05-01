import { Outlet, NavLink, useNavigate } from 'react-router-dom'
import { removeToken } from '../utils/auth'

const links = [
  { to: '/posts', label: 'Posts' },
  { to: '/pages', label: 'Pages' },
  { to: '/media', label: 'Media' },
  { to: '/settings', label: 'Settings' },
]

export default function Layout() {
  const navigate = useNavigate()
  function logout() {
    removeToken()
    navigate('/login', { replace: true })
  }
  return (
    <div style={{ display: 'flex', minHeight: '100vh' }}>
      <aside style={{ width: 200, background: '#1a1a2e', color: '#fff', padding: 20 }}>
        <h3 style={{ marginTop: 0 }}>go-cms</h3>
        {links.map(l => (
          <NavLink key={l.to} to={l.to} style={({ isActive }) => ({ display: 'block', padding: '8px 0', color: isActive ? '#00d4ff' : '#ccc', textDecoration: 'none' })}>{l.label}</NavLink>
        ))}
        <button onClick={logout} style={{ marginTop: 40, width: '100%', padding: 8, cursor: 'pointer' }}>Logout</button>
      </aside>
      <main style={{ flex: 1, padding: 24, background: '#f5f5f5' }}>
        <Outlet />
      </main>
    </div>
  )
}
