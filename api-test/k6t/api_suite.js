import http from 'k6/http';
import { check, group, sleep } from 'k6';
import { Counter } from 'k6/metrics';

// Custom metrics to track failures per endpoint
const failureCounter = new Counter('errors_per_endpoint');

// --- Configuration ---
// Management API is on https://localhost/api/v1
const MGMT_URL = 'https://localhost/api/v1';
// Cache API (Nix compatible) is on its own domain
const CACHE_URL = 'https://test.localhost'; 

const AUTH_TOKEN = 'Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoidGVzdCJ9.QlmOBM7imQkVauXII7Hd9rYAFgW6NKMuvZ4GmVSTgpM';
const CACHE_NAME = 'test';

export const options = {
  stages: [
    { duration: '30s', target: 20 },
    { duration: '1m', target: 20 },
    { duration: '30s', target: 50 },
    { duration: '1m', target: 50 },
    { duration: '30s', target: 0 },
  ],
  thresholds: {
    http_req_failed: ['rate<0.01'],
    http_req_duration: ['p(95)<500'],
  },
};

const params = {
  headers: {
    'Content-Type': 'application/json',
    'Authorization': AUTH_TOKEN,
  },
};

function logError(res, tag) {
  if (res.status !== 200 && res.status !== 201) {
    console.error(`[!] FAILED: ${tag} | Method: ${res.request.method} | URL: ${res.url} | Status: ${res.status} | Body: ${res.body}`);
    failureCounter.add(1, { endpoint: tag });
  }
}

export default function () {
  // --- Management API Group ---
  group('Management API', function () {
    // 1. Get Cache Detail (Verified exists)
    const detailRes = http.get(`${MGMT_URL}/cache/${CACHE_NAME}`, params);
    const detailCheck = check(detailRes, { 'mgmt_detail_200': (r) => r.status === 200 });
    if (!detailCheck) logError(detailRes, 'Mgmt_CacheDetail');

    // 2. Initialize Multipart Upload (Cachix flow)
    const initRes = http.post(`${MGMT_URL}/cache/${CACHE_NAME}/multipart-nar?compression=xz`, null, params);
    const initCheck = check(initRes, { 'mgmt_init_upload_200': (r) => r.status === 200 });
    if (!initCheck) logError(initRes, 'Mgmt_InitUpload');
  });

  // --- Nix/Cache API Group ---
  group('Nix Cache API', function () {
    // 3. Cache Info
    const infoRes = http.get(`${CACHE_URL}/nix-cache-info`, params);
    const infoCheck = check(infoRes, { 'nix_info_200': (r) => r.status === 200 });
    if (!infoCheck) logError(infoRes, 'Nix_CacheInfo');

    // 4. Version
    const verRes = http.get(`${CACHE_URL}/version`, params);
    const verCheck = check(verRes, { 'nix_version_200': (r) => r.status === 200 });
    if (!verCheck) logError(verRes, 'Nix_Version');
  });

  sleep(Math.random() * 2 + 1);
}
