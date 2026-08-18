import { insertFact } from '../../scenarios/fact-insertion.js';

export const options = {
    scenarios: {
        fact_insertion_baseline: {
            executor: 'constant-vus',

            vus: 5,

            duration: '1m',
        },
    },

    thresholds: {
        http_req_failed: ['rate<0.01'],

        http_req_duration: [
            'p(95)<500',
        ],

        checks: [
            'rate>0.99',
        ],
    },
};

export default function () {
    insertFact();
}