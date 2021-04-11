base=$(minikube ip)

pad() {
  printf "\n\n\n"
}

make_dummy_test() {
  output=$(curl "$base:30007/api/tests"\
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
  output=$(curl "$base:30007/api/tests"\
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

make_dummy_test3() {
  output=$(curl "$base:30007/api/tests"\
        -H "Content-Type: application/json"\
        -d "{
            \"Name\": \"Test5\",
            \"Jobs\": [
              {
                \"Name\": \"alpha\",
                \"Group\": \"tests-worker\",
                \"Priority\": 0,
                \"Frequency\":  50,
			          \"Duration\":   60,
			          \"HTTPMethod\": \"GET\",
			          \"HTTPUrl\":    \"http://dummy-service.default.svc.cluster.local:8080\"
              }
            ],
            \"Chaos\": [
              {
                \"Namespace\": \"default\",
                \"Selectors\": {
                  \"app\": \"dummy\"
                },
                \"Timeout\":  10,
			          \"Count\":   1
              }
            ]
          }")
  testid=$(echo $output | python3 -c "import sys, json; print(json.load(sys.stdin)['payload']['testid'])")
}


make_dummy_test3
