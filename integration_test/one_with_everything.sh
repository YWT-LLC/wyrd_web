#!/usr/bin/env bash

set -e

## Setup ##

repos="repos/flutter"
device=""

# Gather flag variables
while [[ "$1" != "" ]]; do
  case $1 in
    --device ) device="-d $2"; shift;;
    --repo-path ) repos="$2"; shift;;
    * ) echo "Invalid input. Aborting."; exit 1
  esac
  shift
done

prefix=$HOME/$repos/wyrd_web

## Tests ##

flutter drive --driver=$prefix/test_driver/integration_test_driver.dart --target=$prefix/integration_test/default_test.dart $device

flutter drive --driver=$prefix/test_driver/integration_test_driver.dart --target=$prefix/integration_test/es_test.dart $device
flutter drive --driver=$prefix/test_driver/integration_test_driver.dart --target=$prefix/integration_test/fr_test.dart $device

flutter drive --driver=$prefix/test_driver/integration_test_driver.dart --target=$prefix/integration_test/lefty_test.dart $device

flutter drive --driver=$prefix/test_driver/integration_test_driver.dart --target=$prefix/integration_test/minimum_config_test.dart $device
flutter drive --driver=$prefix/test_driver/integration_test_driver.dart --target=$prefix/integration_test/maximum_config_test.dart $device