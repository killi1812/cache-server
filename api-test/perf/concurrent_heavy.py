import os
import time
import asyncio
import httpx
import random
from common import (
    get_unique_data,
    MGMT_URL,
    CACHE,
    BASE_URL,
    KEY,
)

# --- Configuration ---
CONCURRENCY = 20
FILE_SIZE_MB = 100

# --- Async Utilities (for Concurrent Test) ---


async def async_init_upload(client):
    url = f"{MGMT_URL}/api/v1/cache/{CACHE}/multipart-nar?compression=xz"
    res = await client.post(url)
    res.raise_for_status()
    return res.json()["uploadId"]


async def async_perform_upload(client, upload_id, data, timeout=120):
    url = f"{BASE_URL}/{upload_id}"
    res = await client.put(url, content=data, timeout=timeout)
    res.raise_for_status()
    return res


async def async_complete_upload(
    client, upload_id, file_hash, size, suffix, timeout=120
):
    store_hash = file_hash[:32]
    url = f"{MGMT_URL}/api/v1/cache/{CACHE}/multipart-nar/{upload_id}/complete"
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
            "cSig": "perfsig",
        }
    }
    res = await client.post(url, json=body, timeout=timeout)
    res.raise_for_status()
    return res


async def async_perform_download(client, file_hash, timeout=120):
    url = f"{BASE_URL}/nar/{file_hash}.nar.xz"
    res = await client.get(url, timeout=timeout)
    res.raise_for_status()
    return res


async def run_single_worker(index, base_data, client, loop):
    """Perform Upload -> Complete -> Download with random jitter to ensure mixing."""
    # 0. Random jitter (0 to 3 seconds) to desynchronize workers
    await asyncio.sleep(random.uniform(0.1, 0.5))

    try:
        # 1. Initialize
        print(f"[*] Worker {index:02d}: Starting INIT")
        upload_id = await async_init_upload(client)

        # 2. Prepare unique data (Offload SHA256 to thread to avoid blocking event loop)
        # Hashing 100MB is CPU-intensive.
        data, file_hash = await loop.run_in_executor(
            None, get_unique_data, base_data, "para", index
        )
        actual_size = len(data)

        # 3. Upload (PUT)
        print(f"[*] Worker {index:02d}: Starting UPLOAD ({file_hash[:8]}...)")
        await async_perform_upload(client, upload_id, data)

        # 4. Complete (Rename to hash)
        print(f"[*] Worker {index:02d}: Starting COMPLETE")
        await async_complete_upload(
            client, upload_id, file_hash, actual_size, f"para{index}"
        )

        # 5. Immediate Download (GET by hash)
        print(f"[*] Worker {index:02d}: Starting DOWNLOAD")
        await async_perform_download(client, file_hash)

        print(f"[+] Worker {index:02d}: SUCCESS")
        return True
    except Exception as e:
        print(f"[!] Worker {index:02d} FAILED: {e}")
        return False


async def main():
    print(
        f"[Phase] Starting Shuffled Async Test ({CONCURRENCY}x {FILE_SIZE_MB}MB Mixed Flow)..."
    )

    # Pre-generate base data
    base_data = os.urandom(FILE_SIZE_MB * 1024 * 1024)

    start_time = time.time()
    loop = asyncio.get_running_loop()

    # Connection limits
    limits = httpx.Limits(
        max_connections=CONCURRENCY, max_keepalive_connections=CONCURRENCY
    )
    headers = {"Authorization": f"Bearer {KEY}"}

    async with httpx.AsyncClient(
        verify=False, limits=limits, headers=headers, timeout=180.0
    ) as client:
        tasks = [
            run_single_worker(i, base_data, client, loop) for i in range(CONCURRENCY)
        ]
        results = await asyncio.gather(*tasks)

    end_time = time.time()
    successful = [r for r in results if r]
    print(f"\n[Phase] Shuffled Async Test Finished in {end_time - start_time:.5f}s.")
    print(f"Final Score: {len(successful)}/{CONCURRENCY} successful.")


if __name__ == "__main__":
    asyncio.run(main())
