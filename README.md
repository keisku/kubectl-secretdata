# kubectl-secret-data

## What is it?

This is a `kubectl` plugin for finding decoded secret data.
Since `kubectl` only outputs base64-encoded secrets, it makes it difficult to check the secret value.
This plugin helps finding a decoded secret data you want by variety of options.

## Usage

```
A kubectl plugin for finding decoded secret data.

Usage:
  kubectl-secret-data [flags]

Flags:
  -A, --all-namespace       If present, find secrets from all namespaces
      --cluster string      The name of the kubeconfig context to use
      --context string      The name of the kubeconfig cluster to use
  -h, --help                help for kubectl-secret-data
      --kubeconfig string   Path to the kubeconfig file to use for CLI requests
  -n, --namespace string    The namespaces where secrets exist. You can set multiple namespaces separated by ","
  -o, --output string       The format of the result (default "yaml")
  -E, --regex string        The regular expression of secret name
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
kubectl-secret-data -n ns-1,ns-2 -o json
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

List all secret data in `ns-1` in `json` with a regex.

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
