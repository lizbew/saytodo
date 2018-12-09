#!/usr/bin/env bash

export GIN_MODE=release
nohup ./saytodo > output.log 2>&1 &
