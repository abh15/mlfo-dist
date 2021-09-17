# Installation of cadvisor + prometheus + grafana
1. Install cadvisor (port 7070)

`sudo docker run 
  --volume=/:/rootfs:ro 
  --volume=/var/run:/var/run:ro 
  --volume=/sys:/sys:ro
  --volume=/var/lib/docker/:/var/lib/docker:ro
  --volume=/dev/disk/:/dev/disk:ro 
  --publish=7070:8080 
  --detach=true
  --name=cadvisor
  --privileged
  --device=/dev/kmsg 
  gcr.io/google-containers/cadvisor:latest `


2. Copy the config/prometheus.yml file to correct path (/home/abhishek/). Install prometheus (port 9090)

`sudo docker run
    -p 9090:9090
    -v /home/abhishek/prometheus.yml:/etc/prometheus/prometheus.yml
    prom/prometheus&
`

3. Add Grafana.

`docker run -d -p 3000:3000 grafana/grafana`

4. Go to configuration > add data source > add prometheus. Add the following HTTP URL and click save and test.

`http://10.66.2.142:9090`

5. Add the following dashboard by downloading and importing JSON file.

`https://grafana.com/grafana/dashboards/193`

6. Restrict the number of CPUs on cloud MLFO to 2.

`sudo docker update --cpus 2 mn.cloud.0`

7. After importing dashboard, go to CPU usage > explore and add the following query and click run query.

`( max_over_time (container_cpu_cfs_throttled_periods_total{name=~"mn.cloud.0"}[1m]) - 
min_over_time (container_cpu_cfs_throttled_periods_total{name=~"mn.cloud.0"}[1m])) `


# CPU metrics explained
sum(rate(container_cpu_usage_seconds_total{name=~"mn.cloud.0"}[10s]))*100

container_cpu_cfs_throttled_seconds_total gives the time for which cpu was throttled in ns
The following gives number of periods for which the cpu was throttled because the limits(1 cpu) were exceeded. Each period is equal to 100,000 microseconds:

( max_over_time (container_cpu_cfs_throttled_periods_total{name=~"mn.cloud.0"}[1m]) - 
min_over_time (container_cpu_cfs_throttled_periods_total{name=~"mn.cloud.0"}[1m])) 