# Tic Tac Go
This is a simple implementation of web-based Tic Tac Toe game using Go and HTMX.

## How To
```shell
go run main.go
```

Join the game at [`http://localhost:8080`](http://localhost:8080). You will be redirected to a URL with game ID in the query parameters so it can be used for other players to join.

PlayerIDs are stored in cookies so it is easy to re-join existing games and easily copy/paste invite without including your own player ID.
