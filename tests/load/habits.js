import http from 'k6/http';
import { check } from 'k6';

const token = __ENV.TOKEN;
const targetRPS = Number(__ENV.RPS || 100);

if (!token) {
  throw new Error('TOKEN is required');
}

export const options = {
  scenarios: {
    habits_load: {
      executor: 'constant-arrival-rate',
      rate: targetRPS,
      timeUnit: '1s',
      duration: '30s',
      preAllocatedVUs: 20,
      maxVUs: 100,
    },
  },
  
  thresholds: {
    http_req_failed: ['rate<0.01'],
    http_req_duration: ['p(95)<100'],
  },
};

export default function () {
  const response = http.get(
    'http://127.0.0.1:8080/habits',
    {
      headers: {
        Authorization: `Bearer ${token}`,
      },
    },
  );

  check(response, {
    'status is 200': (r) => r.status === 200,
  });
}
