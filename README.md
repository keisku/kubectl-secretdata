# kubectl-secret-data

## What is it?

This is a `kubectl` plugin for finding decoded secret data.
Since `kubectl` outputs base64-encoded secrets basically, it makes it difficult to check the secret value. And searching secrets also is difficult.
This CLI helps checking secret value and finding a decoded secret data you want with productive search flags.

## Usage

```
A kubectl plugin for finding decoded secret data.

Usage:
  kubectl-secret-data [flags]

Flags:
  -A, --all-namespaces            If present, find secrets from all namespaces
      --cluster string            The name of the kubeconfig context to use
      --context string            The name of the kubeconfig cluster to use
  -h, --help                      help for kubectl-secret-data
      --kubeconfig string         Path to the kubeconfig file to use for CLI requests
  -m, --multi-namespaces string   The multi namespacess separated by "," where secrets exist.
  -n, --namespace string          The namespaces where secrets exist
  -o, --output string             The format of the result (default "yaml")
  -E, --regex string              The regular expression of secret name
```

### Example

List all secret data in `ns-1` in `yaml`(default).

```shell
kubectl-secret-data -n ns-1
# OR
kubectl-secret-data -n ns-1 -o yaml
```

<details>
<summary>Output</summary>

```yaml
ns-1: # Namespace
  - private-data-a: # Secrete Name
      password: lkiugubau # Secret Data Key
      user: smith
  - private-data-b:
      password: hiahgeoawngleawngaw
      user: bob
  - super-private-data-a:
      password: hoge
      user: foo
  - super-private-data-b:
      password: fuga
      user: bar
```

</details>

List all secret data in `ns-1` in `json`.

```shell
kubectl-secret-data -n ns-1 -o json
```

<details>
<summary>Output</summary>

```json
{
  "ns-1": [
    {
      "private-data-a": {
        "password": "lkiugubau",
        "user": "smith"
      }
    },
    {
      "private-data-b": {
        "password": "hiahgeoawngleawngaw",
        "user": "bob"
      }
    },
    {
      "super-private-data-a": {
        "password": "hoge",
        "user": "foo"
      }
    },
    {
      "super-private-data-b": {
        "password": "fuga",
        "user": "bar"
      }
    }
  ]
}
```

</details>

List all secret data in `ns-1` and `ns-2` in `json`.
**You can specify multiple namespace.**

```shell
kubectl-secret-data -m ns-1,ns-2 -o json
#OR
kubectl-secret-data --multi-namespaces ns-1,ns-2 -o json
```

<details>
<summary>Output</summary>

```json
{
  "ns-1": [
    {
      "private-data-a": {
        "password": "lkiugubau",
        "user": "smith"
      }
    },
    {
      "private-data-b": {
        "password": "hiahgeoawngleawngaw",
        "user": "bob"
      }
    },
    {
      "super-private-data-a": {
        "password": "hoge",
        "user": "foo"
      }
    },
    {
      "super-private-data-b": {
        "password": "fuga",
        "user": "bar"
      }
    }
  ],
  "ns-2": [
    {
      "important-value-x": {
        "password": "abcd",
        "user": "sam"
      }
    },
    {
      "important-value-y": {
        "password": "xyz",
        "user": "alice"
      }
    }
  ]
}
```

</details>

List secret data by matching regex in `ns-1` in `json`.

```shell
kubectl-secret-data -n ns-1 -E "^super-.*"
```

<details>
<summary>Output</summary>

```json
{
  "ns-1": [
    {
      "super-private-data-a": {
        "password": "hoge",
        "user": "foo"
      }
    },
    {
      "super-private-data-b": {
        "password": "fuga",
        "user": "bar"
      }
    }
  ]
}
```

</details>

## Installing

### Pre-built binaries

See the [release](https://github.com/kei6u/kubectl-secret-data/releases) page for the full list of pre-built assets.

#### Linux

```bash
curl -L -o kubectl-secret-data.tar.gz https://github.com/kei6u/kubectl-secret-data/releases/download/v0.3.2/kubectl-secret-data_0.3.2_Linux_arm64.tar.gz
tar -xvf kubectl-secret-data.tar.gz
mv kubectl-secret-data /usr/local/bin/kubectl-secret-data
```

#### OSX

```bash
curl -L -o kubectl-secret-data.tar.gz https://github.com/kei6u/kubectl-secret-data/releases/download/v0.3.2/kubectl-secret-data_0.3.2_Darwin_arm64.tar.gz
tar -xvf kubectl-secret-data.tar.gz
mv kubectl-secret-data /usr/local/bin/kubectl-secret-data
```

### Source

```
go get github.com/kei6u/kubectl-secret-data
```
