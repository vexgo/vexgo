import { useState, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { api } from '../api/client'

export default function PostEdit() {
  const { id } = useParams()
  const navigate = useNavigate()
  const [form, setForm] = useState({ title: '', content: '', status: 'published', meta_title: '', meta_description: '', og_image: '', tags: '', category_id: '' })
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    if (id) loadPost()
  }, [id])

  async function loadPost() {
    const data = await api.get(`/posts/${id}`)
    setForm({ title: data.title || '', content: data.content || '', status: data.status || 'published', meta_title: data.meta_title || '', meta_description: data.meta_description || '', og_image: data.og_image || '', tags: data.tags || '', category_id: data.category_id || '' })
  }

  function update(field, value) {
    setForm(f => ({ ...f, [field]: value }))
  }

  async function handleSubmit(e) {
    e.preventDefault()
    setLoading(true)
    try {
      if (id) {
        await api.put(`/posts/${id}`, form)
      } else {
        await api.post('/posts', form)
      }
      navigate('/posts')
    } catch (err) {
      alert(err.message)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div>
      <h2>{id ? 'Edit Post' : 'New Post'}</h2>
      <form onSubmit={handleSubmit} style={{ background: '#fff', padding: 20, borderRadius: 4 }}>
        <div style={{ marginBottom: 12 }}>
          <label>Title<br/><input value={form.title} onChange={e => update('title', e.target.value)} required style={{ width: '100%', padding: 8 }} /></label>
        </div>
        <div style={{ marginBottom: 12 }}>
          <label>Content<br/><textarea value={form.content} onChange={e => update('content', e.target.value)} rows={10} style={{ width: '100%', padding: 8 }} /></label>
        </div>
        <div style={{ marginBottom: 12 }}>
          <label>Status<br/>
            <select value={form.status} onChange={e => update('status', e.target.value)} style={{ padding: 8 }}>
              <option value="published">Published</option>
              <option value="draft">Draft</option>
            </select>
          </label>
        </div>
        <div style={{ marginBottom: 12 }}>
          <label>Meta Title<br/><input value={form.meta_title} onChange={e => update('meta_title', e.target.value)} style={{ width: '100%', padding: 8 }} /></label>
        </div>
        <div style={{ marginBottom: 12 }}>
          <label>Meta Description<br/><textarea value={form.meta_description} onChange={e => update('meta_description', e.target.value)} rows={3} style={{ width: '100%', padding: 8 }} /></label>
        </div>
        <div style={{ marginBottom: 12 }}>
          <label>OG Image URL<br/><input value={form.og_image} onChange={e => update('og_image', e.target.value)} style={{ width: '100%', padding: 8 }} /></label>
        </div>
        <div style={{ marginBottom: 12 }}>
          <label>Tags<br/><input value={form.tags} onChange={e => update('tags', e.target.value)} placeholder="comma,separated" style={{ width: '100%', padding: 8 }} /></label>
        </div>
        <div style={{ marginBottom: 12 }}>
          <label>Category ID<br/><input type="number" value={form.category_id} onChange={e => update('category_id', e.target.value)} style={{ padding: 8 }} /></label>
        </div>
        <button type="submit" disabled={loading} style={{ padding: '10px 20px', background: '#00d4ff', color: '#fff', border: 'none', borderRadius: 4, cursor: 'pointer' }}>{loading ? 'Saving...' : 'Save'}</button>
        <button type="button" onClick={() => navigate('/posts')} style={{ marginLeft: 8, padding: '10px 20px', cursor: 'pointer' }}>Cancel</button>
      </form>
    </div>
  )
}
