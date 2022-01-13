package main

import (
	"bytes"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest/fake"
	cmdtesting "k8s.io/kubectl/pkg/cmd/testing"
	"k8s.io/kubectl/pkg/scheme"
)

func TestOptions_Run(t *testing.T) {
	f := newFactory(t)
	tests := []struct {
		options Options
		want    string
		wantErr string
	}{
		{
			options: Options{
				AllNamespaces: true,
				Regex:         "^secret[0-9]$",
				Output:        "json",
			},
			want: `{
    "test1": {
        "secret1": {
            "key1": "value1",
            "key2": "value2",
            "key3": "value3"
        }
    },
    "test2": {
        "secret2": {
            "key1": "value1",
            "key2": "value2",
            "key3": "value3"
        }
    }
}
`,
		},
		{
			options: Options{
				Namespace:  "test1",
				SecretName: "secret1",
				Output:     "json",
			},
			want: `{
    "test1": {
        "secret1": {
            "key1": "value1",
            "key2": "value2",
            "key3": "value3"
        }
    }
}
`,
		},
		{
			options: Options{
				Namespace:  "test1",
				SecretName: "secret1",
			},
			want: `test1:
  secret1:
    key1: value1
    key2: value2
    key3: value3

`,
		},
		{
			options: Options{
				Namespace:  "aaaaaaaaaaaaaaa",
				SecretName: "aaaaaaaaaaaaaaa",
			},
			want:    "",
			wantErr: "*v1.Pod is unexpected type",
		},
		{
			options: Options{
				MultiNamespaces: []string{"1000", "any", "test1", "kube-system"},
				Regex:           "secret",
			},
			want: `test1:
  secret1:
    key1: value1
    key2: value2
    key3: value3

`,
		},
		{
			options: Options{
				AllNamespaces: true,
				Regex:         "^notfound$",
			},
			want:    "",
			wantErr: "No secrets found",
		},
		{
			options: Options{
				AllNamespaces: true,
				Regex:         "secret",
			},
			want: `test1:
  secret1:
    key1: value1
    key2: value2
    key3: value3
test2:
  secret2:
    key1: value1
    key2: value2
    key3: value3

`,
		},
		{
			options: Options{
				AllNamespaces: true,
			},
			want: `international:
  greeding:
    english: hello
    japanese: konnichiwa
    spanish: hola
kube-system:
  foo:
    banana: value2
    dangerous-0138033: value2
    dictionary: value1
  hoge:
    private-value: value2
    somthing-secret: value1
  konnectivity-agent-token:
    ca.crt: value1
    namespace: value3
    token: value2
test1:
  secret1:
    key1: value1
    key2: value2
    key3: value3
test2:
  nodata: null
  privatevalue2:
    key1: value1
  secret2:
    key1: value1
    key2: value2
    key3: value3

`,
		},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			var out bytes.Buffer
			o := NewOptions(genericclioptions.IOStreams{Out: &out})
			setOptionValues(t, o, tt.options)
			err := o.Run(f)
			if err == nil {
				assert.Equal(t, "", tt.wantErr)
			} else {
				assert.EqualError(t, err, tt.wantErr)
			}
			assert.Equal(t, tt.want, out.String())
		})
	}
}

func setOptionValues(t *testing.T, o *Options, in Options) {
	t.Helper()

	o.Namespace = in.Namespace
	o.AllNamespaces = in.AllNamespaces
	if in.Output != "" {
		o.Output = in.Output
	}
	o.LabelSelector = in.LabelSelector
	if in.Regex != "" {
		o.Regex = in.Regex
	}
	var err error
	o.CompliedRegex, err = regexp.Compile(in.Regex)
	if err != nil {
		t.Fatalf("regex compile %s: %v", in.Regex, err)
	}
	o.SecretName = in.SecretName
	o.MultiNamespaces = in.MultiNamespaces
}

