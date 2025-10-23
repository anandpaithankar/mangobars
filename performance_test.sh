#!/bin/bash

echo "=== Mangobars Performance Test ==="
echo

# Test with different worker counts
echo "Testing with different worker configurations:"
echo

echo "1. Default workers (auto-detected):"
time ./mangobars_enhanced -i large_hosts.csv -o perf_test_1.csv

echo
echo "2. 10 workers:"
time ./mangobars_enhanced -i large_hosts.csv -workers 10 -o perf_test_2.csv

echo
echo "3. 50 workers:"
time ./mangobars_enhanced -i large_hosts.csv -workers 50 -o perf_test_3.csv

echo
echo "4. Different timeout (1 second):"
time ./mangobars_enhanced -i large_hosts.csv -timeout 1000 -o perf_test_4.csv

echo
echo "Performance test completed!"