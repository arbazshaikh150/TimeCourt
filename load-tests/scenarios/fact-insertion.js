import { uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';

import { createFactPayload } from '../data/fact-data.js';
import { postFact } from '../helper/fact-http.js';
import { checkFactInsertion } from '../helper/checks.js';

export function insertFact() {
    // Creating a payload from the data
    const payload = createFactPayload();
    // Generating the idempotency key
    const idempotencyKey = uuidv4();

    // Making an api call
    const response = postFact(
        payload,
        idempotencyKey
    );
    // Check assertion
    checkFactInsertion(response);

    return response;
}