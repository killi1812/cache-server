import os
import time
import requests
import hashlib

# --- Shared Configuration ---
HOST = "localhost"
CACHE = os.getenv("CACHE", "test")
PROTOCOL = "https"
KEY = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoidGVzdCJ9.QlmOBM7imQkVauXII7Hd9rYAFgW6NKMuvZ4GmVSTgpM"

CACHE_DOMAIN = f"{CACHE}.{HOST}"
MGMT_DOMAIN = HOST
BASE_URL = f"{PROTOCOL}://{CACHE_DOMAIN}"
MGMT_URL = f"{PROTOCOL}://{MGMT_DOMAIN}"

# Disable insecure request warnings for self-signed certs
requests.packages.urllib3.disable_warnings()

# --- Sync Utilities (for Sequential Tests) ---


def init_upload(session=None):
    url = f"{MGMT_URL}/api/v1/cache/{CACHE}/multipart-nar?compression=xz"
    headers = {"Authorization": f"Bearer {KEY}"}
    if session:
        res = session.post(url, headers=headers, verify=False)
    else:
        res = requests.post(url, headers=headers, verify=False)
    res.raise_for_status()
    return res.json()["uploadId"]


def perform_upload(session, upload_id, data, timeout=120):
    url = f"{BASE_URL}/{upload_id}"
    headers = {"Authorization": f"Bearer {KEY}"}
    res = session.put(url, headers=headers, data=data, verify=False, timeout=timeout)
    res.raise_for_status()
    return res


def complete_upload(session, upload_id, file_hash, size, suffix, timeout=120):
    store_hash = file_hash[:32]
    url = f"{MGMT_URL}/api/v1/cache/{CACHE}/multipart-nar/{upload_id}/complete"
    headers = {"Authorization": f"Bearer {KEY}", "Content-Type": "application/json"}
    body = {
        "narInfoCreate": {
            "cFileHash": f"{file_hash}",
            "cFileSize": size,
            "cStoreHash": store_hash,
            "cStoreSuffix": f"perf-test-{suffix}",
            "cNarHash": f"sha256:{file_hash}",
            "cNarSize": size,
            "cReferences": [],
            "cDeriver": f"perf-{suffix}.drv",
            "cSig": "perfsig",
        }
    }
    res = session.post(url, headers=headers, json=body, verify=False, timeout=timeout)
    res.raise_for_status()
    return res


def perform_download(session, file_hash, timeout=120):
    url = f"{BASE_URL}/nar/{file_hash}.nar.xz"
    headers = {"Authorization": f"Bearer {KEY}"}
    res = session.get(url, headers=headers, verify=False, timeout=timeout)
    res.raise_for_status()
    return res


def get_unique_data(base_data, tag, index):
    prefix = f"{time.time_ns()}_{tag}_{index}_".encode()
    data = prefix + base_data
    file_hash = hashlib.sha256(data).hexdigest()
    return data, file_hash
