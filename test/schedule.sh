base=$(minikube ip)

schedule_test1() {
  curl "$base:30007/api/test-schedules"\
    -H "Content-Type: application/json"\
    -d "{
        \"Name\": \"TestSchedule1\",
        \"TestID\": \"Test1\",
        \"CronSpec\": \"* * * * *\"
      }"
}

schedule_test1
