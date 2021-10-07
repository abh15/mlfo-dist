#!/bin/bash
prometheus_ss_exporter -p 9200 -c /app/ss_exporter_src/cnfg.yml&
./app/node_exporter/node_exporter&
/app/mlfo



