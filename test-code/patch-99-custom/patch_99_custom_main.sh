#!/bin/bash -ex
do_it () {
  source ./patch_99_custom.sh

  # keep the 99_custom_v103.yaml untouched
  cp -f ./99_custom_v103.yaml ./99_custom_v103_11.yaml

  patch_99_custom ./99_custom_v103_11.yaml ./99_custom_v103_11_tmp.yaml 1
}

do_it
