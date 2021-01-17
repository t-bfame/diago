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
                \"Group\": \"test-worker\",
                \"Priority\": 0,
                \"Frequency\":  5,
			          \"Duration\":   30,
			          \"HTTPMethod\": \"GET\",
			          \"HTTPUrl\":    \"https://www.google.com\"
              },
              {
                \"Name\": \"beta\",
                \"Group\": \"test-worker\",
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
            \"Name\": \"Test2\",
            \"Jobs\": [
              {
                \"Name\": \"alpha\",
                \"Group\": \"test-worker\",
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

kubectl apply -f test/test.yaml

make_dummy_test
make_dummy_test2
