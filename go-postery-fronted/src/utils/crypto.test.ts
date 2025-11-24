import { md5Hash } from './crypto'

describe('md5Hash', () => {
  it('should hash password correctly', () => {
    const password = 'test123'
    const hashed = md5Hash(password)
    
    // MD5 hash of 'test123' should be 'cc03e747a6afbbcbf8be7668acfebee5'
    expect(hashed).toBe('cc03e747a6afbbcbf8be7668acfebee5')
    expect(hashed.length).toBe(32) // MD5 produces 32 character hex string
    expect(hashed).toMatch(/^[a-f0-9]{32}$/) // Should be lowercase hex
  })

  it('should produce consistent results', () => {
    const password = 'myPassword123'
    const hash1 = md5Hash(password)
    const hash2 = md5Hash(password)
    
    expect(hash1).toBe(hash2)
  })

  it('should handle empty string', () => {
    const hashed = md5Hash('')
    expect(hashed).toBe('d41d8cd98f00b204e9800998ecf8427e') // MD5 of empty string
    expect(hashed.length).toBe(32)
  })
})