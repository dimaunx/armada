package image

import (
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/dimaunx/armada/pkg/defaults"

	"github.com/pkg/errors"

	"sigs.k8s.io/kind/pkg/cluster/nodes"
	"sigs.k8s.io/kind/pkg/cluster/nodeutils"

	"github.com/dimaunx/armada/pkg/cluster"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	dockerclient "github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	kind "sigs.k8s.io/kind/pkg/cluster"
)

// ExportLogsFlagpole is a list of cli flags for export logs command
type LoadImagesFlagpole struct {
	Clusters []string
	Images   []string
	Debug    bool
}

// LoadImageCommand returns a new cobra.Command under load command for armada
func LoadImageCommand(provider *kind.Provider) *cobra.Command {
	flags := &LoadImagesFlagpole{}
	cmd := &cobra.Command{
		Args:  cobra.NoArgs,
		Use:   "docker-images",
		Short: "Load docker images in to the cluster",
		Long:  "Load docker images in to the cluster",
		RunE: func(cmd *cobra.Command, args []string) error {

			if flags.Debug {
				log.SetLevel(log.DebugLevel)
			}

			ctx := context.Background()
			dockerCli, err := dockerclient.NewEnvClient()
			if err != nil {
				log.Fatal(err)
			}

			var targetClusters []string
			if len(flags.Clusters) > 0 {
				targetClusters = append(targetClusters, flags.Clusters...)
			} else {
				configFiles, err := ioutil.ReadDir(defaults.KindConfigDir)
				if err != nil {
					log.Fatal(err)
				}
				for _, configFile := range configFiles {
					clName := strings.FieldsFunc(configFile.Name(), func(r rune) bool { return strings.ContainsRune(" -.", r) })[2]
					targetClusters = append(targetClusters, clName)
				}
			}

			if len(targetClusters) > 0 {
				for _, imageName := range flags.Images {
					localImageID, err := getLocalImageID(dockerCli, ctx, imageName)
					if err != nil {
						log.Fatal(err)
					}

					selectedNodes, err := getSelectedNodes(provider, imageName, localImageID, targetClusters)
					if err != nil {
						log.Error(err)
					}
					if len(selectedNodes) > 0 {
						// Create temp dor to images tar
						tempDirName, err := ioutil.TempDir("", "image-tar")
						if err != nil {
							log.Fatal(err)
						}
						// on macOS $TMPDIR is typically /var/..., which is not mountable
						// /private/var/... is the mountable equivalent
						if runtime.GOOS == "darwin" && strings.HasPrefix(tempDirName, "/var/") {
							tempDirName = filepath.Join("/private", tempDirName)
						}
						defer os.RemoveAll(tempDirName)

						imageTarPath, err := save(imageName, tempDirName)
						if err != nil {
							log.Fatal(err)
						}

						log.Infof("loading image: %s to nodes: %s ...", imageName, selectedNodes)
						var wg sync.WaitGroup
						wg.Add(len(selectedNodes))
						for _, node := range selectedNodes {
							go func(node nodes.Node) {
								err = LoadImage(imageTarPath, imageName, node, &wg)
								if err != nil {
									defer wg.Done()
									log.Fatal(err)
								}
							}(node)
						}
						wg.Wait()
					}
				}
			}
			return nil
		},
	}
	cmd.Flags().StringSliceVarP(&flags.Clusters, "cluster", "c", []string{}, "comma separated list of cluster names to load the image in to.")
	cmd.Flags().StringSliceVarP(&flags.Images, "image", "i", []string{}, "comma separated list images to load.")
	cmd.Flags().BoolVarP(&flags.Debug, "debug", "v", false, "set log level to debug")
	return cmd
}

// getLocalImageID returns image id by name/reference
func getLocalImageID(dockerCli *dockerclient.Client, ctx context.Context, imageName string) (string, error) {
	imageFilter := filters.NewArgs()
	imageFilter.Add("reference", imageName)
	result, err := dockerCli.ImageList(ctx, types.ImageListOptions{
		All:     false,
		Filters: imageFilter,
	})
	if err != nil {
		return "", err
	}
	if len(result) == 0 {
		return "", errors.Errorf("Image %s not found locally.", imageName)
	}
	return result[0].ID, nil
}

// getSelectedNodes return a list of nodes that don't have the image for multiple clusters
func getSelectedNodes(provider *kind.Provider, imageName, localImageID string, clusters []string) ([]nodes.Node, error) {
	var selectedNodes []nodes.Node
	for _, clName := range clusters {
		known, err := cluster.IsKnown(clName, provider)
		if err != nil {
			return nil, err
		}
		if known {
			nodeList, err := provider.ListInternalNodes(clName)
			if err != nil {
				return nil, err
			}
			if len(nodeList) == 0 {
				return nil, errors.Errorf("no nodes found for cluster %q", clName)
			}

			// pick only the nodes that don't have the image
			for _, node := range nodeList {
				nodeImageID, err := nodeutils.ImageID(node, imageName)
				if err != nil || nodeImageID != localImageID {
					selectedNodes = append(selectedNodes, node)
					log.Debugf("%s: image: %q with ID %q not present on node %q", clName, imageName, localImageID, node.String())
				}
				if nodeImageID == localImageID {
					log.Infof("%s: ✔ image with ID %q already present on node %q", clName, nodeImageID, node.String())
				}
			}
		} else {
			return selectedNodes, errors.Errorf("cluster %q not found.", clName)
		}
	}
	return selectedNodes, nil
}

func save(imageName, tempDirName string) (string, error) {
	imageTarPath := filepath.Join(tempDirName, "image.tar")
	cmdName := "docker"
	cmdArgs := []string{"save", "-o", imageTarPath, imageName}

	cmd := exec.Command(cmdName, cmdArgs...)
	buf := &bytes.Buffer{}
	cmd.Stdout = buf
	cmd.Stderr = buf

	log.Debugf("saving image: %q tar to %q ...", imageName, imageTarPath)
	if err := cmd.Run(); err != nil {
		return "", errors.Wrapf(err, "failed to save image: %q, location: %q", imageName, imageTarPath)
	}
	return imageTarPath, nil
}

func LoadImage(imageTarPath, imageName string, node nodes.Node, wg *sync.WaitGroup) error {
	f, err := os.Open(imageTarPath)
	if err != nil {
		return errors.Wrapf(err, "failed to open an image: %s, location: %q", imageName, imageTarPath)
	}
	defer f.Close()

	log.Debugf("loading image: %q, path: %q to node: %q ...", imageName, imageTarPath, node.String())
	err = nodeutils.LoadImageArchive(node, f)
	if err != nil {
		return errors.Wrapf(err, "failed to loading image: %q, node %q", imageName, node.String())
	}
	log.Infof("✔ image: %q was loaded to node: %q.", imageName, node.String())
	wg.Done()
	return nil
}
