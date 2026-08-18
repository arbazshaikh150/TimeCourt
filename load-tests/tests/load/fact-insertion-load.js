import { insertFact } from '../../scenarios/fact-insertion.js';

export const options = {
    scenarios: {
        fact_insertion_load: {
            executor: 'ramping-vus',

            startVUs: 0,

            stages: [
                {
                    duration: '30s',
                    target: 10,
                },

                {
                    duration: '1m',
                    target: 10,
                },

                {
                    duration: '30s',
                    target: 25,
                },

                {
                    duration: '1m',
                    target: 25,
                },

                {
                    duration: '30s',
                    target: 50,
                },

                {
                    duration: '1m',
                    target: 50,
                },

                {
                    duration: '30s',
                    target: 0,
                },
            ],
        },
    },

    thresholds: {
        http_req_failed: [
            'rate<0.01',
        ],

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