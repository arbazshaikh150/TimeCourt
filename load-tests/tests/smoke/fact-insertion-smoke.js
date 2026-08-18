import { insertFact } from '../../scenarios/fact-insertion.js';

export const options = {
    vus: 1,
    iterations: 5,
    thresholds: {
        http_req_failed: ['rate<0.01'],
        checks: ['rate>0.99'],
    },
};

export default function () {
    insertFact();
}