import React, { useMemo, useState } from "react";
import toast, { Toaster } from "react-hot-toast";
import { makeUUID, mustJson, httpJson, sleep } from "./lib/api.js";

function pickStoredUser() {
  try {
    return localStorage.getItem("gozon_user_id") || "";
  } catch {
    return "";
  }
}

function storeUser(v) {
  try {
    localStorage.setItem("gozon_user_id", v);
  } catch {}
}

function pretty(x) {
  try {
    return JSON.stringify(x, null, 2);
  } catch {
    return String(x);
  }
}

export default function App() {
  const [userId, setUserId] = useState(pickStoredUser());
  const [balance, setBalance] = useState(null);

  const [topupAmount, setTopupAmount] = useState(10000);
  const [orderAmount, setOrderAmount] = useState(3000);
  const [orderDesc, setOrderDesc] = useState("order");
  const [limit, setLimit] = useState(50);

  const [orders, setOrders] = useState([]);
  const [orderId, setOrderId] = useState("");
  const [orderDetails, setOrderDetails] = useState(null);

  const [idemKind, setIdemKind] = useState("topup");
  const [idemN, setIdemN] = useState(5);
  const [idemLog, setIdemLog] = useState(null);

  const [busy, setBusy] = useState(false);

  const resolvedUserId = useMemo(() => userId.trim(), [userId]);

  function ensureUserId() {
    const v = resolvedUserId;
    if (!v) throw new Error("user_id пустой");
    return v;
  }

  async function refreshBalance() {
    const uid = ensureUserId();
    const data = await mustJson("GET", "/api/v1/payments/account/balance", { userId: uid });
    setBalance(data.balance);
    return data.balance;
  }

  async function createAccount() {
    const uid = ensureUserId();
    const idemKey = makeUUID();
    const data = await mustJson("POST", "/api/v1/payments/account", { userId: uid, idemKey, body: {} });
    toast.success("Счёт создан");
    setBalance(data.balance ?? 0);
  }

  async function topup() {
    const uid = ensureUserId();
    const amount = Number(topupAmount);
    if (!Number.isFinite(amount) || amount <= 0) throw new Error("amount должен быть > 0");

    const idemKey = makeUUID();
    const data = await mustJson("POST", "/api/v1/payments/account/topup", { userId: uid, idemKey, body: { amount } });
    toast.success("Баланс пополнен");
    setBalance(data.balance);
  }

  async function createOrder() {
    const uid = ensureUserId();
    const amount = Number(orderAmount);
    if (!Number.isFinite(amount) || amount <= 0) throw new Error("amount должен быть > 0");

    const idemKey = makeUUID();
    const data = await mustJson("POST", "/api/v1/orders", { userId: uid, idemKey, body: { amount, description: orderDesc || "order" } });
    const id = data?.order?.order_id || "";
    toast.success("Заказ создан");
    setOrderId(id);
    await listOrders();
  }

  async function listOrders() {
    const uid = ensureUserId();
    const l = Math.min(200, Math.max(1, Number(limit) || 50));
    const data = await mustJson("GET", `/api/v1/orders?limit=${encodeURIComponent(l)}`, { userId: uid });
    setOrders(data.orders || []);
    return data.orders || [];
  }

  async function getOrder() {
    const uid = ensureUserId();
    const id = orderId.trim();
    if (!id) throw new Error("order_id пустой");
    const data = await mustJson("GET", `/api/v1/orders/${encodeURIComponent(id)}`, { userId: uid });
    setOrderDetails(data);
    return data;
  }

  async function pollFinal() {
    const uid = ensureUserId();
    const id = orderId.trim();
    if (!id) throw new Error("order_id пустой");

    const deadline = Date.now() + 20000;
    while (Date.now() < deadline) {
      const data = await mustJson("GET", `/api/v1/orders/${encodeURIComponent(id)}`, { userId: uid });
      setOrderDetails(data);
      const st = data?.order?.status;
      if (st === "FINISHED" || st === "CANCELLED") {
        toast.success(`Статус: ${st}`);
        return data;
      }
      await sleep(400);
    }
    throw new Error("Не дождались FINISHED/CANCELLED за 20 секунд");
  }

  async function runIdempotency() {
    const uid = ensureUserId();
    const n = Math.min(30, Math.max(2, Number(idemN) || 5));
    const idemKey = makeUUID();

    let calls = [];
    if (idemKind === "topup") {
      const amount = Number(topupAmount) || 1000;
      calls = Array.from({ length: n }, () =>
        httpJson("POST", "/api/v1/payments/account/topup", { userId: uid, idemKey, body: { amount } })
      );
    } else {
      const amount = Number(orderAmount) || 1000;
      const description = (orderDesc || "idem-order") + " (idem)";
      calls = Array.from({ length: n }, () =>
        httpJson("POST", "/api/v1/orders", { userId: uid, idemKey, body: { amount, description } })
      );
    }

    const results = await Promise.all(calls);
    setIdemLog({ idemKey, kind: idemKind, repeats: n, results });

    // convenience refresh
    await refreshBalance();
    await listOrders();

    toast.success("Идемпотентность проверена");
  }

  async function wrap(fn) {
    try {
      setBusy(true);
      await fn();
    } catch (e) {
      toast.error(e?.message || String(e));
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="wrap">
      <Toaster position="top-right" />
      <header>
        <div>
          <h1>GoZon — Frontend</h1>
          <div className="muted">
            POST требуют заголовок <span className="kbd">Idempotency-Key</span>; UI ставит его сам.
          </div>
        </div>
        <button className="secondary" disabled={busy} onClick={() => {
          const v = "user-" + makeUUID().slice(0, 8);
          setUserId(v);
          storeUser(v);
          toast.success("user_id сгенерирован");
        }}>
          Сгенерировать user_id
        </button>
      </header>

      <section className="card">
        <h2>Пользователь и баланс</h2>
        <div className="grid2">
          <div className="field">
            <label>user_id</label>
            <input
              value={userId}
              onChange={(e) => { setUserId(e.target.value); storeUser(e.target.value); }}
              placeholder="например, test-user-1"
            />
          </div>

          <div className="field">
            <label>Баланс</label>
            <div className="row">
              <div className="pill">{balance === null ? "—" : String(balance)}</div>
              <button className="secondary" disabled={busy} onClick={() => wrap(refreshBalance)}>Обновить</button>
            </div>
          </div>
        </div>

        <div className="row" style={{ marginTop: 10 }}>
          <button disabled={busy} onClick={() => wrap(createAccount)}>Создать счёт</button>

          <div style={{ flex: 1 }} />

          <input
            style={{ maxWidth: 220 }}
            type="number"
            min="1"
            value={topupAmount}
            onChange={(e) => setTopupAmount(e.target.value)}
            placeholder="Сумма пополнения"
          />
          <button disabled={busy} onClick={() => wrap(topup)}>Пополнить</button>
        </div>
      </section>

      <section className="card">
        <h2>Заказы</h2>

        <div className="grid3">
          <div className="field">
            <label>amount</label>
            <input type="number" min="1" value={orderAmount} onChange={(e) => setOrderAmount(e.target.value)} />
          </div>
          <div className="field">
            <label>description</label>
            <input value={orderDesc} onChange={(e) => setOrderDesc(e.target.value)} placeholder="Описание" />
          </div>
          <div className="field">
            <label>&nbsp;</label>
            <button disabled={busy} onClick={() => wrap(createOrder)}>Создать заказ</button>
          </div>
        </div>

        <div className="row" style={{ marginTop: 10 }}>
          <button className="secondary" disabled={busy} onClick={() => wrap(listOrders)}>Обновить список</button>
          <span className="muted">limit</span>
          <input style={{ maxWidth: 120 }} type="number" min="1" max="200" value={limit} onChange={(e) => setLimit(e.target.value)} />
        </div>

        <div className="split">
          <div>
            <h3>Список</h3>
            <div className="list">
              {orders.map((o) => (
                <div key={o.order_id} className="item" onClick={() => { setOrderId(o.order_id); wrap(getOrder); }}>
                  <b>{o.order_id}</b><br />
                  status: <b>{o.status}</b> · amount: {o.amount}<br />
                  <span className="muted">{o.description}</span>
                </div>
              ))}
              {orders.length === 0 && <div className="muted">Пока пусто (или не нажал “Обновить список”).</div>}
            </div>
          </div>

          <div>
            <h3>Детали</h3>
            <div className="row" style={{ marginBottom: 10 }}>
              <input value={orderId} onChange={(e) => setOrderId(e.target.value)} placeholder="order_id" />
              <button className="secondary" disabled={busy} onClick={() => wrap(getOrder)}>Получить</button>
              <button className="secondary" disabled={busy} onClick={() => wrap(pollFinal)}>Ожидать финал</button>
            </div>

            <pre className="pre">{orderDetails === null ? "—" : pretty(orderDetails)}</pre>
          </div>
        </div>
      </section>

      <section className="card">
        <h2>Проверка идемпотентности</h2>
        <div className="muted" style={{ marginBottom: 10 }}>
          Шлём несколько одинаковых POST с одним ключом — ожидаем “эффект один”.
        </div>

        <div className="row">
          <select value={idemKind} onChange={(e) => setIdemKind(e.target.value)}>
            <option value="topup">Topup</option>
            <option value="createOrder">Create order</option>
          </select>
          <input style={{ maxWidth: 120 }} type="number" min="2" max="30" value={idemN} onChange={(e) => setIdemN(e.target.value)} />
          <span className="muted">повторов</span>
          <button className="secondary" disabled={busy} onClick={() => wrap(runIdempotency)}>Запустить</button>
        </div>

        <pre className="pre" style={{ marginTop: 10 }}>
          {idemLog === null ? "—" : pretty(idemLog)}
        </pre>
      </section>

      <section className="card">
        <h2>Порты</h2>
        <div className="muted">
          Frontend: <span className="kbd">http://localhost:3000</span> ·
          Gateway: <span className="kbd">http://localhost:8080</span> ·
          Swagger: <span className="kbd">http://localhost:8088</span> ·
          Kafka UI: <span className="kbd">http://localhost:8085</span>
        </div>
      </section>
    </div>
  );
}
