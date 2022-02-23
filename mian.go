package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path/filepath"

	"github.com/ghodss/yaml"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func main() {
	clusterURL, err := url.Parse(os.Getenv("cluster_url"))
	if err != nil {
		log.Fatalln(err)
	}

	config := &rest.Config{
		Host:    clusterURL.Host,
		APIPath: clusterURL.Path,
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: true,
		},
		BearerToken: os.Getenv("token"),
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalln(err)
	}

	cm, err := clientset.CoreV1().
		ConfigMaps(os.Getenv("namespace")).
		Get(context.TODO(), os.Getenv("name"), metav1.GetOptions{})

	data := make(map[string]string)
	if os.Getenv("files") != "" {
		//get file content
		files := make([]string, 0)
		err = yaml.Unmarshal([]byte(os.Getenv("files")), &files)
		if err != nil {
			err = nil
			err = json.Unmarshal([]byte(os.Getenv("files")), &files)
			if err != nil {
				log.Fatalln(err)
			}
		}
		for i := range files {
			rb, err := ioutil.ReadFile(files[i])
			if err != nil {
				log.Fatalln(err)
			}
			data[filepath.Base(files[i])] = string(rb)
		}
	}
	if os.Getenv("data") != "" || os.Getenv("data") != "{}" {
		params := make(map[string]string)
		err = yaml.Unmarshal([]byte(os.Getenv("data")), &params)
		if err != nil {
			err = nil
			err = json.Unmarshal([]byte(os.Getenv("data")), &params)
			if err != nil {
				log.Fatalln(err)
			}
		}
		for k := range params {
			data[k] = params[k]
		}

	}

	if errors.IsNotFound(err) {
		cm = &corev1.ConfigMap{}
		cm.Namespace = os.Getenv("namespace")
		cm.Name = os.Getenv("name")
		cm.Data = data
		_, err = clientset.CoreV1().ConfigMaps(cm.Namespace).Create(context.TODO(), cm, metav1.CreateOptions{})
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Printf("create configmap(%s) success\n", cm.Name)
		return
	}
	if cm.Data == nil {
		cm.Data = make(map[string]string)
	}
	for k := range data {
		cm.Data[k] = data[k]
	}
	cm.Namespace = os.Getenv("namespace")
	cm.Name = os.Getenv("name")
	_, err = clientset.CoreV1().ConfigMaps(cm.Namespace).Update(context.TODO(), cm, metav1.UpdateOptions{})
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Printf("update configmap(%s) success\n", cm.Name)
}