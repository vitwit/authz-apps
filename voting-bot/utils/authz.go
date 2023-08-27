package utils

import (
	"context"
	"crypto/tls"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	authztypes "github.com/cosmos/cosmos-sdk/x/authz"
)

// HasAuthzGrant returns true if authz permission is exist for the provided
// parameters.
func HasAuthzGrant(endpoint, granter, grantee, typeURL string) (bool, error) {
	creds := credentials.NewTLS(&tls.Config{InsecureSkipVerify: false})
	conn, err := grpc.Dial(endpoint, grpc.WithTransportCredentials(creds))
	if err != nil {
		log.Printf("Failed to connect to %s: %v", endpoint, err)
		return false, err
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	client := authztypes.NewQueryClient(conn)

	grants, err := client.Grants(ctx, &authztypes.QueryGrantsRequest{
		Granter:    granter,
		Grantee:    grantee,
		MsgTypeUrl: typeURL,
	})
	if err != nil {
		return false, err
	}

	if len(grants.Grants) > 0 {
		return true, nil
	}

	return false, nil

}
