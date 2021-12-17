package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"
)

func run(ctx context.Context, name string, opt options) error {
	ss, err := findSecrets(ctx, name, opt)
	if err != nil {
		return err
	}
	if opt.Regex == "" {
		return printSecrets(ss, opt.Output)
	}
	secrets := make([]v1.Secret, 0, len(ss))
	for _, s := range ss {
		match, err := regexp.MatchString(opt.Regex, s.Name)
		if err != nil {
			return err
		}
		if match {
			secrets = append(secrets, s)
		}
	}
	return printSecrets(secrets, opt.Output)
}

func findSecrets(ctx context.Context, name string, opt options) ([]v1.Secret, error) {
	cmds, err := secretCommands(ctx, name, opt)
	if err != nil {
		return nil, err
	}
	if len(cmds) == 1 && name != "" {
		b, err := cmds[0].CombinedOutput()
		if err != nil {
			return nil, fmt.Errorf("failed to get %s in %s: %+v", name, opt.Namespace, err)
		}
		var s v1.Secret
		if err := unmarshalSecret(b, &s, opt.Output); err != nil {
			return nil, err
		}
		return []v1.Secret{s}, nil
	}
	var ss []v1.Secret
	for _, cmd := range cmds {
		b, err := cmd.CombinedOutput()
		if err != nil {
			return nil, fmt.Errorf("failed to get a secret: %+v", err)
		}
		var sl v1.SecretList
		if err := unmarshalSecretList(b, &sl, opt.Output); err != nil {
			return nil, err
		}
		ss = append(ss, sl.Items...)
	}
	return ss, nil
}

func unmarshalSecret(b []byte, s *v1.Secret, format string) error {
	switch format {
	case "json":
		if err := json.Unmarshal(b, &s); err != nil {
			return fmt.Errorf("failed to json unmarshal secret: %+v", err)
		}
	case "yaml":
		if err := yaml.Unmarshal(b, &s); err != nil {
			return fmt.Errorf("failed to yaml unmarshal secret: %+v", err)
		}
	}
	return nil
}

func unmarshalSecretList(b []byte, sl *v1.SecretList, format string) error {
	switch format {
	case "json":
		if err := json.Unmarshal(b, &sl); err != nil {
			return fmt.Errorf("failed to json unmarshal secret: %+v", err)
		}
	case "yaml":
		if err := yaml.Unmarshal(b, &sl); err != nil {
			return fmt.Errorf("failed to yaml unmarshal secret: %+v", err)
		}
	}
	return nil
}

func secretCommands(ctx context.Context, name string, opt options) ([]*exec.Cmd, error) {
	if name != "" && opt.Namespace != "" {
		return nil, errors.New("specify name or namespace")
	}
	o := opt.toKubectlOptions()
	if name != "" {
		return []*exec.Cmd{kubectlGetSecret(ctx, append([]string{name, "-n", opt.Namespace}, o...)...)}, nil
	}
	ns := []string{opt.Namespace}
	if opt.AllNamespaces {
		namespaces, err := getAllNamespaces(ctx)
		if err != nil {
			return nil, err
		}
		for _, n := range namespaces {
			ns = append(ns, n.Name)
		}
	}
	if len(opt.MultiNamespaces) != 0 {
		ns = append(ns, strings.Split(opt.MultiNamespaces, ",")...)
	}
	cmds := make([]*exec.Cmd, len(ns))
	for i, n := range ns {
		cmds[i] = kubectlGetSecret(ctx, append([]string{"-n", n}, o...)...)
	}
	return cmds, nil
}

func getAllNamespaces(ctx context.Context) ([]v1.Namespace, error) {
	b, err := kubectlGetNamespace(ctx, "-o", "json").CombinedOutput()
	if err != nil {
		return nil, err
	}
	var nl v1.NamespaceList
	if err := json.Unmarshal(b, &nl); err != nil {
		return nil, err
	}
	return nl.Items, nil
}

func printSecrets(ss []v1.Secret, output string) error {
	m := make(map[string][]interface{}, len(ss))
	for _, s := range ss {
		data := make(map[string]string)
		for k, v := range s.Data {
			data[k] = string(v)
		}
		m[s.Namespace] = append(m[s.Namespace], map[string]map[string]string{
			s.Name: data,
		})
	}
	var b []byte
	var err error
	switch output {
	case "json":
		b, err = json.MarshalIndent(m, "", "    ")
		if err != nil {
			return err
		}
	case "yaml":
		b, err = yaml.Marshal(m)
		if err != nil {
			return err
		}
	}
	fmt.Printf("%s\n", string(b))
	return nil
}
