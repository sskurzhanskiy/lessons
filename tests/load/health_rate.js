import http from 'k6/http';
import { check } from 'k6';

const targetRPS = Number(__ENV.RPS || 1000)

export const options = {
  scenarios: {
    constant_load: {
      executor: 'constant-arrival-rate',
      rate: targetRPS,
      timeUnit: '1s',
      duration: '30s',
      preAllocatedVUs: 50,
      maxVUs: 200,
    },
  },
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
