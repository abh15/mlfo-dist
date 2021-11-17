#!/bin/bash
docker build . -f docker/Dockerfile -t mlfo:meter
docker tag mlfo:meter abh15/mlfo:meter
docker push abh15/mlfo:meter