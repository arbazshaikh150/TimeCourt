import { ruleInsertion } from '../../scenarios/rule-insertion.js';


export const options = {
    vus: 1,

    iterations: 6,

    thresholds: {
        http_req_failed: ['rate<0.01'],
        http_req_duration: ['p(95)<1000'],
    },
};


export default function () {
    ruleInsertion(3);
}