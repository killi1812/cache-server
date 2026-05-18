#!/bin/bash

TEST_RESULT_DIR="test-results"

# TODO: add other tests

for i in {1..4}; do
  echo "----- Runnig test ${i} golang cloud -----"
  ./run-perf-suite.sh &> "${TEST_RESULT_DIR}/test_go_c_${i}.txt"
done

# echo "----- Runnig test ${i} golang local variant -----"
#   ./storage-test.sh &> "${TEST_RESULT_DIR}/test_go_l_${i}.txt"

# echo "----- Runnig test ${i} python -----"
#   ./storage-test.sh &> "${TEST_RESULT_DIR}/test_py${i}.txt"
