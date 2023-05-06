package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/hpcloud/tail"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
)

func sliceContains(watches []string, e string) bool {
	for _, a := range watches {
		if a == e {
			return true
		}
	}
	return false
}

func killPod(namespace, podName string, clientset kubernetes.Interface) {
	ctx := context.Background()
	err := clientset.CoreV1().Pods(namespace).Delete(ctx, podName, metav1.DeleteOptions{})
	if err != nil {
		log.Fatal(err)
	}
}

func retiree(file string, clientset kubernetes.Interface) {
	t, err := tail.TailFile(file, tail.Config{Follow: true})
	if err != nil {
		log.Fatal(err)
	}
	for line := range t.Lines {
		if strings.Contains(line.Text, os.Getenv("DEATHNOTE")) {
			fmt.Println("death note: " + line.Text)
			tokenizedFilename := strings.Split(file, "_")
			killPod(tokenizedFilename[1], tokenizedFilename[0], clientset)
		}
	}
}

func main() {
	var watches []string

	for {
		files, err := ioutil.ReadDir("/var/log/containers")
		if err != nil {
			log.Fatal(err)
		}

		config := ctrl.GetConfigOrDie()
		clientset := kubernetes.NewForConfigOrDie(config)

		for _, file := range files {
			if !file.IsDir() && strings.Contains(file.Name(), os.Getenv("POD_FILTER")) {
				if !sliceContains(watches, file.Name()) {
					fmt.Println("watching " + "/var/log/containers/" + file.Name())
					watches = append(watches, file.Name())
					go retiree("/var/log/containers/"+file.Name(), clientset)
				}
			}
		}

		time.Sleep(10 * time.Second)
	}
}
