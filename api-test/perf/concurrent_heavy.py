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

    current_step = "INIT"
    try:
        # 1. Initialize
        upload_id = await async_init_upload(client)

        # 2. Prepare unique data (Offload SHA256 to thread to avoid blocking event loop)
        # Hashing 100MB is CPU-intensive.
        data, file_hash = await loop.run_in_executor(
            None, get_unique_data, base_data, "para", index
        )
        actual_size = len(data)

        current_step = "UPLOAD"
        # 3. Upload (PUT)
        await async_perform_upload(client, upload_id, data)

        current_step = "COMPLEATE"
        # 4. Complete (Rename to hash)
        await async_complete_upload(
            client, upload_id, file_hash, actual_size, f"para{index}"
        )

        current_step = "DOWNLOAD"
        # 5. Immediate Download (GET by hash)
        await async_perform_download(client, file_hash)

        current_step = "DONE"
        return True
    except httpx.HTTPStatusError as e:
        print(
            f"[!] Worker {index:02d} FAILED at [{current_step}]: HTTP {e.response.status_code} - {e.response.text[:200]}"
        )
        return False
    except httpx.RequestError as e:
        print(f"[!] Worker {index:02d} FAILED at [{current_step}]: {e.request}")
        # TODO: retry go to step
        return False
    except Exception as e:
        print(f"[!] Worker {index:02d} FAILED at [{current_step}]: {e}")
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
        verify=False, limits=limits, headers=headers, timeout=None
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
