const BASE = '/api'

export interface HijackRule {
  id: string
  source: string
  source_port: number
  target: string
  target_tls: boolean
  target_host: string
  description: string
  enabled: boolean
  created_at: string
  updated_at: string
}

export async function getRules(): Promise<HijackRule[]> {
  const r = await fetch(`${BASE}/rules`)
  const data = await r.json()
  return data.rules
}

export async function createRule(rule: Partial<HijackRule>): Promise<HijackRule> {
  const r = await fetch(`${BASE}/rules`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(rule),
  })
  if (!r.ok) throw new Error((await r.json()).error)
  return r.json()
}

export async function updateRule(id: string, rule: Partial<HijackRule>): Promise<HijackRule> {
  const r = await fetch(`${BASE}/rules/${id}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(rule),
  })
  if (!r.ok) throw new Error((await r.json()).error)
  return r.json()
}

export async function deleteRule(id: string): Promise<void> {
  const r = await fetch(`${BASE}/rules/${id}`, { method: 'DELETE' })
  if (!r.ok) throw new Error((await r.json()).error)
}

export async function enableRule(id: string): Promise<HijackRule> {
  const r = await fetch(`${BASE}/rules/${id}/enable`, { method: 'POST' })
  if (!r.ok) throw new Error((await r.json()).error)
  return r.json()
}

export async function disableRule(id: string): Promise<HijackRule> {
  const r = await fetch(`${BASE}/rules/${id}/disable`, { method: 'POST' })
  if (!r.ok) throw new Error((await r.json()).error)
  return r.json()
}

export async function getStatus(): Promise<any> {
  const r = await fetch(`${BASE}/status`)
  return r.json()
}

export async function cleanupHosts(): Promise<void> {
  await fetch(`${BASE}/cleanup`, { method: 'POST' })
}
