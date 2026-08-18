import { uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';
function randomEffectivePeriod() {
    // Start somewhere between Jan 1, 2026 and Dec 31, 2026
    const startOfYear = new Date('2026-01-01T00:00:00Z').getTime();
    const endOfYear = new Date('2026-12-31T23:59:59Z').getTime();

    // Pick a random start time
    const startTime =
        startOfYear +
        Math.random() * (endOfYear - startOfYear);

    // Random duration between 1 hour and 180 days
    const minDuration = 60 * 60 * 1000;
    const maxDuration = 180 * 24 * 60 * 60 * 1000;

    const duration =
        minDuration +
        Math.random() * (maxDuration - minDuration);

    const endTime = startTime + duration;

    return {
        startTime: new Date(startTime).toISOString(),
        endTime: new Date(endTime).toISOString(),
    };
}

// Creating the fact payload
// TODO : Knowledge time must be updated from the client side only
export function createFactPayload() {
    const { startTime, endTime } = randomEffectivePeriod();

    return {
        tenant_id: 'tenant-001',

        fact_key: 'customer.phone',

        fact_version: 1,

        subject_id: `customer-k6-${uuidv4()}`,

        fact_effective_start_time: startTime,

        fact_effective_end_time: endTime,

        knowledge_time: new Date().toISOString(),

        fact_value: '+91-9876543210',

        authority: 'customer',

        confidence: 'HIGH',

        source: 'load-test',
    };
}