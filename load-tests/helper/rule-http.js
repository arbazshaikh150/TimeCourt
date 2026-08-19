import http from 'k6/http';

import {
    RULE_ENDPOINT,
} from '../config/env.js';


export function createRule(
    payload,
    idempotencyKey
) {
    const params = {
        headers: {
            'Content-Type': 'application/json',

            'Idempotency-Key': idempotencyKey,
        },
    };

    return http.post(
        RULE_ENDPOINT,
        JSON.stringify(payload),
        params
    );
}