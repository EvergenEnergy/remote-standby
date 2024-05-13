#! /bin/sh
# requires https://github.com/kward/shunit2/blob/master/doc/RELEASE_NOTES-2.1.8.md

SERIAL=test-end-to-end

testWritesOptimisationPlan() {
    planPath=/tmp/written-plan.json
    docker exec command_standby_local rm $planPath
    publishPlan=$(docker exec mosquitto mosquitto_pub -h mosquitto -p 1883 -t cmd/local/standby/${SERIAL}/plan -m '{"site_id":"integration-test-site","optimisation_timestamp":{"seconds":1715566795,"nanos":0},"optimisation_intervals":[],"setpoint_type":1}')
    sleep 3
    writtenPlan=$(docker exec command_standby_local cat $planPath)
    planSeconds=$(echo $writtenPlan | jq .optimisation_timestamp.seconds)
    assertEquals "didn't find expected data in written plan" "$planSeconds" "1715566795"
    docker exec command_standby_local rm $planPath
}

. shunit2/shunit2
