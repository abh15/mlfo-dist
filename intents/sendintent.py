#!/usr/bin/python3
import requests
import sys
import time

# IP range 10.0.1.10 to 10.0.1.59

timegap=5
cohortdistr = [10]
intentdist=['intent.yaml']#, 'intent.yaml', 'intent.yaml', 'intent.yaml', 'intent.yaml']

def send(intentfile, ipstart, cohortsize):
    files = {'file': open(intentfile, 'rb')}
    payload = {'ipstart': ipstart, "cohortsize" : cohortsize}
    r = requests.post('http://10.66.2.142:8001/receive', files=files, data=payload)
       

def main():
    octtraker = 10
    if len(cohortdistr)!= len(intentdist):
        print("Unequal distributions. Please check input")
        sys.exit()
    for i in range(len(cohortdistr)):
        cohort= cohortdistr[i]
        intent= intentdist[i]
        send(intent, octtraker, cohort)
        octtraker=octtraker+cohort
        time.sleep(timegap)


main()
