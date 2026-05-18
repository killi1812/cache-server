#!/bin/python
import os
import time
import requests
import hashlib
import concurrent.futures

# --- Configuration ---
HOST = "localhost"
CACHE = os.getenv("CACHE", "test")
PROTOCOL = "https"
KEY = os.getenv("CACHETOKEN")

CACHE_DOMAIN = f"{CACHE}.{HOST}"
MGMT_DOMAIN = HOST
BASE_URL = f"{PROTOCOL}://{CACHE_DOMAIN}"
MGMT_URL = f"{PROTOCOL}://{MGMT_DOMAIN}"

CONCURRENCY = 30
FILE_SIZE_MB = 100

# Pre-generate 100MB of random data once to save time
print(f"Generating {FILE_SIZE_MB}MB base data...")
base_data = os.urandom(FILE_SIZE_MB * 1024 * 1024)


def init_upload():
    url = f"{MGMT_URL}/api/v1/cache/{CACHE}/multipart-nar?compression=xz"
    headers = {"Authorization": f"Bearer {KEY}"}
    res = requests.post(url, headers=headers, verify=False)
    res.raise_for_status()
    return res.json()["uploadId"]


def complete_upload(upload_id, file_hash, size, index):
    url = f"{MGMT_URL}/api/v1/cache/{CACHE}/multipart-nar/{upload_id}/complete"
    headers = {"Authorization": f"Bearer {KEY}", "Content-Type": "application/json"}
    body = {
        "narInfoCreate": {
            "cFileHash": file_hash,
            "cFileSize": size,
            "cStoreHash": f"para{index}",
            "cStoreSuffix": f"file{index}",
            "cNarHash": f"sha256:{file_hash}",
            "cNarSize": size,
            "cReferences": [],
            "cDeriver": f"para{index}.drv",
            "cSig": "perfsig",
        }
    }
    res = requests.post(url, headers=headers, json=body, verify=False)
    res.raise_for_status()


def run_single_test(index):
    try:
        # 1. Initialize
        upload_id = init_upload()

        # 2. Prepare unique data
        prefix = f"{time.time_ns()}_para_{index}_".encode()
        unique_data = prefix + base_data
        actual_size = len(unique_data)
        file_hash = hashlib.sha256(unique_data).hexdigest()

        # 3. Upload (PUT)
        upload_url = f"{BASE_URL}/{upload_id}"
        headers = {"Authorization": f"Bearer {KEY}"}
        put_res = requests.put(
            upload_url, headers=headers, data=unique_data, verify=False
        )
        put_res.raise_for_status()

        # 4. Complete
        complete_upload(upload_id, file_hash, actual_size, index)

        return file_hash
    except Exception as e:
        print(f"Worker {index} failed: {e}")
        return None


def main():
    print(
        f"--- Running Concurrent Heavy Python Test ({CONCURRENCY}x {FILE_SIZE_MB}MB) ---"
    )

    start_time = time.time()

    # Use ThreadPoolExecutor for concurrent requests
    with concurrent.futures.ThreadPoolExecutor(max_workers=CONCURRENCY) as executor:
        # Submit all tasks
        futures = [executor.submit(run_single_test, i) for i in range(CONCURRENCY)]

        # Wait for all to complete
        hashes = [f.result() for f in concurrent.futures.as_completed(futures)]

    end_time = time.time()
    duration = end_time - start_time

    successful = [h for h in hashes if h is not None]
    print(f"Finished {len(successful)}/{CONCURRENCY} uploads in {duration:.2f}s")

    # Optional: Concurrent Download verification
    if successful:
        print(f"Starting concurrent downloads of {len(successful)} unique files...")
        dl_start = time.time()

        def download_file(h):
            url = f"{BASE_URL}/nar/{h}.nar.xz"
            headers = {"Authorization": f"Bearer {KEY}"}
            r = requests.get(url, headers=headers, verify=False)
            r.raise_for_status()
            return len(r.content)

        with concurrent.futures.ThreadPoolExecutor(max_workers=CONCURRENCY) as executor:
            dl_futures = [executor.submit(download_file, h) for h in successful]
            results = [f.result() for f in concurrent.futures.as_completed(dl_futures)]

        dl_end = time.time()
        print(f"Concurrent Download finished in {dl_end - dl_start:.2f}s")


if __name__ == "__main__":
    # Disable insecure request warnings for self-signed certs
    requests.packages.urllib3.disable_warnings()
    main()
