#!/bin/bash
# LCA官网文件生成脚本
# 用法: bash scripts/setup-website.sh
# 在 /Users/cj/Documents/lca_web 目录下生成所有官网源码文件

set -e
WEB_DIR="/Users/cj/Documents/lca_web"

echo "=== 生成LCA官网源码文件 ==="

# ---- src/lib/auth.ts ----
mkdir -p "$WEB_DIR/src/lib"
cat > "$WEB_DIR/src/lib/auth.ts" << 'FILEOF'
import { SignJWT, jwtVerify } from 'jose'
const secret = new TextEncoder().encode(process.env.JWT_SECRET || 'lca-web-secret')
export async function signToken(payload: { userId: number; email: string }): Promise<string> {
  return new SignJWT(payload as unknown as Record<string, unknown>).setProtectedHeader({ alg: 'HS256' }).setExpirationTime('7d').sign(secret)
}
export async function verifyToken(token: string) {
  const { payload } = await jwtVerify(token, secret)
  return payload as unknown as { userId: number; email: string }
}
FILEOF
echo "  auth.ts OK"

# ---- src/lib/license.ts ----
cat > "$WEB_DIR/src/lib/license.ts" << 'FILEOF'
import { readFileSync } from 'fs'
import { join } from 'path'

interface LicensePayload {
  product: string
  machine_id: string
  type: string
  issued_at: number
  expires_at: number
}

let cachedPrivateKey: CryptoKey | null = null

async function getPrivateKey(): Promise<CryptoKey> {
  if (cachedPrivateKey) return cachedPrivateKey
  const keyPath = join(process.cwd(), 'keys', 'private.pem')
  const pem = readFileSync(keyPath, 'utf-8')
  const pemContent = pem.replace('-----BEGIN PRIVATE KEY-----', '').replace('-----END PRIVATE KEY-----', '').replace(/\s/g, '')
  const keyBuffer = Buffer.from(pemContent, 'base64')
  const key = await crypto.subtle.importKey('pkcs8', keyBuffer, { name: 'RSASSA-PKCS1-v1_5', hash: 'SHA-256' }, false, ['sign'])
  cachedPrivateKey = key
  return key
}

export async function generateLicenseKey(machineId: string, type: 'monthly' | 'yearly' | 'permanent'): Promise<string> {
  const now = Math.floor(Date.now() / 1000)
  let expiresAt = 0
  if (type === 'monthly') expiresAt = now + 30 * 24 * 3600
  else if (type === 'yearly') expiresAt = now + 365 * 24 * 3600
  const payload: LicensePayload = { product: 'lca', machine_id: machineId, type, issued_at: now, expires_at: expiresAt }
  const payloadB64 = Buffer.from(JSON.stringify(payload)).toString('base64')
  const privateKey = await getPrivateKey()
  const signature = await crypto.subtle.sign('RSASSA-PKCS1-v1_5', privateKey, Buffer.from(payloadB64))
  const sigB64 = Buffer.from(signature).toString('base64')
  return `${payloadB64}.${sigB64}`
}
FILEOF
echo "  license.ts OK"

# ---- src/lib/wechat-pay.ts ----
cat > "$WEB_DIR/src/lib/wechat-pay.ts" << 'FILEOF'
const BASE_URL = 'https://api.mch.weixin.qq.com'
interface CreateOrderParams { orderId: string; description: string; amount: number; notifyUrl: string }
export async function createNativeOrder(params: CreateOrderParams) {
  // TODO: 对接微信支付V3 Native下单API, 待商户资质审核后替换
  return { code_url: `weixin://wxpay/bizpayurl?pr=${params.orderId}`, order_id: params.orderId }
}
export function verifyWechatNotification(body: string, signature: string, timestamp: string, nonce: string): boolean {
  // TODO: 验证微信支付回调签名
  return true
}
FILEOF
echo "  wechat-pay.ts OK"

# ---- src/app/page.tsx ----
cat > "$WEB_DIR/src/app/page.tsx" << 'FILEOF'
import Link from 'next/link'

