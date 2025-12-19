import json
import os
import unittest
import uuid
import urllib.error
import urllib.request


API_BASE_URL = os.getenv("API_GATEWAY_URL", "http://localhost:8080").rstrip("/")
API_V1_URL = f"{API_BASE_URL}/api/v1"


def _parse_json(raw_body, content_type):
    if not raw_body:
        return None
    if content_type and "application/json" in content_type:
        return json.loads(raw_body.decode("utf-8"))
    return None


def _request_json(method, url, payload=None):
    headers = {}
    data = None
    if payload is not None:
        data = json.dumps(payload).encode("utf-8")
        headers["Content-Type"] = "application/json"

    req = urllib.request.Request(url, data=data, headers=headers, method=method)
    try:
        with urllib.request.urlopen(req, timeout=10) as resp:
            raw_body = resp.read()
            return resp.status, _parse_json(raw_body, resp.headers.get("Content-Type"))
    except urllib.error.HTTPError as err:
        raw_body = err.read()
        return err.code, _parse_json(raw_body, err.headers.get("Content-Type"))


class ApiGatewayTests(unittest.TestCase):
    def test_health(self):
        req = urllib.request.Request(f"{API_BASE_URL}/health", method="GET")
        try:
            with urllib.request.urlopen(req, timeout=10) as resp:
                self.assertEqual(resp.status, 200)
        except urllib.error.URLError as err:
            self.fail(f"API gateway is not reachable at {API_BASE_URL}: {err}")

    def test_create_work_success_and_duplicate(self):
        work_id = f"test-{uuid.uuid4().hex}"
        payload = {
            "workId": work_id,
            "name": "Test Work",
            "description": "Integration test work",
        }

        status, body = _request_json("POST", f"{API_V1_URL}/works", payload)
        self.assertEqual(status, 201)
        self.assertIsNotNone(body)
        self.assertEqual(body.get("workId"), work_id)
        self.assertEqual(body.get("name"), payload["name"])
        self.assertEqual(body.get("description"), payload["description"])
        self.assertTrue(body.get("createdAt"))

        status, body = _request_json("POST", f"{API_V1_URL}/works", payload)
        self.assertEqual(status, 409)
        self.assertIsNotNone(body)
        self.assertEqual(body.get("code"), "WORK_ALREADY_EXISTS")

    def test_create_work_validation_error(self):
        payload = {
            "workId": "",
            "name": "",
            "description": "",
        }
        status, body = _request_json("POST", f"{API_V1_URL}/works", payload)
        self.assertEqual(status, 400)
        self.assertIsNotNone(body)
        self.assertEqual(body.get("code"), "VALIDATION_ERROR")


if __name__ == "__main__":
    unittest.main()
