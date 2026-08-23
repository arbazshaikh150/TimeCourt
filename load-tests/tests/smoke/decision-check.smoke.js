import { checkDecision } from '../../helper/decision-http.js';
import { checkDecisionStatus } from '../../helper/checks.js';
import { createDecisionPayload } from '../../data/decision-data.js';

export const options = {
    vus: 1,
    iterations: 10,
};

export default function () {
    const payload = createDecisionPayload();

    const response = checkDecision(payload);

    checkDecisionStatus(response);
}