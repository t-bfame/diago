base=$(minikube ip)

pad() {
  printf "\n\n\n"
}

get_test() {
  curl "$base:30007/api/tests/$testid" | python3 -m json.tool
  pad
}

submit_test() {
  curl "$base:30007/api/start-test/$testid" | python3 -m json.tool
  pad
}

stop_test() {
  curl "$base:30007/api/stop-test/$testid" | python3 -m json.tool
  pad
}

get_instance() {
  curl "$base:30007/api/test-instances/$testid" | python3 -m json.tool
  pad
}

# testid=""
# pad
# make_dummy_test2

# # strip quotes from id
# testid=$(echo $testid | tr -d '"')
# echo "created test with id: $testid"
# pad

# get_test
# submit_test
# get_instance
# # stop_test
# get_instance

# sleep 5

testid="Test2"

echo "starting test with id: $testid"
pad

get_test
submit_test

sleep 30

# stop_test
get_instance
