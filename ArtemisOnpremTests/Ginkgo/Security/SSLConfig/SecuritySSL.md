# Starting situation
- Artemis setup contains 3 brokers. 
- That port 61617 is opened with SSL.

# Happy cases
- The Artemis setup uses SSL for its communication. [1]
- SSL is setup correctly via the bundle, i.e. every namespace has a bundle present. [2]

# Fault cases
- Communication in the Artemis setup uses SSL but can be accessed without SSL. [3]
- It is not possible to connect to the Artemis setup with a producer/consumer which has the correct configmap (bundle) mounted. []

# Test cases
|#|Test case|Desired outcome|Actual outcome|
|---|---|---|---|
| [case_1](case1_test.go) | Setup a producer that sends a message to a queue. The producer should have the bundle mounted. Check to see whether message is there and can be retrieved. | The message is produced on the queue and can be retrieved. ||
| [case_2](case2_test.go) | Check every namespace in the cluster to see if the bundle configmap is present. | Every namespace in the cluster has the bundle. ||
| [case_3](case3_test.go) | Setup a producer that sends a message to a queue. The producer should not have the bundle mounted. Check to see whether message is there and can be retrieved. | The message is not produced inside the queue and cannot be retrieved. ||
| [case_4](case4_test.go) | Setup a producer that sends a message to a queue. The producer should have the bundle mounted. Check to see whether message is there and can be retrieved. | The message is not produced inside the queue and cannot be retrieved. |???|

*Case 3 and 4 will be put in another folder called SSL because they do not need to have the configmap mounted.*

# Documentation review
| # | Test case | Desired outcome |
| --- | --- | --- | 
| # | Review documentation in ADO WIKI. | Confirm that the documentation accurately reflects the behavior of automatic queue creation, including any configurable parameters and troubleshooting steps. | 