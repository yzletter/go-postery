import { useState } from 'react'
import { md5Hash } from '../utils/crypto'

export default function PasswordHashDemo() {
  const [password, setPassword] = useState('')
  const [hashedPassword, setHashedPassword] = useState('')

  const handlePasswordChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const newPassword = e.target.value
    setPassword(newPassword)
    if (newPassword) {
      setHashedPassword(md5Hash(newPassword))
    } else {
      setHashedPassword('')
    }
  }

  return (
    <div className="p-4 bg-gray-50 rounded-lg">
      <h3 className="text-lg font-semibold mb-4">密码MD5哈希演示</h3>
      <div className="space-y-4">
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            输入密码
          </label>
          <input
            type="password"
            value={password}
            onChange={handlePasswordChange}
            className="input"
            placeholder="输入密码查看MD5哈希"
          />
        </div>
        {hashedPassword && (
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              MD5哈希值（32位小写）
            </label>
            <div className="p-3 bg-gray-100 rounded-md font-mono text-sm break-all">
              {hashedPassword}
            </div>
            <p className="text-xs text-gray-500 mt-2">
              长度: {hashedPassword.length} 字符
            </p>
          </div>
        )}
      </div>
    </div>
  )
}