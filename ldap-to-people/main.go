package main

import (
	"context"

	"github.com/benthosdev/benthos/v4/public/service"

	_ "github.com/benthosdev/benthos/v4/public/components/io"
	_ "github.com/benthosdev/benthos/v4/public/components/pure"
	_ "github.com/ugent-library/benthos-pipelines/ldap-to-people/benthosldap"
)

func main() {
	service.RunCLI(context.Background())
}
