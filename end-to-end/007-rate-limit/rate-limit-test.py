#!/usr/bin/env python

import sys
import time
import requests


class QotM(object):
    def __init__(self, target):
        self.url = "http://%s/qotm/" % target

    def decipher(self, r):
        return r.status_code

    def get(self, headers):
        return self.decipher(requests.get(self.url, headers=headers))


def test_qotm_ratelimit(base, test_list, iterations=100, target_success_rate=0.9):
    q = QotM(base)
    ran = 0
    succeeded = 0

    for iteration in range(iterations):
        for headers, expected_code in test_list:
            ran += 1
            code = q.get(headers)
            if code == expected_code:
                succeeded += 1
            else:
                print("%s: expected %d, got %d" % (headers, expected_code, code))

    success_rate = (succeeded / ran)
    sys.stdout.write("\n")
    print("Ran           %d" % ran)
    print("Succeeded     %d" % succeeded)
    print("Failed        %d" % (ran - succeeded))
    print("Success rate  %f%%" % (success_rate))

    # This is a bit flaky, requests are sampled by Envoy and could timeout
    return 0 if (success_rate > target_success_rate) else 1


if __name__ == "__main__":
    base = sys.argv[1]

    test_list = []

    # No matching headers, won't even go through ratelimit-service filter
    test_list.append((None, 200))
    # Header instructing dummy ratelimit-service to allow request
    test_list.append(({'x-ambassador-test-allow': 'true'}, 200))
    # Header instructing dummy ratelimit-service to reject request
    test_list.append(({'x-ambassador-test-allow': 'over my dead body'}, 429))

    sys.exit(test_qotm_ratelimit(base, test_list))
