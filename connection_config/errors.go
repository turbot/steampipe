package connection_config

import (
	"errors"
	"log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func HandleGrpcError(err error, connection, call string) error {
	// if this is a not implemented error we silently swallow it
	status, ok := status.FromError(err)
	if !ok {
		return err
	}

	// ignore unimplemented error
	if status.Code() == codes.Unimplemented {
		log.Printf("[INFO] connection '%s' returned 'Unimplemented' error for call '%s' - plugin version does not support this call", connection, call)
		return nil
	}

	return errors.New(status.Message())
}
