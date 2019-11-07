module github.com/dimaunx/armada

go 1.12

require (
	github.com/Microsoft/go-winio v0.4.14 // indirect
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/docker v1.13.1
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/gobuffalo/packr/v2 v2.7.1
	github.com/imdario/mergo v0.3.8 // indirect
	github.com/opencontainers/go-digest v1.0.0-rc1 // indirect
	github.com/pkg/errors v0.8.1
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cobra v0.0.5
	golang.org/x/crypto v0.0.0-20191106202628-ed6320f186d4 // indirect
	golang.org/x/net v0.0.0-20191105084925-a882066a44e0 // indirect
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45 // indirect
	golang.org/x/sys v0.0.0-20191105231009-c1f44814a5cd // indirect
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0 // indirect
	golang.org/x/tools v0.0.0-20191107010934-f79515f33823 // indirect
	gopkg.in/yaml.v2 v2.2.5
	k8s.io/api v0.0.0-20191107030003-665c8a257c1a
	k8s.io/apiextensions-apiserver v0.0.0-20191107032734-f60a3abe8be9
	k8s.io/apimachinery v0.0.0-20191107025710-52af7e4fbc87
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/utils v0.0.0-20191030222137-2b95a09bc58d // indirect
	sigs.k8s.io/kind v0.5.1
)

replace (
	k8s.io/api => k8s.io/api v0.0.0-20190409021203-6e4e0e4f393b
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20190409022649-727a075fdec8
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190404173353-6a84e37a896d
	k8s.io/client-go => k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
)
