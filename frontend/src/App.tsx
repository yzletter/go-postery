import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom'
import { AuthProvider, useAuth } from './contexts/AuthContext'
import Navbar from './components/Navbar'
import Home from './pages/Home'
import PostDetail from './pages/PostDetail'
import CreatePost from './pages/CreatePost'
import EditPost from './pages/EditPost'
import Login from './pages/Login'
import Profile from './pages/Profile'
import Settings from './pages/Settings'
import InterviewAgent from './pages/InterviewAgent'
import Follows from './pages/Follows'
import Messages from './pages/Messages'
import Search from './pages/Search'
import Lottery from './pages/Lottery'

const LoadingScreen = () => (
  <div className="min-h-screen flex items-center justify-center">
    <div className="w-8 h-8 border-4 border-primary-600 border-t-transparent rounded-full animate-spin" />
  </div>
)

type AuthRouteProps = {
  children: React.ReactElement
}

// 保护需要登录的路由
function AuthRoute({ children }: AuthRouteProps) {
  const { user, isLoading } = useAuth()

  if (isLoading) {
    return <LoadingScreen />
  }

  if (!user) {
    return <Navigate to="/login" replace />
  }

  return children
}

function AppRoutes() {
  const protectedRoutes = [
    { path: '/search', element: <Search /> },
    { path: '/follows', element: <Follows /> },
    { path: '/messages', element: <Messages /> },
    { path: '/create', element: <CreatePost /> },
    { path: '/edit/:id', element: <EditPost /> },
    { path: '/profile', element: <Profile /> },
    { path: '/settings', element: <Settings /> },
    { path: '/interview-agent', element: <InterviewAgent /> },
  ]

  return (
    <Routes>
      <Route path="/" element={<Home />} />
      <Route path="/post/:id" element={<PostDetail />} />
      <Route path="/postDetailDTO/:id" element={<PostDetail />} />
      <Route path="/login" element={<Login />} />
      <Route path="/agent" element={<Navigate to="/interview-agent" replace />} />
      <Route path="/lottery" element={<Lottery />} />
      <Route path="/admin/*" element={<Navigate to="/" replace />} />
      {protectedRoutes.map(({ path, element }) => (
        <Route
          key={path}
          path={path}
          element={<AuthRoute>{element}</AuthRoute>}
        />
      ))}
      <Route path="/users/:userId" element={<Profile />} />
    </Routes>
  )
}

function App() {
  return (
    <AuthProvider>
      <Router>
        <div className="min-h-screen flex flex-col">
          <Navbar />
          <main className="flex-1 max-w-7xl mx-auto w-full px-4 sm:px-6 lg:px-8 py-8">
            <AppRoutes />
          </main>
        </div>
      </Router>
    </AuthProvider>
  )
}

export default App
