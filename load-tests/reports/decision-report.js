function formatNumber(value, decimals = 2) {
    if (value === undefined || value === null || Number.isNaN(value)) {
        return 'N/A';
    }

    return Number(value).toFixed(decimals);
}

function formatInteger(value) {
    if (value === undefined || value === null || Number.isNaN(value)) {
        return 'N/A';
    }

    return Number(value).toLocaleString();
}

export function createDecisionSummary(data) {
    const requests = data.metrics.http_reqs?.values || {};
    const duration = data.metrics.http_req_duration?.values || {};
    const failed = data.metrics.http_req_failed?.values || {};

    const actualRps = requests.rate;
    const totalRequests = requests.count;

    // k6 exposes median as `med`, not `p(50)`.
    const p50 = duration.med;
    const p95 = duration['p(95)'];
    const p99 = duration['p(99)'];

    const failedRate =
        failed.rate !== undefined && failed.rate !== null
            ? failed.rate * 100
            : null;

    const targetRps = 1000;
    const durationSeconds = 60;
    const expectedRequests = targetRps * durationSeconds;

    return `
                    DECISION LOAD TEST

Target
------
Target RPS         : ${formatNumber(targetRps, 0)}
Duration           : ${durationSeconds}s
Expected Requests  : ${formatInteger(expectedRequests)}

Throughput
----------
Actual RPS         : ${formatNumber(actualRps)}
Total Requests     : ${formatInteger(totalRequests)}

Latency
-------
p50                : ${formatNumber(p50)} ms
p95                : ${formatNumber(p95)} ms
p99                : ${formatNumber(p99)} ms

Reliability
-----------
Failed Requests    : ${formatNumber(failedRate)}%

Conclusion
----------
At a target load of ${formatNumber(targetRps, 0)} requests/second,
the system achieved ${formatNumber(actualRps)} requests/second.

p99 latency        : ${formatNumber(p99)} ms
Failure rate       : ${formatNumber(failedRate)}%

`;
}