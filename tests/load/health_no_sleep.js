import http from 'k6/http';
import { check } from 'k6';

export const options = {
	vus: 5,
	duration: '20s',
};

export default function() {
	const response = http.get('http://127.0.0.1:8080/healthz');
	
	check(response, {
		'status is 200': (r) => r.status === 200,
	});
};