export default function Home() {
  return (
    <div className="min-h-screen">
      <nav className="fixed top-0 w-full z-50 border-b border-[#1e2d45] bg-[#0a1628]/90 backdrop-blur-md">
        <div className="max-w-7xl mx-auto px-6 h-16 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="w-9 h-9 rounded-lg bg-gradient-to-br from-[#00d4ff] to-[#0099cc] flex items-center justify-center text-[#0a1628] font-extrabold text-lg">L</div>
            <span className="text-xl font-bold bg-gradient-to-r from-[#00d4ff] to-[#0099cc] bg-clip-text text-transparent">LCA</span>
          </div>
          <div className="flex items-center gap-6">
            <Link href="/login" className="text-sm text-[#8899aa] hover:text-[#00d4ff] transition">登录</Link>
            <Link href="/register" className="text-sm px-4 py-2 rounded-lg bg-gradient-to-r from-[#00d4ff] to-[#0099cc] text-[#0a1628] font-semibold">注册</Link>
          </div>
        </div>
      </nav>

      <section className="pt-32 pb-20 px-6 text-center relative overflow-hidden">
        <div className="absolute inset-0 bg-[radial-gradient(ellipse_at_center,rgba(0,212,255,0.08),transparent_70%)]" />
        <div className="max-w-4xl mx-auto relative">
          <h1 className="text-5xl md:text-6xl font-extrabold mb-6 leading-tight">
            <span className="bg-gradient-to-r from-[#00d4ff] to-[#6366f1] bg-clip-text text-transparent">日志收集</span>
            <span className="text-[#e8edf5]">智能分析平台</span>
          </h1>
          <p className="text-lg text-[#8899aa] mb-10 max-w-2xl mx-auto">LCA是企业级日志采集、智能分析与可视化平台</p>
          <div className="flex gap-4 justify-center">
            <Link href="/register" className="px-8 py-3 rounded-lg bg-gradient-to-r from-[#00d4ff] to-[#0099cc] text-[#0a1628] font-bold text-lg">立即开始</Link>
            <a href="#features" className="px-8 py-3 rounded-lg border border-[#1e2d45] text-[#e8edf5] font-semibold text-lg hover:border-[#00d4ff] transition">了解更多</a>
          </div>
        </div>
      </section>

      <section id="features" className="py-20 px-6">
        <div className="max-w-6xl mx-auto">
          <h2 className="text-3xl font-bold text-center mb-12">核心特性</h2>
          <div className="grid md:grid-cols-4 gap-6">
            {[
              { icon: '📡', title: '实时采集', desc: '多源日志实时采集，Agent分布式部署' },
              { icon: '🔍', title: '智能分析', desc: 'Elasticsearch全文检索，AI异常检测' },
              { icon: '📊', title: '多维可视化', desc: '丰富图表仪表盘，日志趋势热力图' },
              { icon: '🛡️', title: '安全可靠', desc: 'RBAC权限控制，数据备份与恢复' },
            ].map((f) => (
              <div key={f.title} className="p-6 rounded-xl bg-[#111d31] border border-[#1e2d45] hover:border-[#00d4ff]/30 transition">
                <div className="text-3xl mb-4">{f.icon}</div>
                <h3 className="text-lg font-semibold mb-2">{f.title}</h3>
                <p className="text-sm text-[#8899aa]">{f.desc}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      <section id="pricing" className="py-20 px-6">
        <div className="max-w-5xl mx-auto">
          <h2 className="text-3xl font-bold text-center mb-4">选择方案</h2>
          <p className="text-center text-[#8899aa] mb-12">按需选择，灵活升级</p>
          <div className="grid md:grid-cols-3 gap-6">
            {[
              { type: 'monthly', name: '月付', price: '9.9', unit: '/月', features: ['全功能使用', '社区支持'] },
              { type: 'yearly', name: '年付', hot: true, price: '99', unit: '/年', features: ['全功能使用', '优先技术支持', '省17%'] },
              { type: 'permanent', name: '永久', price: '699', unit: '一次', features: ['终身使用', '专属技术支持', '最优性价比'] },
            ].map((p) => (
              <div key={p.type} className={`p-8 rounded-xl border ${p.hot ? 'border-[#00d4ff] bg-[#111d31] shadow-[0_0_30px_rgba(0,212,255,0.15)]' : 'border-[#1e2d45] bg-[#111d31]'} flex flex-col`}>
                {p.hot && <div className="text-xs font-bold text-[#0a1628] bg-[#00d4ff] rounded-full px-3 py-1 self-start mb-4">推荐</div>}
                <h3 className="text-xl font-semibold mb-2">{p.name}</h3>
                <div className="mb-6"><span className="text-4xl font-extrabold text-[#00d4ff]">¥{p.price}</span><span className="text-[#8899aa] text-sm">{p.unit}</span></div>
                <ul className="text-sm text-[#8899aa] space-y-2 mb-8 flex-1">{p.features.map((f) => (<li key={f} className="flex items-center gap-2"><span className="text-[#00d4ff]">✓</span>{f}</li>))}</ul>
                <Link href="/register" className="w-full py-3 rounded-lg text-center font-semibold bg-gradient-to-r from-[#00d4ff] to-[#0099cc] text-[#0a1628]">立即购买</Link>
              </div>
            ))}
          </div>
        </div>
      </section>

      <footer className="py-8 border-t border-[#1e2d45] text-center text-sm text-[#8899aa]">LCA Log Analytics Platform &copy; 2024-2026</footer>
    </div>
  )
}
FILEOF
echo "  page.tsx OK"

# ---- API Routes ----
mkdir -p "$WEB_DIR/src/app/api/auth/register" "$WEB_DIR/src/app/api/auth/login" "$WEB_DIR/src/app/api/license/generate" "$WEB_DIR/src/app/api/license/list" "$WEB_DIR/src/app/api/payment/create" "$WEB_DIR/src/app/api/payment/notify" "$WEB_DIR/src/app/register" "$WEB_DIR/src/app/login" "$WEB_DIR/src/app/dashboard"

cat > "$WEB_DIR/src/app/api/auth/register/route.ts" << 'FILEOF'
import { NextRequest, NextResponse } from 'next/server'
import { prisma } from '@/lib/db'
import bcrypt from 'bcryptjs'
export async function POST(req: NextRequest) {
  try {
    const { email, password } = await req.json()
    if (!email || !password || password.length < 6) return NextResponse.json({ error: '邮箱和密码(至少6位)必填' }, { status: 400 })
    const exists = await prisma.user.findUnique({ where: { email } })
    if (exists) return NextResponse.json({ error: '邮箱已注册' }, { status: 400 })
    const hash = await bcrypt.hash(password, 10)
    const user = await prisma.user.create({ data: { email, password: hash } })
    return NextResponse.json({ id: user.id, email: user.email })
  } catch (e: unknown) { return NextResponse.json({ error: (e as Error).message }, { status: 500 }) }
}
FILEOF
echo "  api/auth/register OK"

cat > "$WEB_DIR/src/app/api/auth/login/route.ts" << 'FILEOF'
import { NextRequest, NextResponse } from 'next/server'
import { prisma } from '@/lib/db'
import bcrypt from 'bcryptjs'
import { signToken } from '@/lib/auth'
export async function POST(req: NextRequest) {
  try {
    const { email, password } = await req.json()
    if (!email || !password) return NextResponse.json({ error: '邮箱和密码必填' }, { status: 400 })
    const user = await prisma.user.findUnique({ where: { email } })
    if (!user) return NextResponse.json({ error: '邮箱或密码错误' }, { status: 401 })
    const valid = await bcrypt.compare(password, user.password)
    if (!valid) return NextResponse.json({ error: '邮箱或密码错误' }, { status: 401 })
    const token = await signToken({ userId: user.id, email: user.email })
    return NextResponse.json({ token, user: { id: user.id, email: user.email } })
  } catch (e: unknown) { return NextResponse.json({ error: (e as Error).message }, { status: 500 }) }
}
FILEOF
echo "  api/auth/login OK"

cat > "$WEB_DIR/src/app/api/license/generate/route.ts" << 'FILEOF'
import { NextRequest, NextResponse } from 'next/server'
import { prisma } from '@/lib/db'
import { verifyToken } from '@/lib/auth'
import { generateLicenseKey } from '@/lib/license'
export async function POST(req: NextRequest) {
  try {
    const auth = req.headers.get('authorization')
    if (!auth?.startsWith('Bearer ')) return NextResponse.json({ error: '未登录' }, { status: 401 })
    const { userId } = await verifyToken(auth.slice(7))
    const { machine_id, type, order_id } = await req.json()
    if (!machine_id || !type) return NextResponse.json({ error: 'machine_id和type必填' }, { status: 400 })
    if (!['monthly', 'yearly', 'permanent'].includes(type)) return NextResponse.json({ error: 'type无效' }, { status: 400 })
    const licenseKey = await generateLicenseKey(machine_id, type)
    const now = new Date()
    let expiresAt: Date | null = null
    if (type === 'monthly') expiresAt = new Date(now.getTime() + 30 * 24 * 3600 * 1000)
    else if (type === 'yearly') expiresAt = new Date(now.getTime() + 365 * 24 * 3600 * 1000)
    const license = await prisma.license.create({ data: { userId, licenseKey, machineId: machine_id, type, expiresAt, orderId: order_id || null } })
    return NextResponse.json({ license_key: licenseKey, id: license.id, expires_at: expiresAt })
  } catch (e: unknown) { return NextResponse.json({ error: (e as Error).message }, { status: 500 }) }
}
FILEOF
echo "  api/license/generate OK"

cat > "$WEB_DIR/src/app/api/license/list/route.ts" << 'FILEOF'
import { NextRequest, NextResponse } from 'next/server'
import { prisma } from '@/lib/db'
import { verifyToken } from '@/lib/auth'
export async function GET(req: NextRequest) {
  try {
    const auth = req.headers.get('authorization')
    if (!auth?.startsWith('Bearer ')) return NextResponse.json({ error: '未登录' }, { status: 401 })
    const { userId } = await verifyToken(auth.slice(7))
    const licenses = await prisma.license.findMany({ where: { userId }, orderBy: { createdAt: 'desc' } })
    return NextResponse.json({ licenses })
  } catch (e: unknown) { return NextResponse.json({ error: (e as Error).message }, { status: 500 }) }
}
FILEOF
echo "  api/license/list OK"

cat > "$WEB_DIR/src/app/api/payment/create/route.ts" << 'FILEOF'
import { NextRequest, NextResponse } from 'next/server'
import { prisma } from '@/lib/db'
import { verifyToken } from '@/lib/auth'
import { createNativeOrder } from '@/lib/wechat-pay'
const PRICES: Record<string, number> = { monthly: 990, yearly: 9900, permanent: 69900 }
export async function POST(req: NextRequest) {
  try {
    const auth = req.headers.get('authorization')
    if (!auth?.startsWith('Bearer ')) return NextResponse.json({ error: '未登录' }, { status: 401 })
    const { userId } = await verifyToken(auth.slice(7))
    const { type, machine_id } = await req.json()
    if (!type || !PRICES[type]) return NextResponse.json({ error: 'type无效' }, { status: 400 })
    if (!machine_id) return NextResponse.json({ error: 'machine_id必填' }, { status: 400 })
    const amount = PRICES[type]
    const order = await prisma.order.create({ data: { userId, type, amount, status: 'pending' } })
    const host = req.headers.get('host') || 'localhost:3001'
    const protocol = req.headers.get('x-forwarded-proto') || 'https'
    const result = await createNativeOrder({ orderId: String(order.id), description: `LCA ${type}授权码`, amount, notifyUrl: `${protocol}://${host}/api/payment/notify` })
    return NextResponse.json({ order_id: order.id, code_url: result.code_url, amount })
  } catch (e: unknown) { return NextResponse.json({ error: (e as Error).message }, { status: 500 }) }
}
FILEOF
echo "  api/payment/create OK"

cat > "$WEB_DIR/src/app/api/payment/notify/route.ts" << 'FILEOF'
import { NextRequest, NextResponse } from 'next/server'
import { prisma } from '@/lib/db'
import { generateLicenseKey } from '@/lib/license'
export async function POST(req: NextRequest) {
  try {
    const body = await req.text()
    const data = JSON.parse(body)
    const orderId = parseInt(data.resource?.ciphertext?.out_trade_no || '0')
    if (!orderId) return NextResponse.json({ code: 'FAIL', message: '无效订单' }, { status: 400 })
    const order = await prisma.order.update({ where: { id: orderId }, data: { status: 'paid', paidAt: new Date() } })
    const machineId = `pending_${order.userId}_${order.id}`
    const licenseKey = await generateLicenseKey(machineId, order.type as 'monthly' | 'yearly' | 'permanent')
    const now = new Date()
    let expiresAt: Date | null = null
    if (order.type === 'monthly') expiresAt = new Date(now.getTime() + 30 * 24 * 3600 * 1000)
    else if (order.type === 'yearly') expiresAt = new Date(now.getTime() + 365 * 24 * 3600 * 1000)
    await prisma.license.create({ data: { userId: order.userId, licenseKey, machineId, type: order.type, expiresAt, orderId: order.id } })
    return NextResponse.json({ code: 'SUCCESS', message: 'OK' })
  } catch { return NextResponse.json({ code: 'FAIL', message: '处理失败' }, { status: 500 }) }
}
FILEOF
echo "  api/payment/notify OK"

# ---- Auth Pages (simplified) ----
cat > "$WEB_DIR/src/app/register/page.tsx" << 'FILEOF'
'use client'
import { useState } from 'react'
import Link from 'next/link'
import { useRouter } from 'next/navigation'
export default function Register() {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const router = useRouter()
  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault(); setError('')
    if (password.length < 6) { setError('密码至少6位'); return }
    setLoading(true)
    try {
      const res = await fetch('/api/auth/register', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ email, password }) })
      const data = await res.json()
      if (!res.ok) { setError(data.error); return }
      router.push('/login')
    } catch { setError('注册失败') } finally { setLoading(false) }
  }
  return (
    <div className="min-h-screen flex items-center justify-center px-4">
      <div className="w-full max-w-md">
        <div className="text-center mb-8">
          <div className="w-14 h-14 rounded-xl bg-gradient-to-br from-[#00d4ff] to-[#0099cc] flex items-center justify-center text-[#0a1628] font-extrabold text-2xl mx-auto mb-4">L</div>
          <h1 className="text-2xl font-bold">注册 LCA 账号</h1>
        </div>
        <form onSubmit={handleSubmit} className="space-y-4 bg-[#111d31] p-8 rounded-xl border border-[#1e2d45]">
          {error && <div className="text-red-400 text-sm bg-red-400/10 p-3 rounded-lg">{error}</div>}
          <div><label className="block text-sm mb-1 text-[#8899aa]">邮箱</label><input type="email" value={email} onChange={e => setEmail(e.target.value)} required className="w-full px-4 py-3 bg-[#0a1628] border border-[#1e2d45] rounded-lg text-[#e8edf5] focus:border-[#00d4ff] outline-none" /></div>
          <div><label className="block text-sm mb-1 text-[#8899aa]">密码</label><input type="password" value={password} onChange={e => setPassword(e.target.value)} required className="w-full px-4 py-3 bg-[#0a1628] border border-[#1e2d45] rounded-lg text-[#e8edf5] focus:border-[#00d4ff] outline-none" placeholder="至少6位" /></div>
          <button type="submit" disabled={loading} className="w-full py-3 rounded-lg bg-gradient-to-r from-[#00d4ff] to-[#0099cc] text-[#0a1628] font-bold disabled:opacity-50">{loading ? '注册中...' : '注 册'}</button>
        </form>
        <p className="text-center text-sm text-[#8899aa] mt-4">已有账号？<Link href="/login" className="text-[#00d4ff]">去登录</Link></p>
      </div>
    </div>
  )
}
FILEOF
echo "  register/page.tsx OK"

