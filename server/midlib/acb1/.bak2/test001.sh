#!/bin/bash

curl -X POST "http://127.0.0.1:9019/api/acb1/track_add?bulk=$( one-line ./data001.json | query-escape )"
