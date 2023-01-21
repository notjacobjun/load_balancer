# load_balancer
Created a load balancer from scratch using Go. Currently using the Round Robin technique to cycle through available backends

### Run the application
Starting from the root directory of the project run this command 'go go run ./cmd/load_balancer/main.go --backends=http://localhost:3031,http://localhost:3032,http://localhost:3033,http://localhost:3034'

- Note that you can replace the list of backends for other backends url's if you would like