cat > "$WEB_DIR/src/app/login/page.tsx" << 'FILEOF'
'use client'
import { useState } from 'react'
import Link from 'next/link'
import { useRouter } from 'next/navigation'
export default function Login() {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const router = useRouter()
  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault(); setError(''); setLoading(true)
    try {
      const res = await fetch('/api/auth/login', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ email, password }) })
      const data = await res.json()
      if (!res.ok) { setError(data.error); return }
      localStorage.setItem('lca_web_token', data.token)
      localStorage.setItem('lca_web_user', JSON.stringify(data.user))
      router.push('/dashboard')
    } catch { setError('登录失败') } finally { setLoading(false) }
  }
  return (
    <div className="min-h-screen flex items-center justify-center px-4">
      <div className="w-full max-w-md">
        <div className="text-center mb-8">
          <div className="w-14 h-14 rounded-xl bg-gradient-to-br from-[#00d4ff] to-[#0099cc] flex items-center justify-center text-[#0a1628] font-extrabold text-2xl mx-auto mb-4">L</div>
          <h1 className="text-2xl font-bold">登录 LCA</h1>
        </div>
        <form onSubmit={handleSubmit} className="space-y-4 bg-[#111d31] p-8 rounded-xl border border-[#1e2d45]">
          {error && <div className="text-red-400 text-sm bg-red-400/10 p-3 rounded-lg">{error}</div>}
          <div><label className="block text-sm mb-1 text-[#8899aa]">邮箱</label><input type="email" value={email} onChange={e => setEmail(e.target.value)} required className="w-full px-4 py-3 bg-[#0a1628] border border-[#1e2d45] rounded-lg text-[#e8edf5] focus:border-[#00d4ff] outline-none" /></div>
          <div><label className="block text-sm mb-1 text-[#8899aa]">密码</label><input type="password" value={password} onChange={e => setPassword(e.target.value)} required className="w-full px-4 py-3 bg-[#0a1628] border border-[#1e2d45] rounded-lg text-[#e8edf5] focus:border-[#00d4ff] outline-none" /></div>
          <button type="submit" disabled={loading} className="w-full py-3 rounded-lg bg-gradient-to-r from-[#00d4ff] to-[#0099cc] text-[#0a1628] font-bold disabled:opacity-50">{loading ? '登录中...' : '登 录'}</button>
        </form>
        <p className="text-center text-sm text-[#8899aa] mt-4">没有账号？<Link href="/register" className="text-[#00d4ff]">去注册</Link></p>
      </div>
    </div>
  )
}
FILEOF
echo "  login/page.tsx OK"

