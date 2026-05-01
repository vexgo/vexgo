import { useState, useEffect } from 'react'
import { api } from '../api/client'

export default function Settings() {
  const [settings, setSettings] = useState({})
  const [loading, setLoading] = useState(false)

  useEffect(() => { load() }, [])

  async function load() {
    const data = await api.get('/settings')
    const map = {}
    ;(data.data || data).forEach(s => { map[s.key] = s.value })
    setSettings(map)
  }

  async function handleSave(e) {
    e.preventDefault()
    setLoading(true)
    try {
      for (const [key, value] of Object.entries(settings)) {
        await api.put('/settings', { key, value })
      }
      alert('Settings saved!')
    } catch (err) {
      alert(err.message)
    } finally {
      setLoading(false)
    }
  }

  function update(key, value) {
    setSettings(s => ({ ...s, [key]: value }))
  }

  return (
    <div>
      <h2>Settings</h2>
      <form onSubmit={handleSave} style={{ background: '#fff', padding: 20, borderRadius: 4 }}>
        <div style={{ marginBottom: 12 }}>
          <label>Site Title<br/><input value={settings.site_title || ''} onChange={e => update('site_title', e.target.value)} style={{ width: '100%', padding: 8 }} /></label>
        </div>
        <div style={{ marginBottom: 12 }}>
          <label>Site Description<br/><textarea value={settings.site_description || ''} onChange={e => update('site_description', e.target.value)} rows={3} style={{ width: '100%', padding: 8 }} /></label>
        </div>
        <div style={{ marginBottom: 12 }}>
          <label>Posts Per Page<br/><input type="number" value={settings.posts_per_page || 10} onChange={e => update('posts_per_page', e.target.value)} style={{ padding: 8 }} /></label>
        </div>
        <button type="submit" disabled={loading} style={{ padding: '10px 20px', background: '#00d4ff', color: '#fff', border: 'none', borderRadius: 4, cursor: 'pointer' }}>{loading ? 'Saving...' : 'Save Settings'}</button>
      </form>
    </div>
  )
}
