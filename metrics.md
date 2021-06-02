container_cpu_cfs_throttled_seconds_total gives the time for which cpu was throttled in ns

sum(rate(container_cpu_usage_seconds_total{name=~"mn.fog.1"}[10s]))*100
sudo docker update --cpus 2 mn.fog.1


The following gives number of periods for which the cpu was throttled because the limits(1 cpu) were exceeded. Each period is equal to 100,000 microseconds:

( max_over_time (container_cpu_cfs_throttled_periods_total{name=~"mn.fog.1"}[1m]) - 
min_over_time (container_cpu_cfs_throttled_periods_total{name=~"mn.fog.1"}[1m])) 