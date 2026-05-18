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
CONCURRENCY = 20
FILE_SIZE_MB = 100


def run_single_worker(index, base_data):
    """Perform Upload -> Complete -> Download sequentially within a concurrent thread using an isolated session."""
    with requests.Session() as session:
        try:
            # 1. Initialize
            upload_id = init_upload(session)

            # 2. Prepare unique data
            data, file_hash = get_unique_data(base_data, "para", index)
            actual_size = len(data)

            # 3. Upload (PUT)
            perform_upload(session, upload_id, data)

            # 4. Complete (Rename to hash)
            complete_upload(session, upload_id, file_hash, actual_size, f"para{index}")

            # 5. Immediate Download (GET by hash)
            perform_download(session, file_hash)

            return True
        except Exception as e:
            print(f"Concurrent worker {index} failed: {e}")
            return False


def main():
    print(
        f"[Phase] Starting Concurrent Python Test ({CONCURRENCY}x {FILE_SIZE_MB}MB Upload -> Complete -> Download)..."
    )
    base_data = os.urandom(FILE_SIZE_MB * 1024 * 1024)

    start_time = time.time()

    with concurrent.futures.ThreadPoolExecutor(max_workers=CONCURRENCY) as executor:
        futures = [
            executor.submit(run_single_worker, i, base_data) for i in range(CONCURRENCY)
        ]
        results = [f.result() for f in concurrent.futures.as_completed(futures)]

    end_time = time.time()
    successful = [r for r in results if r]
    print(
        f"[Phase] Python Concurrent Test Finished in {end_time - start_time:.5f}s. Successful: {len(successful)}/{CONCURRENCY}"
    )


if __name__ == "__main__":
    main()
