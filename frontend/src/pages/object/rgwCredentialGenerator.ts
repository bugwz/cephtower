export type RGWCredentialKind = 'access' | 'secret'

// Alphabet sizes divide 256, so masking random bytes introduces no modulo bias.
export function generateRGWCredential(kind: RGWCredentialKind, random: Pick<Crypto, 'getRandomValues'> | undefined = globalThis.crypto): string {
  if (!random?.getRandomValues) throw new Error('安全随机数不可用')
  if (kind !== 'access' && kind !== 'secret') throw new Error('不支持的凭据类型')
  const alphabet = kind === 'access' ? 'ABCDEFGHIJKLMNOPQRSTUVWXYZ234567' : 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_'
  const bytes = new Uint8Array(kind === 'access' ? 20 : 40)
  try {
    random.getRandomValues(bytes)
    return Array.from(bytes, byte => alphabet[byte & (alphabet.length - 1)]).join('')
  } finally {
    bytes.fill(0)
  }
}
