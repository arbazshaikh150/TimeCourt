import { checkDecision } from '../../helper/decision-http.js';
import { checkDecisionStatus } from '../../helper/checks.js';
import { createDecisionPayload } from '../../data/decision-data.js';

export const options = {
    vus: 5,
    duration: '30s',

    thresholds: {
        http_req_failed: ['rate<0.01'],
        http_req_duration: ['p(95)<500'],
        checks: ['rate>0.99'],
    },
};

export default function () {
    const payload = createDecisionPayload();

    const response = checkDecision(payload);

    checkDecisionStatus(response);
}