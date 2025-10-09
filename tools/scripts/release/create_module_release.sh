DEV_VER=$1
EXPERIMENTAL_VER=$2
FAST_VER=$3
REGULAR_VER=$4

cat <<EOF
targetLandscapes:
  - dev
  - stage
  - prod
channels:
  - channel: dev
    version: $DEV_VER
  - channel: experimental
    version: $EXPERIMENTAL_VER
  - channel: fast
    version: $FAST_VER
  - channel: regular
    version: $REGULAR_VER
EOF
