#!/bin/bash
#

set -e

services/echo/echo & services/hmac/hmac & envoy -c envoy.yaml