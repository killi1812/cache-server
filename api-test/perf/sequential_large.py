import os
import time
import requests
import concurrent.futures
from common import (
    init_upload,
    perform_upload,
    complete_upload,
    perform_download,
    get_unique_data,
)

# --- Configuration ---
COUNT = 5
FILE_SIZE_MB = 1000


def upload_worker(index, base_data):
    """Worker to perform a full upload lifecycle (Init -> Put -> Complete)."""
    with requests.Session() as session:
        try:
            # 1. Initialize
            upload_id = init_upload(session)

            # 2. Prepare unique data
            data, file_hash = get_unique_data(base_data, "large", index)
            actual_size = len(data)

            # 3. Upload (PUT)
            perform_upload(session, upload_id, data)

            # 4. Complete (Rename to hash)
            complete_upload(session, upload_id, file_hash, actual_size, f"large{index}")

            return file_hash
        except Exception as e:
            print(f"Large Upload worker {index} failed: {e}")
            return None


def download_worker(file_hash):
    """Worker to perform a download by hash."""
    with requests.Session() as session:
        try:
            perform_download(session, file_hash)
            return True
        except Exception as e:
            print(f"Large Download of {file_hash} failed: {e}")
            return False


def main():
    print(f"[Phase] Starting Sequential Large ({COUNT}x {FILE_SIZE_MB}MB)...")
    base_data = os.urandom(FILE_SIZE_MB * 1024 * 1024)

    # 1. Parallel Upload Phase
    print(f"--- 1. Parallel Uploads ({COUNT} workers) ---")
    start_time = time.time()
    with concurrent.futures.ThreadPoolExecutor(max_workers=COUNT) as executor:
        futures = [executor.submit(upload_worker, i, base_data) for i in range(COUNT)]
        hashes = [f.result() for f in concurrent.futures.as_completed(futures)]

    hashes = [h for h in hashes if h]
    end_time = time.time()
    print(f"Upload Phase Finished in {end_time - start_time:.2f}s")

    # 2. Parallel Download Phase (After all uploads)
    if hashes:
        print(f"--- 2. Parallel Downloads ({len(hashes)} workers) ---")
        start_time = time.time()
        with concurrent.futures.ThreadPoolExecutor(max_workers=COUNT) as executor:
            futures = [executor.submit(download_worker, h) for h in hashes]
            results = [f.result() for f in concurrent.futures.as_completed(futures)]
        end_time = time.time()
        print(f"Download Phase Finished in {end_time - start_time:.5f}s")


if __name__ == "__main__":
    main()
