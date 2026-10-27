import http from 'k6/http';
import { check } from 'k6';

export const options = {
  stages: [
    { duration: '10s', target: 5},
    { duratuin: '20s', tagret: 20},
    { duration: '20s', target: 50},
    { duration: '10s', target: 0}
  ],
  thresholds: {
    http_req_failed: ['rate<0.01'],
    http_req_duration: ['p(95)<100'],
  },
};

export default function() {
  const response = http.get('http://127.0.0.1:8080/healthz');

  check(response, {
    'status is 200': (r) => r.status === 200,
  });
}
