import {
    createRulePayload,
} from '../data/rule-data.js';
import { uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';
import { createRule } from '../helper/rule-http.js';

import {
    checkRuleCreation,
} from '../helper/checks.js';


export function ruleInsertion(factCount = 3) {

    const payload =
        createRulePayload(factCount);

    const idempotencyKey =
        uuidv4();

    const response =
        createRule(
            payload,
            idempotencyKey
        );

    checkRuleCreation(response);
}