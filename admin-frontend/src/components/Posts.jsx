import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { api } from '../api/client'

export default function Posts() {
  const [posts, setPosts] = useState([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const pageSize = 10

  async function load() {
    const data = await api.get(`/posts?page=${page}&page_size=${pageSize}`)
    setPosts(data.data || data)
    setTotal(data.total || data.length)
  }

  useEffect(() => { load() }, [page])

  async function handleDelete(id) {
    if (!confirm('Delete this post?')) return
    await api.del(`/posts/${id}`)
    load()
  }

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <h2>Posts</h2>
        <Link to="/posts/new" style={{ padding: '8px 16px', background: '#00d4ff', color: '#fff', textDecoration: 'none', borderRadius: 4 }}>New Post</Link>
      </div>
      <table style={{ width: '100%', borderCollapse: 'collapse', background: '#fff' }}>
        <thead><tr><th style={{ textAlign: 'left', padding: 8 }}>Title</th><th style={{ textAlign: 'left', padding: 8 }}>Status</th><th style={{ textAlign: 'left', padding: 8 }}>Slug</th><th style={{ textAlign: 'left', padding: 8 }}>Actions</th></tr></thead>
        <tbody>
          {posts.map(p => (
            <tr key={p.id} style={{ borderTop: '1px solid #eee' }}>
              <td style={{ padding: 8 }}>{p.title}</td>
              <td style={{ padding: 8 }}>{p.status}</td>
              <td style={{ padding: 8 }}>{p.slug}</td>
              <td style={{ padding: 8 }}>
                <Link to={`/posts/${p.id}`} style={{ marginRight: 8 }}>Edit</Link>
                <button onClick={() => handleDelete(p.id)} style={{ color: 'red', border: 'none', background: 'none', cursor: 'pointer' }}>Delete</button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
      <div style={{ marginTop: 12 }}>
        <button disabled={page <= 1} onClick={() => setPage(p => p - 1)}>Prev</button>
        <span style={{ margin: '0 8px' }}>Page {page}</span>
        <button disabled={posts.length < pageSize} onClick={() => setPage(p => p + 1)}>Next</button>
      </div>
    </div>
  )
}
