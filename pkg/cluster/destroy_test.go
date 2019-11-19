package cluster

import (
	"io/ioutil"
	"strings"
	"sync"

	"github.com/dimaunx/armada/pkg/config"
	"github.com/gobuffalo/packr/v2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	kind "sigs.k8s.io/kind/pkg/cluster"
)

var _ = Describe("Destroy cluster", func() {
	BeforeEach(func() {
		flags := config.Flagpole{
			Wait:        0,
			NumClusters: 1,
		}

		box := packr.New("configs", "../../configs")

		cl := config.Cluster{
			Cni:                 "kindnet",
			Name:                "cl1",
			PodSubnet:           "10.4.0.0/14",
			ServiceSubnet:       "100.1.0.0/14",
			DNSDomain:           "cl1.local",
			KubeAdminAPIVersion: "kubeadm.k8s.io/v1beta2",
			NumWorkers:          0,
		}

		var wg sync.WaitGroup
		wg.Add(1)
		err := Create(&cl, &flags, box, &wg)
		Ω(err).ShouldNot(HaveOccurred())
		wg.Wait()
	})
	Context("component: Destruction", func() {
		It("Should destroy a cluster", func() {
			configFiles, err := ioutil.ReadDir(config.KindConfigDir)
			Ω(err).ShouldNot(HaveOccurred())

			for _, file := range configFiles {
				clName := strings.FieldsFunc(file.Name(), func(r rune) bool { return strings.ContainsRune(" -.", r) })[2]
				err := Destroy(clName)
				Ω(err).ShouldNot(HaveOccurred())
				known, err := kind.IsKnown(clName)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(known).ShouldNot(BeTrue())
			}
		})
	})
})
