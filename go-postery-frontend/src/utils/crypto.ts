import CryptoJS from 'crypto-js'

/**
 * 对字符串进行MD5哈希，返回32位小写哈希值
 * @param input 输入字符串
 * @returns 32位MD5哈希值（小写）
 */
export function md5Hash(input: string): string {
  return CryptoJS.MD5(input).toString().toLowerCase()
}