import http from 'k6/http';
import { check, sleep } from 'k6';

// --- Configuration ---
const MGMT_URL = 'https://localhost/api/v1';
const AUTH_TOKEN = 'Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoidGVzdCJ9.QlmOBM7imQkVauXII7Hd9rYAFgW6NKMuvZ4GmVSTgpM';
const CACHE_NAME = 'test';

export const options = {
  executor: 'ramping-vus',
  stages: [
    { duration: '5m', target: 1000 }, // Ramp up from 0 to 500 users over 5 minutes
    { duration: '2m', target: 1000 }, // Hold at 500 users to see if it stabilizes
    { duration: '1m', target: 0 },   // Cool down
  ],
  thresholds: {
    http_req_failed: [{ threshold: 'rate<0.01', abortOnFail: true }],
    http_req_duration: ['p(99)<1000'],
  },
};

const params = {
  headers: {
    'Content-Type': 'application/json',
    'Authorization': AUTH_TOKEN,
  },
};

export default function () {
  // We use a light-weight endpoint for pure concurrency testing
  const res = http.get(`${MGMT_URL}/cache/${CACHE_NAME}`, params);

  check(res, {
    'status is 200': (r) => r.status === 200,
  });

  // Short sleep to prevent immediate local socket exhaustion 
  // while still maintaining high pressure.
  sleep(0.1);
}
