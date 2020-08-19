base=$(minikube ip)

pad() {
  printf "\n\n\n"
}

make_dummy_test() {
  output=$(curl "$base:30007/tests"\
        -H "Content-Type: application/json"\
        -d "{
            \"Name\": \"Test1\",
            \"Jobs\": [
              {
                \"Name\": \"alpha\",
                \"Group\": \"diago-worker\",
                \"Priority\": 0,
                \"Frequency\":  5,
			          \"Duration\":   30,
			          \"HTTPMethod\": \"GET\",
			          \"HTTPUrl\":    \"https://www.google.com\"
              }
            ]
          }")
  testid=$(echo $output | python3 -c "import sys, json; print(json.load(sys.stdin)['payload']['testid'])")
}

make_dummy_test2() {
  output=$(curl "$base:30007/tests"\
        -H "Content-Type: application/json"\
        -d "{
            \"Name\": \"Test1\",
            \"Jobs\": [
              {
                \"ID\": \"1\",
                \"Name\": \"alpha\",
                \"Group\": \"diago-worker\",
                \"Priority\": 0,
                \"Frequency\":  5,
			          \"Duration\":   5,
			          \"HTTPMethod\": \"GET\",
			          \"HTTPUrl\":    \"https://www.google.com\"
              }
            ]
          }")
  testid=$(echo $output | python3 -c "import sys, json; print(json.load(sys.stdin)['payload']['testid'])")
}

get_test() {
  curl "$base:30007/tests/$testid" | python3 -m json.tool
  pad
}

submit_test() {
  curl "$base:30007/start-test/$testid" | python3 -m json.tool
  pad
}

stop_test() {
  curl "$base:30007/stop-test/$testid" | python3 -m json.tool
  pad
}

get_instance() {
  curl "$base:30007/test-instances/$testid" | python3 -m json.tool
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

sleep 5

stop_test
get_instance