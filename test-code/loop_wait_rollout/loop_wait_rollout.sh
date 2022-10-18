#!/bin/bash -e

wait_rollout() {
  namespace=$1
  resource_type=$2
  name=$3

  kubectl rollout status --watch=true -n $namespace $resource_type $name
}


loop_wait_rollout_logging_audit() {
  local NS=cattle-logging-system

  for i in $(seq 1 $1)
  do
    local EXIT_CODE=0 # reset each loop
    sleep 2

    # logging operator
    wait_rollout $NS deployment rancher-logging || EXIT_CODE=$?
    if [ $EXIT_CODE != 0 ]; then
      echo "continue waiting rollout deployment rancher-logging, $i"
      continue
    fi

    # agent to grab log
    wait_rollout $NS daemonset rancher-logging-root-fluentbit || EXIT_CODE=$?
    if [ $EXIT_CODE != 0 ]; then
      echo "continue waiting rollout daemonset rancher-logging-root-fluentbit, $i"
      continue
    fi

    wait_rollout $NS daemonset rancher-logging-rke2-journald-aggregator || EXIT_CODE=$?
    if [ $EXIT_CODE != 0 ]; then
      echo "continue waiting rollout daemonset rancher-logging-rke2-journald-aggregator, $i"
      continue
    fi

    wait_rollout $NS daemonset rancher-logging-kube-audit-fluentbit || EXIT_CODE=$?
    if [ $EXIT_CODE != 0 ]; then
      echo "continue waiting rollout daemonset rancher-logging-kube-audit-fluentbit, $i"
      continue
    fi

    # fluentd, a known issue: https://github.com/harvester/harvester/issues/2787
    # wait_rollout cattle-logging-system statefulset rancher-logging-root-fluentd
    # wait_rollout cattle-logging-system statefulset rancher-logging-kube-audit-fluentd

    break
  done

  if [ $EXIT_CODE != 0 ]; then
    echo "fail to wait rollout logging audit"
    return $EXIT_CODE
  fi

  echo "success to wait rollout logging audit"
  return 0
}

loop_wait_rollout_event() {
  local NS=cattle-logging-system
  local NAME=harvester-default-event-tailer

  for i in $(seq 1 $1)
  do
    local EXIT_CODE=0 # reset each loop
    sleep 2

    wait_rollout $NS statefulset $NAME || EXIT_CODE=$?
    if [ $EXIT_CODE != 0 ]; then
      echo "continue waiting rollout statefulset $NAME, $i"
      continue
    fi

    break
  done

  if [ $EXIT_CODE != 0 ]; then
    echo "fail to wait rollout event"
    return $EXIT_CODE
  fi

  echo "success to wait rollout event"
  return 0
}


loop_wait_rollout_logging_audit 20

loop_wait_rollout_event 5
