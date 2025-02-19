package test

import (
	"github.com/askrella/askrella-ssh-provider/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

var (
	testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
		"askrella-ssh": providerserver.NewProtocol6WithError(provider.New("test")()),
	}
)
