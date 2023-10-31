## COM-Chitty-Chat
1. Open the program in an IDE
2. Open atleast two terminals (one for the server and one for the client)
3. From the COM-CHITTY-CHAT folder run the following command in one of the terminals: go run server/server.go -port 5454 (this is now your server)
4. In the other terminal run the following command: go run client/client.go -sPort 5454 (this is now your client)
5. You can now type messages to yourself. If you want more clients just open more terminals and type the client command.
6. A log of every RPC method call can be seen in the server script.
7. Enjoy our Chitty Chat application