# ---- Dashboard ----
cat > "$WEB_DIR/src/app/dashboard/page.tsx" << 'FILEOF'
'use client'
import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
interface License { id: number; licenseKey: string; machineId: string; type: string; status: string; expiresAt: string | null; createdAt: string }
export default function Dashboard() {
  const [licenses, setLicenses] = useState<License[]>([])
  const [machineId, setMachineId] = useState('')
  const [selectedType, setSelectedType] = useState<'monthly' | 'yearly' | 'permanent'>('yearly')
  const [loading, setLoading] = useState(false)
  const [copiedId, setCopiedId] = useState<number | null>(null)
  const [user, setUser] = useState<{ id: number; email: string } | null>(null)
  const router = useRouter()
  useEffect(() => {
    const token = localStorage.getItem('lca_web_token')
    const userData = localStorage.getItem('lca_web_user')
    if (!token || !userData) { router.push('/login'); return }
    setUser(JSON.parse(userData))
    fetchLicenses(token)
  }, [])
  async function fetchLicenses(token: string) {
    try {
      const res = await fetch('/api/license/list', { headers: { Authorization: `Bearer ${token}` } })
      if (res.ok) { const data = await res.json(); setLicenses(data.licenses || []) }
    } catch { /* ignore */ }
  }
  async function handleGenerate() {
    if (!machineId.trim()) return
    const token = localStorage.getItem('lca_web_token')
    if (!token) return
    setLoading(true)
    try {
      const res = await fetch('/api/license/generate', { method: 'POST', headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` }, body: JSON.stringify({ machine_id: machineId, type: selectedType }) })
      if (res.ok) { setMachineId(''); fetchLicenses(token) }
    } catch { /* ignore */ } finally { setLoading(false) }
  }
  function copyKey(id: number, key: string) { navigator.clipboard.writeText(key); setCopiedId(id); setTimeout(() => setCopiedId(null), 2000) }
  function handleLogout() { localStorage.removeItem('lca_web_token'); localStorage.removeItem('lca_web_user'); router.push('/login') }
  const typeLabel: Record<string, string> = { monthly: '月付', yearly: '年付', permanent: '永久' }
  return (
    <div className="min-h-screen">
      <nav className="border-b border-[#1e2d45] bg-[#0a1628]/90 backdrop-blur-md">
        <div className="max-w-6xl mx-auto px-6 h-16 flex items-center justify-between">
          <div className="flex items-center gap-3"><div className="w-8 h-8 rounded-lg bg-gradient-to-br from-[#00d4ff] to-[#0099cc] flex items-center justify-center text-[#0a1628] font-extrabold">L</div><span className="font-bold text-lg">LCA 控制台</span></div>
          <div className="flex items-center gap-4"><span className="text-sm text-[#8899aa]">{user?.email}</span><button onClick={handleLogout} className="text-sm text-[#8899aa] hover:text-[#00d4ff]">退出</button></div>
        </div>
      </nav>
      <div className="max-w-6xl mx-auto px-6 py-8">
        <h2 className="text-2xl font-bold mb-6">我的授权码</h2>
        <div className="bg-[#111d31] p-6 rounded-xl border border-[#1e2d45] mb-8">
          <h3 className="font-semibold mb-4">生成新授权码</h3>
          <div className="flex flex-col md:flex-row gap-4">
            <input value={machineId} onChange={e => setMachineId(e.target.value)} placeholder="输入机器ID（在LCA管理后台获取）" className="flex-1 px-4 py-3 bg-[#0a1628] border border-[#1e2d45] rounded-lg text-[#e8edf5] focus:border-[#00d4ff] outline-none text-sm" />
            <select value={selectedType} onChange={e => setSelectedType(e.target.value as 'monthly' | 'yearly' | 'permanent')} className="px-4 py-3 bg-[#0a1628] border border-[#1e2d45] rounded-lg text-[#e8edf5] outline-none">
              <option value="monthly">月付 ¥9.9</option><option value="yearly">年付 ¥99</option><option value="permanent">永久 ¥699</option>
            </select>
            <button onClick={handleGenerate} disabled={loading || !machineId.trim()} className="px-6 py-3 rounded-lg bg-gradient-to-r from-[#00d4ff] to-[#0099cc] text-[#0a1628] font-bold disabled:opacity-50 whitespace-nowrap">{loading ? '生成中...' : '生成授权码'}</button>
          </div>
        </div>
        <div className="space-y-4">
          {licenses.length === 0 && <p className="text-[#8899aa] text-center py-12">暂无授权码，请先生成</p>}
          {licenses.map(lic => (
            <div key={lic.id} className="bg-[#111d31] p-6 rounded-xl border border-[#1e2d45]">
              <div className="flex items-center justify-between mb-3">
                <div className="flex items-center gap-3">
                  <span className="px-2 py-1 text-xs font-bold rounded bg-[#00d4ff]/10 text-[#00d4ff]">{typeLabel[lic.type] || lic.type}</span>
                  <span className={`px-2 py-1 text-xs font-bold rounded ${lic.status === 'active' ? 'bg-green-400/10 text-green-400' : 'bg-red-400/10 text-red-400'}`}>{lic.status === 'active' ? '有效' : lic.status}</span>
                </div>
                <button onClick={() => copyKey(lic.id, lic.licenseKey)} className="text-xs text-[#00d4ff] hover:underline">{copiedId === lic.id ? '已复制!' : '复制授权码'}</button>
              </div>
              <div className="text-xs text-[#8899aa] space-y-1">
                <p>机器ID: {lic.machineId}</p>
                <p>创建时间: {new Date(lic.createdAt).toLocaleString('zh-CN')}</p>
                {lic.expiresAt && <p>到期时间: {new Date(lic.expiresAt).toLocaleString('zh-CN')}</p>}
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
FILEOF
echo "  dashboard/page.tsx OK"

# ---- .env.local ----
cat > "$WEB_DIR/.env.local" << 'FILEOF'
JWT_SECRET=lca-web-secret-change-in-production
# WECHAT_MCH_ID=
# WECHAT_APP_ID=
# WECHAT_API_KEY=
# WECHAT_SERIAL_NO=
FILEOF
echo "  .env.local OK"

# ---- Prisma setup ----
cd "$WEB_DIR"
npx prisma generate 2>/dev/null
npx prisma db push 2>/dev/null
echo ""
echo "=== 完成！==="
echo "运行官网: cd $WEB_DIR && npm run dev"
echo "访问: http://localhost:3001"
