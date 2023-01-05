# libp2p-message

## Steps to run the project
1. Run `go build -o message main.go`
2. Run these commands in 3 separate terminals:
```
./message -i 0 -p 3021
```
```
 ./message -i 1 -p 3022
```
```
./message -i 2 -p 3023
```
3. You should see the nodes connecting to each other and sending messages.