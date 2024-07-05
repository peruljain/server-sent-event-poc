# Problem Statement 
1. Developed a POC for sending server-sent events to client
2. Developed a logs reader 

# Approach

## Approach for poc
1. Create go server which will send timestamp to client in every 2 seconds
2. Create simple html which will show timestamp in browser

## Approach for logs reader
1. Create a function which will generate random logs and call this function in async way which will generate logs 
   in every 5 seconds
2. Create a function which will tail the logs and send to the client via SSE (server sent event)

![alt text](image.png)

# Steps to run

1. go run main.go
2. Open client.html in browser
3. For poc : http://localhost:8080/events
4. For logs : http://localhost:8080/logs

