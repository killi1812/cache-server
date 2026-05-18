#!/bin/bash

TEST_RESULT_DIR="test-results"

# TODO: add other tests

for i in {1..4}; do
echo "----- Runnig test ${i} python -----"
  ./storage-test.sh &> "${TEST_RESULT_DIR}/test_py${1}.txt"
done
