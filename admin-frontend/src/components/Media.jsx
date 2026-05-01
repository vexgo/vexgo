import { useState, useEffect } from 'react'
import { api } from '../api/client'

export default function Media() {
  const [files, setFiles] = useState([])
  const [uploading, setUploading] = useState(false)

  async function load() {
    const data = await api.get('/media')
    setFiles(data.data || data)
  }

  useEffect(() => { load() }, [])

  async function handleUpload(e) {
    const file = e.target.files[0]
    if (!file) return
    setUploading(true)
    const formData = new FormData()
    formData.append('file', file)
    try {
      await fetch('/api/media/upload', {
        method: 'POST',
        headers: { 'Authorization': 'Bearer ' + localStorage.getItem('cms_token') },
        body: formData,
      })
      load()
    } catch (err) {
      alert(err.message)
    } finally {
      setUploading(false)
    }
  }

  async function handleDelete(id) {
    if (!confirm('Delete this file?')) return
    await api.del(`/media/${id}`)
    load()
  }

  return (
    <div>
      <h2>Media Library</h2>
      <div style={{ background: '#fff', padding: 20, borderRadius: 4, marginBottom: 20 }}>
        <input type="file" onChange={handleUpload} disabled={uploading} />
        {uploading && <span style={{ marginLeft: 8 }}>Uploading...</span>}
      </div>
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(150px, 1fr))', gap: 16 }}>
        {files.map(f => (
          <div key={f.id} style={{ background: '#fff', padding: 12, borderRadius: 4, position: 'relative' }}>
            <img src={f.url} alt={f.filename} style={{ width: '100%', height: 100, objectFit: 'cover' }} />
            <p style={{ margin: '8px 0 4px', fontSize: 12, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{f.filename}</p>
            <button onClick={() => handleDelete(f.id)} style={{ color: 'red', border: 'none', background: 'none', cursor: 'pointer', fontSize: 12 }}>Delete</button>
          </div>
        ))}
      </div>
    </div>
  )
}
