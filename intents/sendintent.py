#!/usr/bin/python3
import requests
import sys
import time

"""" 
IP range 10.0.1.10 to 10.0.1.59
Run python3 sendintentpy. 
Pls check length of all arrays.
If sameserverdist is ues then that cohort shared fedavg server with the cohort before that one
"""

#timegap in seconds is period in between two intents 
timegap=2
cohortdistr = [10, 10]
intentdist=['intent.yaml', '2intent.yaml']#, 'intent.yaml', 'intent.yaml', 'intent.yaml', 'intent.yaml']
sameserverdist= ['no','no']
avgalgodist=['FedAvg', 'FedAvg']
fracfitdist= ['0.3','0.3']
minfitdist=['1','1']
minavdist= ['1','1']
numrounddist =['10','10']


def send(intentfile, ipstart, cohortsize, sameserver, avgalgo, fracfit, minfit, minav, numround):
    files = {'file': open(intentfile, 'rb')}
    payload = {'ipstart': ipstart, 'cohortsize' : cohortsize, 'sameserver': sameserver, 'avgalgo': avgalgo,
                'fracfit':fracfit, 'minfit':minfit, 'minav':minav, 'numround':numround}
    r = requests.post('http://10.66.2.142:8001/receive', files=files, data=payload)
       
def main():
    octtraker = 10
    for i in range(len(cohortdistr)):
        cohort= cohortdistr[i]
        intent= intentdist[i]
        sameserver= sameserverdist[i]
        avgalgo=avgalgodist[i]
        fracfit= fracfitdist[i]
        minfit= minfitdist[i]
        minav= minavdist[i]
        numround= numrounddist[i]
        send(intent, octtraker, cohort, sameserver, avgalgo, fracfit, minfit, minav, numround)
        octtraker=octtraker+cohort
        time.sleep(timegap)


main()
