import { createDecisionPayload } from "../../data/decision-data.js";
import { checkDecisionStatus } from "../../helper/checks.js";
import { checkDecision } from "../../helper/decision-http.js";
import { recordDecisionMetrics } from "../../metrics/decision-metrics.js";
import { createDecisionSummary } from "../../reports/decision-report.js";

export const options = {
    scenarios: {
        decision_load: {
            executor: 'constant-arrival-rate',

            rate: 1000,
            timeUnit: '1s',

            duration: '60s',

            preAllocatedVUs: 50,
            maxVUs: 500,
        },
    },

    thresholds: {
        http_req_failed: [
            'rate<0.01',
        ],

        http_req_duration: [
            'p(50)<100',
            'p(95)<250',
            'p(99)<500',
        ],

        checks: [
            'rate>0.99',
        ],

        decision_success_rate: [
            'rate>0.99',
        ],
    },
};

export function setup() {
    return createDecisionPayload();
}

export default function (data) {
    const payload = data;

    const response = checkDecision(payload);

    checkDecisionStatus(response);

    recordDecisionMetrics(response);
}

export function handleSummary(data) {
    const summary = createDecisionSummary(data);

    return {
        stdout: summary,
    };
}
