import os
import time
import requests
import concurrent.futures
from common import init_upload, get_unique_data, BASE_URL, KEY

# --- Configuration ---
FILE_SIZE_MB = 1000
COUNT = 5


def run_single_large(index, base_data):
    try:
        # 1. Initialize
        upload_id = init_upload()

        # 2. Prepare unique data
        data, _ = get_unique_data(base_data, "large", index)

        # 3. Upload (Timed)
        upload_url = f"{BASE_URL}/{upload_id}"
        headers = {"Authorization": f"Bearer {KEY}"}
        requests.put(
            upload_url, headers=headers, data=data, verify=False
        ).raise_for_status()
        return True
    except Exception as e:
        print(f"Large worker {index} failed: {e}")
        return False


def main():
    print(f"[Phase] Starting Parallel Large ({COUNT}x {FILE_SIZE_MB}MB Uploads)...")

    # Pre-generate base data
    base_data = os.urandom(FILE_SIZE_MB * 1024 * 1024)

    start_time = time.time()

    with concurrent.futures.ThreadPoolExecutor(max_workers=COUNT) as executor:
        futures = [
            executor.submit(run_single_large, i, base_data) for i in range(COUNT)
        ]
        results = [f.result() for f in concurrent.futures.as_completed(futures)]

    end_time = time.time()
    successful = [r for r in results if r]
    print(
        f"[Phase] Parallel Large Finished in {end_time - start_time:.5f}s. Successful: {len(successful)}/{COUNT}"
    )


if __name__ == "__main__":
    main()
