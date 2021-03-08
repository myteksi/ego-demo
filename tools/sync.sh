#!/bin/bash

case $1 in
"up")
  rsync -rlptzv --progress --exclude bazel --exclude user.bazelrc --exclude /envoy --exclude .git --exclude compile_commands.json . "$DEV_BOX_USER@$DEV_BOX_IP:~/$DEV_BOX_FOLDER_NAME"
  ;;
"down")
  rsync -rlptzv --progress --delete --exclude=.git --exclude bazel --exclude user.bazelrc --exclude /envoy "$DEV_BOX_USER@$DEV_BOX_IP:~/$DEV_BOX_FOLDER_NAME/*" ./
  ;;  
*)
  echo "Unknown command. Should be 'sync up' or 'sync down'"
esac