package test

import (
	"github.com/askrella/askrella-ssh-provider/internal/provider"
	"github.com/askrella/askrella-ssh-provider/internal/provider/ssh"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

var (
	testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
		"ssh": providerserver.NewProtocol6WithError(provider.New("test")()),
	}

	sshConfig = ssh.SSHConfig{
		Host:     "::1",
		Port:     2222,
		Username: "testuser",
		Password: "testpass",
	}
)
