#!/bin/bash
docker build . -f docker/Dockerfile -t mlfo:mptest1
docker tag mlfo:mptest1 abh15/mlfo:mptest1
docker push abh15/mlfo:mptest1