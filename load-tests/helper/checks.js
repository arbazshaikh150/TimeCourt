import { check } from 'k6';

export function checkFactInsertion(response) {
    return check(response, {
        'fact insertion returned 2xx': (r) =>
            r.status >= 200 && r.status <= 300,
    });
}