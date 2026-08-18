import http from 'k6/http';
import { FACT_ENDPOINT } from '../config/env.js';

export function postFact(payload, idempotencyKey) {
    return http.post(
        FACT_ENDPOINT,
        JSON.stringify(payload),
        {
            headers: {
                'Content-Type': 'application/json',
                'Idempotency-Key': idempotencyKey,
            },
        }
    );
}