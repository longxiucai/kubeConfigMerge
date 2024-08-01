package main

import (
	"fmt"
	"os"

	"github.com/longxiucai/kubeConfigMerge/pkg/merge"
	"github.com/longxiucai/kubeConfigMerge/pkg/util"
	"github.com/spf13/pflag"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

func main() {
	var files []string
	var contextTemplate []string
	var cfgFile string
	contextTemplate = []string{"context"}
	pflag.StringVar(&cfgFile, "output", "merge.config", "out put path to the kubeconfig file")
	pflag.StringSliceVar(&files, "file", []string{}, "path to merge kubeconfig files")
	pflag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		printUsage()
		pflag.PrintDefaults()
	}

	pflag.Parse()
	outConfigs := clientcmdapi.NewConfig()
	if len(files) == 1 && files[0] == "." {
		dir, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		fs, err := os.ReadDir(dir)
		if err != nil {
			panic(err)
		}
		var fileNames []string
		for _, file := range fs {
			if !file.IsDir() {
				fileNames = append(fileNames, file.Name())
			}
		}
		files = fileNames
	}
	for _, f := range files {
		fmt.Printf("Loading KubeConfig file: " + f + " \n")
		loadConfig, err := util.LoadKubeConfig(f)
		if err != nil {
			fmt.Printf("[skip] File " + f + " is not kubeconfig\n")
			continue
		}
		kco := &merge.KubeConfigOption{
			Config:   loadConfig,
			FileName: util.GetFileName(f),
		}
		outConfigs, err = kco.HandleContexts(outConfigs, contextTemplate)
		if err != nil {
			panic(err)
		}
	}
	if len(outConfigs.Contexts) == 0 {
		fmt.Println("No context to merge.")
		return
	}
	err := util.WriteConfig(cfgFile, outConfigs)
	if err != nil {
		panic(err)
	}
}
func printUsage() {
	fmt.Println(`Examples:
# Merge multiple kubeconfig from current directory,default output to ./merge.config
kubeconfigmerge --file .
# Merge multiple kubeconfig,If the context name is repeated, use the file name and a hash of its contents as the context name.
kubeconfigmerge --file=1st.yaml,2nd.yaml,3rd.yaml
# Merge multiple kubeconfig,Specify output file.
kubeconfigmerge --file=1st.yaml,2nd.yaml,3rd.yaml --output=all-config.yaml
	`)
}