func newFactory(t *testing.T) *cmdtesting.TestFactory {
	t.Helper()

	codec := scheme.Codecs.LegacyCodec(scheme.Scheme.PrioritizedVersionsAllGroups()...)
	factory := cmdtesting.NewTestFactory()
	factory.Client = &fake.RESTClient{
		NegotiatedSerializer: scheme.Codecs.WithoutConversion(),
		GroupVersion:         v1.SchemeGroupVersion,
		Client: fake.CreateHTTPClient(func(req *http.Request) (*http.Response, error) {
			obj := newRuntimeObject(t)
			if strings.EqualFold(req.URL.Path, "/secrets") {
				return &http.Response{StatusCode: http.StatusOK, Header: cmdtesting.DefaultHeader(), Body: cmdtesting.ObjBody(codec, obj)}, nil
			}
			if ss := strings.Split(req.URL.Path, "/"); len(ss) == 5 {
				ns, name := ss[2], ss[4]
				for _, item := range obj.(*v1.List).Items {
					u := item.Object.(*unstructured.Unstructured)
					if !strings.EqualFold(ns, u.GetNamespace()) {
						continue
					}
					if strings.EqualFold(name, u.GetName()) {
						return &http.Response{StatusCode: http.StatusOK, Header: cmdtesting.DefaultHeader(), Body: cmdtesting.ObjBody(codec, item.Object)}, nil
					}
				}
				return &http.Response{StatusCode: http.StatusOK, Header: cmdtesting.DefaultHeader(), Body: cmdtesting.ObjBody(codec, &v1.Pod{})}, nil
			}
			return nil, fmt.Errorf("request url: %#v, and request: %#v", req.URL, req)
		}),
	}
	factory.ClientConfigVal = cmdtesting.DefaultClientConfig()
	return factory
}

func newRuntimeObject(t *testing.T) runtime.Object {
	t.Helper()

	return &v1.List{
		TypeMeta: metav1.TypeMeta{
			Kind:       "",
			APIVersion: "",
		},
		Items: []runtime.RawExtension{
			{
				Object: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"metadata": map[string]interface{}{
							"namespace": "kube-system",
							"name":      "konnectivity-agent-token",
						},
						"kind":       "Secret",
						"apiVersion": "v1",
						"data": map[string]interface{}{
							"ca.crt":    []byte("value1"),
							"token":     []byte("value2"),
							"namespace": []byte("value3"),
						},
					},
				},
			},
			{
				Object: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"metadata": map[string]interface{}{
							"namespace": "kube-system",
							"name":      "hoge",
						},
						"kind":       "Secret",
						"apiVersion": "v1",
						"data": map[string]interface{}{
							"somthing-secret": []byte("value1"),
							"private-value":   []byte("value2"),
						},
					},
				},
			},
			{
				Object: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"metadata": map[string]interface{}{
							"namespace": "kube-system",
							"name":      "foo",
						},
						"kind":       "Secret",
						"apiVersion": "v1",
						"data": map[string]interface{}{
							"dictionary":        []byte("value1"),
							"banana":            []byte("value2"),
							"dangerous-0138033": []byte("value2"),
						},
					},
				},
			},
			{
				Object: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"metadata": map[string]interface{}{
							"namespace": "test1",
							"name":      "secret1",
						},
						"kind":       "Secret",
						"apiVersion": "v1",
						"data": map[string]interface{}{
							"key1": []byte("value1"),
							"key2": []byte("value2"),
							"key3": []byte("value3"),
						},
					},
				},
			},
			{
				Object: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"metadata": map[string]interface{}{
							"namespace": "test2",
							"name":      "secret2",
						},
						"kind":       "Secret",
						"apiVersion": "v1",
						"data": map[string]interface{}{
							"key1": []byte("value1"),
							"key2": []byte("value2"),
							"key3": []byte("value3"),
						},
					},
				},
			},
			{
				Object: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"metadata": map[string]interface{}{
							"namespace": "test2",
							"name":      "privatevalue2",
						},
						"kind":       "Secret",
						"apiVersion": "v1",
						"data": map[string]interface{}{
							"key1": []byte("value1"),
						},
					},
				},
			},
			{
				Object: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"metadata": map[string]interface{}{
							"namespace": "test2",
							"name":      "nodata",
						},
						"kind":       "Secret",
						"apiVersion": "v1",
						"data":       nil,
					},
				},
			},
			{
				Object: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"metadata": map[string]interface{}{
							"namespace": "international",
							"name":      "greeding",
						},
						"kind":       "Secret",
						"apiVersion": "v1",
						"data": map[string]interface{}{
							"english":  []byte("hello"),
							"japanese": []byte("konnichiwa"),
							"spanish":  []byte("hola"),
						},
					},
				},
			},
		},
	}
}
