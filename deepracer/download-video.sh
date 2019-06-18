#!/bin/bash -e
export AWS_DEFAULT_REGION=us-east-1
# export AWS_PROFILE=optional_profile

set -x

aws kinesisvideo list-streams

STREAM_ARN=$(aws kinesisvideo list-streams --query 'StreamInfoList[0].StreamARN' --output text)
if [[ "$STREAM_ARN" == "None" ]]; then
  exit
fi

ENDPOINT_URL=$(aws kinesisvideo get-data-endpoint --api-name GET_MEDIA --stream-arn "$STREAM_ARN" --query DataEndpoint --output text)
FN="deepracer-$(date +%F-%H-%M-%S)"
aws kinesis-video-media get-media --endpoint-url "$ENDPOINT_URL" --stream-arn "$STREAM_ARN" --start-selector StartSelectorType=NOW "$FN.mkv"

if [[ "$(du -k "$FN" | cut -f1)" == "0" ]]; then
  # This is what happens if the video stream hasn't really started yet.. simply re-run the script.
  echo "Empty file produced."
else
  # Re-package the video file to make it seekable
  ffmpeg -i "$FN.mkv" -vcodec copy "$FN-copy.mkv"
fi
rm "$FN.mkv"
