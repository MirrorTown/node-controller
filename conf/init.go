package conf

import (
	"fmt"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"k8s.io/client-go/informers"
	coreinformers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientSet "node-controller/generated/clientset/versioned"
	"os"
	"path/filepath"
	"runtime"
)

var (
	ENV                       string
	basePath                  string
	K8sAuth                   k8sAuth
	kubeClient                kubernetes.Interface
	CoreSharedInformerFactory informers.SharedInformerFactory
	PodInformer               coreinformers.PodInformer
)

const LOCAL = "local"
const confPath = "/Users/dasouche/.kube/config"

var config *restclient.Config

type k8sAuth struct {
	file string
}

func init() {
	// 获取当前环境
	setEnvironment()
	// 获取配置文件路径
	filePath := getconfPath()
	// 获取配置文件内容
	if configurationContent, err := ioutil.ReadFile(filePath); err != nil {
		panic(fmt.Sprintf("failed to read configuration file: %s", filePath))
	} else {
		configuration := gjson.ParseBytes(configurationContent)

		// k8sAuth conf
		k8sAuthConf := configuration.Get("conf.k8sAuth")
		K8sAuth.file = k8sAuthConf.Get("file").String()
	}

	// init informer
	InitInformer()
}

func setEnvironment() {
	if env := os.Getenv("ENV"); env == "" {
		ENV = LOCAL
	} else {
		ENV = env
	}
	_, basePath, _, _ = runtime.Caller(1)
}

func getconfPath() string {
	if ENV == LOCAL {
		return filepath.Join(filepath.Dir(basePath), "conf.json")
	} else {
		return confPath
	}
}

func InitVirtulMachineClient() clientSet.Interface {
	nodeCLientSet, err := clientSet.NewForConfig(config)
	ExceptNilErr(err)
	return nodeCLientSet
}

func GetKubeClient() kubernetes.Interface {
	return kubeClient
}

func InitInformer() {
	// 生成一个k8s client
	//var config *restclient.Config
	var err error
	if ENV == LOCAL {
		clientConfig, err := clientcmd.LoadFromFile(K8sAuth.file)
		ExceptNilErr(err)

		config, err = clientcmd.NewDefaultClientConfig(*clientConfig, &clientcmd.ConfigOverrides{}).ClientConfig()
		ExceptNilErr(err)
	} else {
		config, err = restclient.InClusterConfig()
		ExceptNilErr(err)
	}

	//k8sClient, err := kubernetes.NewForConfig(config)
	kubeClient = InitKubeClient()
	ExceptNilErr(err)

	// 创建一个informerFactory
	//sharedInformerFactory := informers.NewSharedInformerFactory(k8sClient, 0)
	// 创建一个informerFactory
	CoreSharedInformerFactory = informers.NewSharedInformerFactory(kubeClient, 0)

	// 创建 informers
	PodInformer = CoreSharedInformerFactory.Core().V1().Pods()
}

func InitKubeClient() kubernetes.Interface {
	//var err error
	//if ENV == LOCAL {
	//	clientConfig, err := clientcmd.LoadFromFile(K8sAuth.file)
	//	ExceptNilErr(err)
	//
	//	config, err = clientcmd.NewDefaultClientConfig(*clientConfig, &clientcmd.ConfigOverrides{}).ClientConfig()
	//	ExceptNilErr(err)
	//} else {
	//	config, err = restclient.InClusterConfig()
	//	ExceptNilErr(err)
	//}

	k8sClient, err := kubernetes.NewForConfig(config)
	ExceptNilErr(err)
	return k8sClient
}

func ExceptNilErr(err error) {
	if err != nil {
		panic(err)
	}
}
