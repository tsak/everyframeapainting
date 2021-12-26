#!/bin/bash

if [ -z "$1" ] || [ -z "$2" ]; then
  echo "Usage:"
  echo "$0: <infile>.mp4 <outfile>.mp4 [-normal|-reverse] [xx%]"
  exit 1
fi

# Determine number of cores
CORES=$(nproc)
INFILE=$1
if [ ! -e "$INFILE" ]; then
  echo "$INFILE doesn't exist"
  exit 1
fi
OUTFILE=$2

# Default to 10% scale factor
PERCENT="10%"
if [ -n "$4" ]; then
  PERCENT="$4"
fi

# Timestamp for temp dir
TIMESTAMP=$(date +%s)

mkdir "$TIMESTAMP" || exit 1
ffmpeg -i "$INFILE" -vf fps=30 -vsync 0 "$TIMESTAMP/m%10d.png"
(
  cd "$TIMESTAMP" || exit 1
  ls m* | xargs -P "$CORES" -i convert {} -resize "$PERCENT" s{}

  if [ "$3" == "-normal" ]; then
    ls sm* | xargs -P "$CORES" -i ../everyframeapainting -in {} -out p{} -normal
  else
    ls sm* | xargs -P "$CORES" -i ../everyframeapainting -in {} -out p{}
  fi
)
ffmpeg -i "$TIMESTAMP/psm%10d.png" -c:v libx264 -vf fps=30 -pix_fmt yuv420p "$OUTFILE"
rm -Rf "$TIMESTAMP"