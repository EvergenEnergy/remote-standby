#! /bin/sh
# requires https://github.com/kward/shunit2/blob/master/doc/RELEASE_NOTES-2.1.8.md

SERIAL=test-end-to-end

testWritesOptimisationPlan() {
    planPath=/tmp/backup/plan.json
    setSeconds=$(date +%s)
    docker exec mosquitto mosquitto_pub -h mosquitto -p 1883 -t cmd/local/standby/${SERIAL}/plan -m '{"site_id":"integration-test-site","optimisation_timestamp":{"seconds":'$setSeconds',"nanos":0},"optimisation_intervals":[],"setpoint_type":1}'
    sleep 3
    planSeconds=$(cat $planPath | jq .optimisation_timestamp.seconds)
    assertEquals "didn't find expected data in written plan" "$planSeconds" "$setSeconds"
}

. shunit2/shunit2
