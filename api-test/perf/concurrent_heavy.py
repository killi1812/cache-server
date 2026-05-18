import os
import time
import requests
import concurrent.futures
from common import init_upload, complete_upload, get_unique_data, BASE_URL, KEY

# --- Configuration ---
CONCURRENCY = 50
FILE_SIZE_MB = 100

def run_single_worker(index, base_data):
    """Perform Upload -> Complete -> Download sequentially within a concurrent thread."""
    try:
        # 1. Initialize
        upload_id = init_upload()
        
        # 2. Prepare unique data
        data, file_hash = get_unique_data(base_data, "para", index)
        actual_size = len(data)
        
        # 3. Upload (PUT)
        upload_url = f"{BASE_URL}/{upload_id}"
        headers = {"Authorization": f"Bearer {KEY}"}
        requests.put(upload_url, headers=headers, data=data, verify=False).raise_for_status()
        
        # 4. Complete (Rename to hash)
        complete_upload(upload_id, file_hash, actual_size, f"para{index}")
        
        # 5. Immediate Download (GET by hash)
        download_url = f"{BASE_URL}/nar/{file_hash}.nar.xz"
        requests.get(download_url, headers=headers, verify=False).raise_for_status()
        
        return True
    except Exception as e:
        print(f"Worker {index} failed during Upload/Complete/Download: {e}")
        return False

def main():
    print(f"[Phase] Starting Concurrent Python Test ({CONCURRENCY}x {FILE_SIZE_MB}MB Upload -> Complete -> Download)...")
    
    # Pre-generate base data
    base_data = os.urandom(FILE_SIZE_MB * 1024 * 1024)
    
    start_time = time.time()
    
    with concurrent.futures.ThreadPoolExecutor(max_workers=CONCURRENCY) as executor:
        futures = [executor.submit(run_single_worker, i, base_data) for i in range(CONCURRENCY)]
        results = [f.result() for f in concurrent.futures.as_completed(futures)]
    
    end_time = time.time()
    successful = [r for r in results if r]
    print(f"[Phase] Python Concurrent Test Finished in {end_time - start_time:.2f}s. Successful: {len(successful)}/{CONCURRENCY}")

if __name__ == "__main__":
    main()
