package main

import (
	"context"
	"os"
	"os/exec"
)

func kubectl() string {
	if v := os.Getenv("KUBECTL_COMMAND"); v != "" {
		return v
	}
	return "kubectl"
}

func kubectlGetSecret(ctx context.Context, opt ...string) *exec.Cmd {
	return exec.CommandContext(ctx, kubectl(), append([]string{"get", "secret"}, opt...)...)
}

func kubectlGetNamespace(ctx context.Context, opt ...string) *exec.Cmd {
	return exec.CommandContext(ctx, kubectl(), append([]string{"get", "namespace"}, opt...)...)
}
