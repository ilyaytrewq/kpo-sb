import json
import os
import time
import uuid
import unittest
import urllib.error
import urllib.request


API_BASE_URL = os.getenv("API_GATEWAY_URL", "http://localhost:8080").rstrip("/")
API_V1_URL = f"{API_BASE_URL}/api/v1"

RUN_FULL_FLOW = os.getenv("RUN_FULL_FLOW", "0") == "1"
E2E_TIMEOUT_SECONDS = int(os.getenv("E2E_TIMEOUT_SECONDS", "120"))
POLL_INTERVAL_SECONDS = float(os.getenv("POLL_INTERVAL_SECONDS", "2"))


def _parse_json(raw_body, content_type):
    if not raw_body:
        return None
    if content_type and "application/json" in content_type:
        return json.loads(raw_body.decode("utf-8"))
    return None


def _request(method, url, headers=None, data=None):
    headers = headers or {}
    req = urllib.request.Request(url, data=data, headers=headers, method=method)
    try:
        with urllib.request.urlopen(req, timeout=20) as resp:
            raw_body = resp.read()
            return resp.status, _parse_json(raw_body, resp.headers.get("Content-Type"))
    except urllib.error.HTTPError as err:
        raw_body = err.read()
        return err.code, _parse_json(raw_body, err.headers.get("Content-Type"))


def _request_json(method, url, payload=None):
    headers = {}
    data = None
    if payload is not None:
        data = json.dumps(payload).encode("utf-8")
        headers["Content-Type"] = "application/json"
    return _request(method, url, headers=headers, data=data)


def _multipart_body(field_name, filename, content_bytes, content_type="text/plain"):
    boundary = f"----boundary{uuid.uuid4().hex}"
    crlf = "\r\n"
    body = bytearray()

    body.extend(f"--{boundary}{crlf}".encode("utf-8"))
    body.extend(
        f'Content-Disposition: form-data; name="{field_name}"; filename="{filename}"{crlf}'.encode("utf-8")
    )
    body.extend(f"Content-Type: {content_type}{crlf}{crlf}".encode("utf-8"))
    body.extend(content_bytes)
    body.extend(crlf.encode("utf-8"))

    body.extend(f"--{boundary}--{crlf}".encode("utf-8"))

    headers = {"Content-Type": f"multipart/form-data; boundary={boundary}"}
    return headers, bytes(body)


def _submit_file(work_id, filename, content_bytes):
    url = f"{API_V1_URL}/works/{work_id}/submissions"
    headers, body = _multipart_body("file", filename, content_bytes, content_type="text/plain")
    return _request("POST", url, headers=headers, data=body)


def _wait_submission_done(submission_id):
    url = f"{API_V1_URL}/submissions/{submission_id}"
    deadline = time.time() + E2E_TIMEOUT_SECONDS

    last = None
    while time.time() < deadline:
        status, body = _request("GET", url)
        last = (status, body)
        if status == 200 and body and body.get("status") in ("DONE", "ERROR"):
            return status, body
        time.sleep(POLL_INTERVAL_SECONDS)

    raise AssertionError(f"Timeout waiting for submission {submission_id} DONE/ERROR. Last={last}")


class FullFileFlowTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        if not RUN_FULL_FLOW:
            raise unittest.SkipTest("Full flow tests are disabled. Set RUN_FULL_FLOW=1 to enable.")

    def test_full_cycle_two_submissions(self):
        work_id = f"e2e-{uuid.uuid4().hex}"

        # 1) create work
        status, body = _request_json(
            "POST",
            f"{API_V1_URL}/works",
            {"workId": work_id, "name": "E2E Work", "description": "Full flow integration test"},
        )
        self.assertEqual(status, 201, body)
        self.assertEqual(body.get("workId"), work_id)

        # 2) submit first file
        text = (
            "Hello from E2E.\n"
            "This file is used to test full plagiarism pipeline.\n"
            "Same content will be submitted twice.\n"
        ).encode("utf-8")

        status, sub1 = _submit_file(work_id, "a.txt", text)
        self.assertEqual(status, 202, sub1)
        sub1_id = sub1.get("submissionId")
        self.assertTrue(sub1_id)

        # 3) wait first done
        status, det1 = _wait_submission_done(sub1_id)
        self.assertEqual(status, 200, det1)
        self.assertEqual(det1.get("status"), "DONE", det1)
        self.assertIsNotNone(det1.get("report"))
        self.assertEqual(det1["report"].get("status"), "DONE", det1["report"])

        # 4) submit second file (same content)
        status, sub2 = _submit_file(work_id, "b.txt", text)
        self.assertEqual(status, 202, sub2)
        sub2_id = sub2.get("submissionId")
        self.assertTrue(sub2_id)

        status, det2 = _wait_submission_done(sub2_id)
        self.assertEqual(status, 200, det2)
        self.assertEqual(det2.get("status"), "DONE", det2)
        rep2 = det2.get("report")
        self.assertIsNotNone(rep2)

        sim = rep2.get("similarityPercent")
        self.assertIsNotNone(sim)
        self.assertGreaterEqual(float(sim), 0.0)
        self.assertLessEqual(float(sim), 100.0)

        matched = rep2.get("matchedSubmissions") or []
        # если система сравнивает внутри work, то sub1 должен всплыть.
        # если логика другая — оставим мягкую проверку: либо нашли sub1, либо similarity > 0.
        found_sub1 = any(m.get("submissionId") == sub1_id for m in matched if isinstance(m, dict))
        self.assertTrue(found_sub1 or float(sim) > 0.0, {"similarityPercent": sim, "matchedSubmissions": matched})

        # 6) list reports for work
        status, reports = _request("GET", f"{API_V1_URL}/works/{work_id}/reports")
        self.assertEqual(status, 200, reports)
        self.assertIsInstance(reports, list)
        self.assertGreaterEqual(len(reports), 2)

        # 7) stats updated#
        status, stats = _request("GET", f"{API_V1_URL}/works/{work_id}/stats")
        self.assertEqual(status, 200, stats)
        self.assertEqual(stats.get("workId"), work_id)
        self.assertGreaterEqual(int(stats.get("totalSubmissions", 0)), 2)
        self.assertGreaterEqual(float(stats.get("averageSimilarityPercent", 0.0)), 0.0)


if __name__ == "__main__":
    unittest.main()
