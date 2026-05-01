import { createBrowserRouter, RouterProvider, Navigate } from 'react-router-dom'
import Login from './components/Login'
import Layout from './components/Layout'
import Posts from './components/Posts'
import PostEdit from './components/PostEdit'
import Pages from './components/Pages'
import PageEdit from './components/PageEdit'
import Media from './components/Media'
import Settings from './components/Settings'
import { isAuthenticated } from './utils/auth'

function AuthGuard({ children }) {
  if (!isAuthenticated()) {
    return <Navigate to="/login" replace />
  }
  return children
}

const router = createBrowserRouter([
  {
    path: '/login',
    element: <Login />,
  },
  {
    path: '/',
    element: (
      <AuthGuard>
        <Layout />
      </AuthGuard>
    ),
    children: [
      { index: true, element: <Navigate to="/posts" replace /> },
      { path: 'posts', element: <Posts /> },
      { path: 'posts/new', element: <PostEdit /> },
      { path: 'posts/:id', element: <PostEdit /> },
      { path: 'pages', element: <Pages /> },
      { path: 'pages/new', element: <PageEdit /> },
      { path: 'pages/:id', element: <PageEdit /> },
      { path: 'media', element: <Media /> },
      { path: 'settings', element: <Settings /> },
    ],
  },
],
  {
    basename: '/admin',
  }
)

export default function App() {
  return <RouterProvider router={router} />
}
