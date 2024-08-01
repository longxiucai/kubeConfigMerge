package util

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"sort"
	"strings"

	ct "github.com/daviddengcn/go-colortext"

	"github.com/bndr/gotabulate"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	clientcmdlatest "k8s.io/client-go/tools/clientcmd/api/latest"
)

var (
	silenceTable bool
)

func LoadKubeConfig(yaml string) (*clientcmdapi.Config, error) {
	loadConfig, err := clientcmd.LoadFromFile(yaml)
	if err != nil {
		return nil, err
	}
	if len(loadConfig.Contexts) == 0 {
		return nil, fmt.Errorf("no kubeconfig in %s ", yaml)
	}
	return loadConfig, err
}

func GetFileName(path string) string {
	n := strings.Split(path, "/")
	result := strings.Split(n[len(n)-1], ".")
	return result[0]
}

// Hash returns the hex form of the sha256 of the argument.
func Hash(data string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(data)))
}

// HashSuf return the string of kubeconfig.
func HashSuf(config *clientcmdapi.Config) string {
	reJSON, err := runtime.Encode(clientcmdlatest.Codec, config)
	if err != nil {
		fmt.Printf("Unexpected error: %v", err)
	}
	sum, _ := hEncode(Hash(string(reJSON)))
	return sum
}

// HashSufString return the string of HashSuf.
func HashSufString(data string) string {
	sum, _ := hEncode(Hash(data))
	return sum
}

// Copied from https://github.com/kubernetes/kubernetes
// /blob/master/pkg/kubectl/util/hash/hash.go
func hEncode(hex string) (string, error) {
	if len(hex) < 10 {
		return "", fmt.Errorf(
			"input length must be at least 10")
	}
	enc := []rune(hex[:10])
	for i := range enc {
		switch enc[i] {
		case '0':
			enc[i] = 'g'
		case '1':
			enc[i] = 'h'
		case '3':
			enc[i] = 'k'
		case 'a':
			enc[i] = 'm'
		case 'e':
			enc[i] = 't'
		}
	}
	return string(enc), nil
}

// WriteConfig write kubeconfig
func WriteConfig(cfgFile string, outConfig *clientcmdapi.Config) error {
	if strings.HasPrefix(cfgFile, "~/") {
		cfgFile = filepath.Join(homeDir(), cfgFile[2:])
	}
	err := clientcmd.WriteToFile(*outConfig, cfgFile)
	if err != nil {
		return err
	}
	fmt.Printf("「%s」 write successful!\n", cfgFile)

	if !silenceTable {
		err = PrintTable(outConfig)
		if err != nil {
			return err
		}
	}

	return nil
}
func homeDir() string {
	usr, err := user.Current()
	if err != nil {
		return "./"
	}
	return usr.HomeDir
}
func PrintYellow(out io.Writer, content string) {
	ct.ChangeColor(ct.Yellow, false, ct.None, false)
	fmt.Fprint(out, content)
	ct.ResetColor()
}

// PrintTable generate table
func PrintTable(config *clientcmdapi.Config) error {
	var table [][]string
	sortedKeys := make([]string, 0)
	for k := range config.Contexts {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)
	ctx := config.Contexts
	for _, k := range sortedKeys {
		namespace := "default"
		head := ""
		if config.CurrentContext == k {
			head = "*"
		}
		if ctx[k].Namespace != "" {
			namespace = ctx[k].Namespace
		}
		if config.Clusters == nil {
			continue
		}
		cluster, ok := config.Clusters[ctx[k].Cluster]
		if !ok {
			continue
		}
		conTmp := []string{head, k, ctx[k].Cluster, ctx[k].AuthInfo, cluster.Server, namespace}
		table = append(table, conTmp)
	}

	if table != nil {
		tabulate := gotabulate.Create(table)
		tabulate.SetHeaders([]string{"CURRENT", "NAME", "CLUSTER", "USER", "SERVER", "Namespace"})
		// Turn On String Wrapping
		tabulate.SetWrapStrings(true)
		// Render the table
		tabulate.SetAlign("center")
		fmt.Println(tabulate.Render("grid", "left"))
	} else {
		return errors.New("context not found")
	}
	return nil
}
func IsFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}
