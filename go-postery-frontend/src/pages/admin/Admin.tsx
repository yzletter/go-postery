import { Navigate, Route, Routes } from 'react-router-dom'
import AdminLayout from './AdminLayout'
import AdminPosts from './AdminPosts'
import AdminComments from './AdminComments'
import AdminUsers from './AdminUsers'

export default function Admin() {
  return (
    <Routes>
      <Route element={<AdminLayout />}>
        <Route index element={<Navigate to="posts" replace />} />
        <Route path="posts" element={<AdminPosts />} />
        <Route path="comments" element={<AdminComments />} />
        <Route path="users" element={<AdminUsers />} />
        <Route path="*" element={<Navigate to="posts" replace />} />
      </Route>
    </Routes>
  )
}

