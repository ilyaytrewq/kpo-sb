export function makeUUID() {
  if (globalThis.crypto?.randomUUID) return globalThis.crypto.randomUUID();
  // fallback
  return "xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx".replace(/[xy]/g, (c) => {
    const r = (Math.random() * 16) | 0;
    const v = c === "x" ? r : (r & 0x3) | 0x8;
    return v.toString(16);
  });
}

export async function httpJson(method, path, { userId, idemKey, body } = {}) {
  const headers = { "Content-Type": "application/json" };
  if (userId) headers["X-User-Id"] = userId;
  if (idemKey) headers["Idempotency-Key"] = idemKey;

  const res = await fetch(path, {
    method,
    headers,
    body: body === undefined ? undefined : JSON.stringify(body),
  });

  const text = await res.text();
  let data = null;
  try {
    data = text ? JSON.parse(text) : null;
  } catch {
    data = text;
  }

  return { ok: res.ok, status: res.status, data };
}

export async function mustJson(method, path, opts) {
  const r = await httpJson(method, path, opts);
  if (!r.ok) {
    const msg = typeof r.data === "string" ? r.data : JSON.stringify(r.data);
    throw new Error(`${method} ${path} -> ${r.status}: ${msg}`);
  }
  return r.data;
}

export function sleep(ms) {
  return new Promise((r) => setTimeout(r, ms));
}
