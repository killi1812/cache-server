import os
import time
import requests
import concurrent.futures
from common import init_upload, complete_upload, get_unique_data, BASE_URL, KEY

# --- Configuration ---
COUNT = 50
FILE_SIZE_MB = 100


def upload_worker(index, base_data):
    try:
        upload_id = init_upload()
        data, file_hash = get_unique_data(base_data, "burst", index)
        actual_size = len(data)
        
        upload_url = f"{BASE_URL}/{upload_id}"
        headers = {"Authorization": f"Bearer {KEY}"}
        requests.put(
            upload_url, headers=headers, data=data, verify=False
        ).raise_for_status()

        # Finalize upload to enable download
        complete_upload(upload_id, file_hash, actual_size, f"burst{index}")

        return file_hash
    except Exception as e:
        print(f"Upload worker {index} failed: {e}")
        return None


def download_worker(file_hash):
    try:
        url = f"{BASE_URL}/nar/{file_hash}.nar.xz"
        headers = {"Authorization": f"Bearer {KEY}"}
        requests.get(url, headers=headers, verify=False).raise_for_status()
        return True
    except Exception as e:
        print(f"Download of {file_hash} failed: {e}")
        return False


def main():
    print(f"[Phase] Starting Sequential Burst ({COUNT}x {FILE_SIZE_MB}MB)...")
    base_data = os.urandom(FILE_SIZE_MB * 1024 * 1024)

    # 1. Parallel Upload Phase
    print(f"--- 1. Parallel Uploads ({COUNT} workers) ---")
    start_time = time.time()
    with concurrent.futures.ThreadPoolExecutor(max_workers=10) as executor:
        futures = [executor.submit(upload_worker, i, base_data) for i in range(COUNT)]
        hashes = [f.result() for f in concurrent.futures.as_completed(futures)]

    hashes = [h for h in hashes if h]
    end_time = time.time()
    print(f"Upload Phase Finished in {end_time - start_time:.2f}s")

    # 2. Parallel Download Phase (After all uploads)
    if hashes:
        print(f"--- 2. Parallel Downloads ({len(hashes)} workers) ---")
        start_time = time.time()
        with concurrent.futures.ThreadPoolExecutor(max_workers=10) as executor:
            futures = [executor.submit(download_worker, h) for h in hashes]
            results = [f.result() for f in concurrent.futures.as_completed(futures)]
        end_time = time.time()
        print(f"Download Phase Finished in {end_time - start_time:.2f}s")


if __name__ == "__main__":
    main()
