package merge

import (
	"fmt"
	"os"
	"strings"

	"github.com/imdario/mergo"
	"github.com/longxiucai/kubeConfigMerge/pkg/util"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// KubeConfigOption kubeConfig option
type KubeConfigOption struct {
	Config   *clientcmdapi.Config
	FileName string
}

const (
	Filename  = "filename"
	Context   = "context"
	User      = "user"
	Cluster   = "cluster"
	Namespace = "namespace"
)

func (kc *KubeConfigOption) HandleContexts(oldConfig *clientcmdapi.Config, contextTemplate []string) (*clientcmdapi.Config, error) {
	newConfig := clientcmdapi.NewConfig()
	var newName string
	generatedName := make(map[string]int)

	for name, ctx := range kc.Config.Contexts {
		newName = kc.generateContextName(name, ctx, contextTemplate)
		generatedName[newName]++
		if generatedName[newName] > 1 {
			newName = fmt.Sprintf("%s-%d", newName, generatedName[newName])
		}

		if checkContextName(newName, oldConfig) {
			fmt.Printf("Context「%s」name already exists, Using the file name and a hash of its contents as the context name\n", newName)
			b, _ := os.ReadFile(kc.FileName)
			suffix := util.HashSufString(string(b))
			newName = kc.FileName + "-" + suffix
		}

		itemConfig := kc.handleContext(oldConfig, newName, ctx)
		newConfig = appendConfig(newConfig, itemConfig)
		fmt.Printf("Add Context: %s from file:%s\n", newName, kc.FileName)
	}
	outConfig := appendConfig(oldConfig, newConfig)
	outConfig.CurrentContext = newName
	return outConfig, nil
}

func (kc *KubeConfigOption) generateContextName(name string, ctx *clientcmdapi.Context, contextTemplate []string) string {
	valueMap := map[string]string{
		Filename:  kc.FileName,
		Context:   name,
		User:      ctx.AuthInfo,
		Cluster:   ctx.Cluster,
		Namespace: ctx.Namespace,
	}

	var contextValues []string
	for _, value := range contextTemplate {
		if v, ok := valueMap[value]; ok {
			if v != "" {
				contextValues = append(contextValues, v)
			}
		}
	}

	return strings.Join(contextValues, "-")
}

func checkContextName(name string, oldConfig *clientcmdapi.Config) bool {
	if _, ok := oldConfig.Contexts[name]; ok {
		return true
	}
	return false
}

func checkClusterAndUserName(oldConfig *clientcmdapi.Config, newClusterName, newUserName string) (bool, bool) {
	var (
		isClusterNameExist bool
		isUserNameExist    bool
	)

	for _, ctx := range oldConfig.Contexts {
		if ctx.Cluster == newClusterName {
			isClusterNameExist = true
		}
		if ctx.AuthInfo == newUserName {
			isUserNameExist = true
		}
	}

	return isClusterNameExist, isUserNameExist
}

func (kc *KubeConfigOption) handleContext(oldConfig *clientcmdapi.Config,
	name string, ctx *clientcmdapi.Context) *clientcmdapi.Config {

	var (
		clusterNameSuffix string
		userNameSuffix    string
	)

	isClusterNameExist, isUserNameExist := checkClusterAndUserName(oldConfig, ctx.Cluster, ctx.AuthInfo)
	newConfig := clientcmdapi.NewConfig()
	suffix := util.HashSufString(name)

	if isClusterNameExist {
		clusterNameSuffix = "-" + suffix
	}
	if isUserNameExist {
		userNameSuffix = "-" + suffix
	}

	userName := fmt.Sprintf("%v%v", ctx.AuthInfo, userNameSuffix)
	clusterName := fmt.Sprintf("%v%v", ctx.Cluster, clusterNameSuffix)
	newCtx := ctx.DeepCopy()
	newConfig.AuthInfos[userName] = kc.Config.AuthInfos[newCtx.AuthInfo]
	newConfig.Clusters[clusterName] = kc.Config.Clusters[newCtx.Cluster]
	newConfig.Contexts[name] = newCtx
	newConfig.Contexts[name].AuthInfo = userName
	newConfig.Contexts[name].Cluster = clusterName

	return newConfig
}

func appendConfig(c1, c2 *clientcmdapi.Config) *clientcmdapi.Config {
	config := clientcmdapi.NewConfig()
	_ = mergo.Merge(config, c1)
	_ = mergo.Merge(config, c2)
	return config
}
