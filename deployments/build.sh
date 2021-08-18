#!/bin/bash
set -euo pipefail

docker build . -t myl7/zyzzyva
#docker push myl7/zyzzyva
docker save myl7/zyzzyva | gzip > bin/myl7-zyzzyva.tar.gz
