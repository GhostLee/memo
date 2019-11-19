module github.com/GhostLee/memo/leosys

go 1.13

require (
	github.com/GhostLee/deno v0.0.0-00010101000000-000000000000
	github.com/gorilla/websocket v1.4.1 // indirect
	github.com/satori/go.uuid v1.2.0
	github.com/tencentyun/scf-go-lib v0.0.0-20190817080819-4a2819cda320
	google.golang.org/grpc v1.25.1
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
)

replace (
	github.com/GhostLee/deno => /home/abellee/go/src/github.com/GhostLee/deno
	github.com/GhostLee/memo/bot => /home/abellee/go/src/github.com/GhostLee/memo/bot
)
