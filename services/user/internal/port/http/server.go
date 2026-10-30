// Package http maps protocols to the service application handler aggregates.
package http

import (
	"github.com/nostosx/servex-contrib/transport/httpx"
	serviceapp "github.com/nostosx/servex-examples/services/user/internal/application"
)

type UserServer struct {
	commands *serviceapp.Command
	queries  *serviceapp.Query
}

func NewUserServer(commands *serviceapp.Command, queries *serviceapp.Query) *UserServer {
	return &UserServer{commands: commands, queries: queries}
}
func Register(server *httpx.Server, handler *UserServer) { RegisterGetUser(server, handler) }
