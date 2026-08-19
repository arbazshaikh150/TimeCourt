import { randomEffectivePeriod } from "./fact-data.js";
import { uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';
const FACT_POOL = [
    {
        fact_key: 'customer.phone',
        fact_required_value: '+91-9876543210',
    },
    {
        fact_key: 'customer.status',
        fact_required_value: 'ACTIVE',
    },
    {
        fact_key: 'customer.country',
        fact_required_value: 'IN',
    },
    {
        fact_key: 'customer.type',
        fact_required_value: 'PREMIUM',
    },
    {
        fact_key: 'customer.verified',
        fact_required_value: 'true',
    },
];

function createFactsKeys(count = 2) {
    return FACT_POOL.slice(0, count);
}

export function createRulePayload(factcount = 2) {
    const { startTime, endTime } = randomEffectivePeriod();

    return {
        tenant_id: 'tenant-001',

        // Keep the business meaning fixed,
        // but make every rule logically unique if required.
        rule_key: `customer.validation.${uuidv4()}`,

        knowledge_time: new Date().toISOString(),

        effective_start_time: startTime,

        effective_end_time: endTime,

        resolution_version: 1,

        interpretor_version: 1,

        source: 'load-test',

        facts_keys: createFactsKeys(factcount),
    };
}