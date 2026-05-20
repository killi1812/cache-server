import http from 'k6/http';
import { check, sleep } from 'k6';

// 1. Define the load testing structure (Ramping Profile)
export const options = {
  stages: [
    { duration: '1m', target: 20 },  // Ramp up from 1 to 20 users over 1 minute
    { duration: '3m', target: 20 },  // Stay at 20 users for 3 minutes (Sustained Load)
    { duration: '1m', target: 50 },  // Ramp up further to 50 users over 1 minute (Stress Peak)
    { duration: '2m', target: 50 },  // Hold at 50 users for 2 minutes
    { duration: '1m', target: 0 },   // Scale down to 0 users over 1 minute (Cool Down)
  ],
  thresholds: {
    http_req_failed: ['rate<0.01'],   // Error rate must be less than 1%
    http_req_duration: ['p(95)<500'],  // 95% of requests must complete under 500ms
  },
};

// 2. Define the simulated user behavior
export default function () {
  const url = 'https://localhost/api/v1/cache/test'; // Replace with your target API endpoint

  const params = {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': 'Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoidGVzdCJ9.QlmOBM7imQkVauXII7Hd9rYAFgW6NKMuvZ4GmVSTgpM',
    },
  };

  // Send a GET request
  const response = http.get(url, params);

  // Validate that the server returned a 200 OK status
  check(response, {
    'status is 200': (r) => r.status === 200,
  });

  // Pause for 1 second between iterations to simulate human pacing
  sleep(1);
}
