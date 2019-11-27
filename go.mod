module github.com/dimaunx/armada

go 1.12

require (
	github.com/Masterminds/semver v1.5.0
	github.com/Microsoft/go-winio v0.4.14 // indirect
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/docker v1.13.1
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/gernest/wow v0.1.0 // indirect
	github.com/gobuffalo/packr/v2 v2.7.1
	github.com/google/pprof v0.0.0-20191105193234-27840fff0d09 // indirect
	github.com/ianlancetaylor/demangle v0.0.0-20181102032728-5e5cf60278f6 // indirect
	github.com/imdario/mergo v0.3.8 // indirect
	github.com/onsi/ginkgo v1.10.3
	github.com/onsi/gomega v1.7.1
	github.com/opencontainers/go-digest v1.0.0-rc1 // indirect
	github.com/pkg/errors v0.8.1
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cobra v0.0.5
	golang.org/x/arch v0.0.0-20191101135251-a0d8588395bd // indirect
	gopkg.in/yaml.v2 v2.2.5
	k8s.io/api v0.0.0-20191107030003-665c8a257c1a
	k8s.io/apiextensions-apiserver v0.0.0-20191107191557-8263dce1d769
	k8s.io/apimachinery v0.0.0-20191107105744-2c7f8d2b0fd8
	k8s.io/client-go v11.0.0+incompatible
	sigs.k8s.io/kind v0.6.0
)

// pinned 1.15.0
replace (
	k8s.io/api => k8s.io/api v0.0.0-20190620084959-7cf5895f2711
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20190620085554-14e95df34f1f
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190612205821-1799e75a0719
	k8s.io/client-go => k8s.io/client-go v0.0.0-20190620085101-78d2af792bab
)
