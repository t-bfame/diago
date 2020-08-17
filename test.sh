base=$(minikube ip)

pad() {
  printf "\n\n\n"
}

make_dummy_test() {
  output=$(curl  -v "$base:30007/tests"\
        -H "Content-Type: application/json"\
        -d "{
            \"Name\": \"Test1\",
            \"Jobs\": [
              {
                \"ID\": \"1\",
                \"Name\": \"alpha\",
                \"Group\": \"diago-worker\",
                \"Priority\": 0,
                \"Frequency\":  10,
			          \"Duration\":   10,
			          \"HTTPMethod\": \"GET\",
			          \"HTTPUrl\":    \"https://www.google.com\"
              }
            ]
          }")
  echo $output
  testid=$(echo $output | jq '.payload.testid')
}

get_test() {
  curl  -v "$base:30007/tests/$testid"
  pad
}

submit_test() {
  curl  -v "$base:30007/start-test/$testid"
  pad
}

stop_test() {
  curl -v "$base:30007/stop-test/$testid"
  pad
}

get_instance() {
  curl  -v "$base:30007/test-instances/$testid"
  pad
}

testid=""
pad
make_dummy_test

# strip quotes from id
testid=$(echo $testid | tr -d '"')
echo "created test with id: $testid"
pad

get_test
submit_test
get_instance
stop_test
get_instance
