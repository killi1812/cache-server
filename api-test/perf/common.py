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

def init_upload():
    """Initialize a multipart upload and return the uploadId."""
    url = f"{MGMT_URL}/api/v1/cache/{CACHE}/multipart-nar?compression=xz"
    headers = {"Authorization": f"Bearer {KEY}"}
    res = requests.post(url, headers=headers, verify=False)
    res.raise_for_status()
    return res.json()["uploadId"]

def complete_upload(upload_id, file_hash, size, suffix):
    """Complete an upload to finalize the renaming logic."""
    # Use the first 32 chars of the hash for the store hash to resemble Nix store paths
    store_hash = file_hash[:32]
    
    url = f"{MGMT_URL}/api/v1/cache/{CACHE}/multipart-nar/{upload_id}/complete"
    headers = {
        "Authorization": f"Bearer {KEY}",
        "Content-Type": "application/json"
    }
    body = {
        "narInfoCreate": {
            "cFileHash": f"sha256:{file_hash}",
            "cFileSize": size,
            "cStoreHash": store_hash,
            "cStoreSuffix": f"perf-test-{suffix}",
            "cNarHash": f"sha256:{file_hash}",
            "cNarSize": size,
            "cReferences": [],
            "cDeriver": f"perf-{suffix}.drv",
            "cSig": "perfsig"
        }
    }
    res = requests.post(url, headers=headers, json=body, verify=False)
    res.raise_for_status()

def get_unique_data(base_data, tag, index):
    """Prepend a unique prefix to base data to ensure a unique SHA256 hash."""
    prefix = f"{time.time_ns()}_{tag}_{index}_".encode()
    data = prefix + base_data
    file_hash = hashlib.sha256(data).hexdigest()
    return data, file_hash
