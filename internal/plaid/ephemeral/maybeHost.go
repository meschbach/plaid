package ephemeral

import (
	"os"
)

func ResolvePlaidSocketPath() string {
	socket, has := os.LookupEnv("PLAID_SOCKET")
	if has {
		return socket
	}

	return "/tmp/plaid.socket"
}
