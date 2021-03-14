# New Relic Challenge


## Task
This server listens to TCP port 4000 at most 5 concurrent clients. connects to the application and write any number of 9 digit numbers, and then close the connection. The Application
writea a de- duplicated list of these numbers to a log file (`numbers.log`).

### Requirements
+ Go 1.15
+ Go modules
### Running from the command line
running the app: `make run`

running tests: `make test`




