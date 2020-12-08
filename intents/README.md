Intent has two important fields
1. `distIntent` which tells if the intent is distributed or not. If `false` intent would be treated as normal local pipeline intent. If `true` it will be treated as distributed and the MLFO will try to deompose it send the result over momo to corresponding edges. Intent sent to the fed server is `distIntent=false` while intent sent to all other edge nodes is `distIntent=true`.

2. `type` which tells what type of distrbuted intent it is e.g federated, splitNN, model chained etc

3. Server is IP/ID for cloud MLFO. Maybe make more intuitive 