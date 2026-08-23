import { Rate, Trend } from 'k6/metrics';

export const decisionSuccessRate = new Rate('decision_success_rate');

export const decisionLatency = new Trend(
    'decision_latency',
    true,
);

export function recordDecisionMetrics(response) {
    const success = response.status === 200;

    decisionSuccessRate.add(success);
    decisionLatency.add(response.timings.duration);
}