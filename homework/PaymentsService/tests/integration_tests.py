#!/usr/bin/env python3
import json
import os
import sys
import time
import unittest
import uuid
import urllib.error
import urllib.parse
import urllib.request

DEFAULT_BASE_URL = "http://158.160.219.201:8080/api/v1"

#
def parse_base_url() -> str:
    base = os.getenv("BASE_URL", DEFAULT_BASE_URL)
    args = [sys.argv[0]]
    i = 1
    while i < len(sys.argv):
        if sys.argv[i] == "--base-url" and i + 1 < len(sys.argv):
            base = sys.argv[i + 1]
            i += 2
            continue
        args.append(sys.argv[i])
        i += 1
    sys.argv[:] = args
    return base


BASE_URL = DEFAULT_BASE_URL


def build_url(path: str) -> str:
    return BASE_URL.rstrip("/") + "/" + path.lstrip("/")


def request(method: str, path: str, headers=None, body=None, timeout=10):
    url = build_url(path)
    headers = headers or {}
    data = None
    if body is not None:
        if isinstance(body, (dict, list)):
            data = json.dumps(body).encode("utf-8")
            headers.setdefault("Content-Type", "application/json")
        elif isinstance(body, str):
            data = body.encode("utf-8")
        else:
            data = body
    req = urllib.request.Request(url, data=data, method=method)
    for k, v in headers.items():
        req.add_header(k, v)
    try:
        with urllib.request.urlopen(req, timeout=timeout) as resp:
            return resp.status, resp.read(), resp.headers.get("Content-Type", "")
    except urllib.error.HTTPError as err:
        return err.code, err.read(), err.headers.get("Content-Type", "")


def parse_json(body: bytes):
    if not body:
        return None
    return json.loads(body.decode("utf-8"))


class IntegrationTests(unittest.TestCase):
    def test_missing_idempotency_key_returns_400(self):
        status, body, _ = request(
            "POST",
            "/payments/account",
            headers={"Content-Type": "application/json"},
            body={},
        )
        self.assertEqual(status, 400, msg=f"body={body!r}")

    def test_account_topup_balance(self):
        user_id = f"it-{uuid.uuid4()}"
        idem_create = str(uuid.uuid4())
        status, body, _ = request(
            "POST",
            "/payments/account",
            headers={"X-User-Id": user_id, "Idempotency-Key": idem_create},
            body=None,
        )
        self.assertEqual(status, 201, msg=f"body={body!r}")
        resp = parse_json(body)
        self.assertEqual(resp.get("user_id"), user_id)
        self.assertEqual(resp.get("balance"), 0)

        amount = 500
        idem_topup = str(uuid.uuid4())
        status, body, _ = request(
            "POST",
            "/payments/account/topup",
            headers={"X-User-Id": user_id, "Idempotency-Key": idem_topup},
            body={"amount": amount},
        )
        self.assertEqual(status, 200, msg=f"body={body!r}")
        resp = parse_json(body)
        self.assertEqual(resp.get("user_id"), user_id)
        self.assertEqual(resp.get("balance"), amount)

        status, body, _ = request(
            "GET",
            "/payments/account/balance",
            headers={"X-User-Id": user_id},
        )
        self.assertEqual(status, 200, msg=f"body={body!r}")
        resp = parse_json(body)
        self.assertEqual(resp.get("user_id"), user_id)
        self.assertEqual(resp.get("balance"), amount)

        status, body, _ = request("GET", "/payments/account/balance")
        self.assertEqual(status, 400, msg=f"body={body!r}")

    def test_orders_flow(self):
        user_id = f"it-{uuid.uuid4()}"
        idem_create = str(uuid.uuid4())
        status, body, _ = request(
            "POST",
            "/payments/account",
            headers={"X-User-Id": user_id, "Idempotency-Key": idem_create},
            body=None,
        )
        self.assertEqual(status, 201, msg=f"body={body!r}")

        amount = 100
        idem_topup = str(uuid.uuid4())
        status, body, _ = request(
            "POST",
            "/payments/account/topup",
            headers={"X-User-Id": user_id, "Idempotency-Key": idem_topup},
            body={"amount": 1000},
        )
        self.assertEqual(status, 200, msg=f"body={body!r}")

        idem_order = str(uuid.uuid4())
        status, body, _ = request(
            "POST",
            "/orders",
            headers={"X-User-Id": user_id, "Idempotency-Key": idem_order},
            body={"amount": amount, "description": "integration-test"},
        )
        self.assertEqual(status, 201, msg=f"body={body!r}")
        resp = parse_json(body)
        order = resp.get("order", {})
        order_id = order.get("order_id")
        self.assertTrue(order_id)
        self.assertEqual(order.get("user_id"), user_id)
        self.assertEqual(order.get("amount"), amount)

        status, body, _ = request(
            "GET",
            f"/orders/{order_id}",
            headers={"X-User-Id": user_id},
        )
        self.assertEqual(status, 200, msg=f"body={body!r}")
        resp = parse_json(body)
        got = resp.get("order", {})
        self.assertEqual(got.get("order_id"), order_id)
        self.assertEqual(got.get("user_id"), user_id)

        found = False
        for _ in range(5):
            qs = urllib.parse.urlencode({"limit": 50})
            status, body, _ = request(
                "GET",
                f"/orders?{qs}",
                headers={"X-User-Id": user_id},
            )
            self.assertEqual(status, 200, msg=f"body={body!r}")
            resp = parse_json(body)
            orders = resp.get("orders", [])
            if any(o.get("order_id") == order_id for o in orders):
                found = True
                break
            time.sleep(0.5)
        self.assertTrue(found, "order not found in list orders")


if __name__ == "__main__":
    BASE_URL = parse_base_url()
    unittest.main()
