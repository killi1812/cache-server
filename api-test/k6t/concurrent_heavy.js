import http from 'k6/http';
import { check, sleep } from 'k6';
import crypto from 'k6/crypto';

// --- Configuration ---
const MGMT_URL = 'https://localhost/api/v1';
const CACHE_URL = 'https://test.localhost';
const AUTH_TOKEN = 'Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoidGVzdCJ9.Pijr3YxjmBgUGuPKox3zwxro6gjfgdoul36XFHUH1Ro';;
const CACHE_NAME = 'test';
const FILE_SIZE_MB = 100;

export const options = {
  stages: [
    { duration: '1m', target: 20 },
    { duration: '3m', target: 20 },
    { duration: '1m', target: 0 },
  ],
  thresholds: {
    http_req_failed: ['rate<0.05'],
  },
};

const sharedData = crypto.randomBytes(FILE_SIZE_MB * 1024 * 1024);

export default function () {
  const index = __VU;
  const iteration = __ITER;

  const fileHash = crypto.sha256(sharedData, 'hex');
  const actualSize = sharedData.byteLength;

  const params = {
    headers: {
      'Authorization': AUTH_TOKEN,
      'Content-Type': 'application/json',
    },
    timeout: '300s',
  };

  // --- Step 1: INIT ---
  const initRes = http.post(`${MGMT_URL}/cache/${CACHE_NAME}/multipart-nar?compression=xz`, null, params);
  if (!check(initRes, { 'init success': (r) => r.status === 200 })) return;
  const uploadId = initRes.json().uploadId;

  // --- Step 2: UPLOAD ---
  const uploadParams = {
    headers: {
      'Authorization': AUTH_TOKEN,
      'Content-Type': 'application/octet-stream',
    },
    timeout: '300s',
  };

  // We pass the shared buffer directly.
  const uploadRes = http.put(`${CACHE_URL}/${uploadId}`, sharedData, uploadParams);
  if (!check(uploadRes, { 'upload success': (r) => r.status === 201 })) return;

  // --- Step 3: COMPLETE ---
  // We append index and iteration to suffix to ensure unique DB entries even if data hash is same
  const completeBody = JSON.stringify({
    narInfoCreate: {
      cFileHash: fileHash,
      cFileSize: actualSize,
      cStoreHash: fileHash.substring(0, 32),
      cStoreSuffix: `k6-vu${index}-it${iteration}`,
      cNarHash: `sha256:${fileHash}`,
      cNarSize: actualSize,
      cReferences: [],
      cDeriver: `k6-vu${index}.drv`,
      cSig: 'perfsig',
    },
  });
  const completeRes = http.post(`${MGMT_URL}/cache/${CACHE_NAME}/multipart-nar/${uploadId}/complete`, completeBody, params);
  if (!check(completeRes, { 'complete success': (r) => r.status === 200 })) return;

  // --- Step 4: DOWNLOAD ---
  const downloadRes = http.get(`${CACHE_URL}/nar/${fileHash}.nar.xz`, params);
  check(downloadRes, { 'download success': (r) => r.status === 200 });

  sleep(Math.random() * 2 + 1);
}
