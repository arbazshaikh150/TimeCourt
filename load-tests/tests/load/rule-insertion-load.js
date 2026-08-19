import { ruleInsertion } from '../../scenarios/rule-insertion.js';


export const options = {
    stages: [
        {
            duration: '30s',
            target: 10,
        },
        {
            duration: '1m',
            target: 25,
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

    thresholds: {
        http_req_failed: [
            'rate<0.01',
        ],

        http_req_duration: [
            'p(95)<500',
            'p(99)<2000',
        ],
    },
};


export default function () {
    ruleInsertion(2);
}