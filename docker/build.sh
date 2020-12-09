#!/bin/bash

docker build . -f docker/Dockerfile -t mlfo:latest
docker tag mlfo:latest abh15/mlfo:latest
docker push abh15/mlfo:latest