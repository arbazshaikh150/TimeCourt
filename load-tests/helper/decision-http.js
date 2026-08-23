import http from 'k6/http';
import { DECISION_ENDPOINT } from '../config/env.js';

export function checkDecision(payload) {
    const params = {
        rule_key: payload.rule_key,
        subject_id: payload.subject_id,
        time_when_to_check: payload.time_when_to_check,
        time_where_to_check: payload.time_where_to_check,
    };

    const query = Object.entries(params)
        .map(([key, value]) => `${key}=${encodeURIComponent(value)}`)
        .join('&');

    return http.get(`${DECISION_ENDPOINT}?${query}`);
}