import { ruleInsertion } from '../../scenarios/rule-insertion.js';


export const options = {
    vus: 4,

    duration: '30s',

    thresholds: {
        http_req_failed: ['rate<0.01'],

        http_req_duration: [
            'p(95)<500',
            'p(99)<1000',
        ],
    },
};


export default function () {
    ruleInsertion(2);
}