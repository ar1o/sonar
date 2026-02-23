#!/usr/bin/env bash
# Section D: SONAR_PATH Override

test_d_path_override() {
  printf "Section D: SONAR_PATH Override"
  mkdir -p /tmp/sonar-qa-alt

  run_env /tmp/sonar-qa-alt init
  assert_exit "D" "D1" 0

  run_env /tmp/sonar-qa-alt config --json
  assert_exit "D" "D2" 0
  assert_stdout_contains "D" "D2" "/tmp/sonar-qa-alt"

  rm -rf /tmp/sonar-qa-alt

  if [ ! -d /tmp/sonar-qa-alt ]; then
    check "D" "D3" "PASS"
  else
    check "D" "D3" "FAIL" "directory /tmp/sonar-qa-alt still exists after cleanup"
  fi
}
