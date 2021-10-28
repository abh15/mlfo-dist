#!/bin/bash
for i in {10..59}
do
   docker update --cpus 2 mn.fc.$i
done