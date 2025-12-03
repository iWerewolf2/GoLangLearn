package main

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/xuri/excelize/v2"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var ctx = context.TODO()
var f *excelize.File
var headerStyle *int

func getClientSet() (*kubernetes.Clientset, error) {
	var kubeconfig string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = filepath.Join(home, ".kube", "config")
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("error building kubeconfig: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("error creating clientset: %w", err)
	}
	return clientset, nil
}

func writeRow(sheet string, row int, data []interface{}) {
	cell, _ := excelize.CoordinatesToCellName(1, row)
	if err := f.SetSheetRow(sheet, cell, &data); err != nil {
		log.Printf("Error writing row to sheet %s: %v", sheet, err)
	}
}

func setHeaders(sheet string, headers []string) {
	if err := f.SetSheetRow(sheet, "A1", &headers); err != nil {
		log.Printf("Error setting headers for sheet %s: %v", sheet, err)
	}
	if err := f.SetCellStyle(sheet, "A1", fmt.Sprintf("%c1", 'A'+len(headers)-1), *headerStyle); err != nil {
		log.Printf("Error applying style to headers for sheet %s: %v", sheet, err)
	}
}

func fetchWorkloads(clientset *kubernetes.Clientset) {
	sheetName := "Workloads"
	f.NewSheet(sheetName)
	headers := []string{"Kind", "Namespace", "Name", "Replicas", "Primary Container Image"}
	setHeaders(sheetName, headers)
	row := 2

	deployments, err := clientset.AppsV1().Deployments(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})
	if err != nil {
		log.Printf("Error getting deployments: %v", err)
	} else {
		for _, item := range deployments.Items {
			image := "N/A"
			if len(item.Spec.Template.Spec.Containers) > 0 {
				image = item.Spec.Template.Spec.Containers[0].Image
			}
			data := []interface{}{"Deployment", item.Namespace, item.Name, *item.Spec.Replicas, image}
			writeRow(sheetName, row, data)
			row++
		}
	}

	statefulsets, err := clientset.AppsV1().StatefulSets(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})
	if err != nil {
		log.Printf("Error getting statefulsets: %v", err)
	} else {
		for _, item := range statefulsets.Items {
			image := "N/A"
			if len(item.Spec.Template.Spec.Containers) > 0 {
				image = item.Spec.Template.Spec.Containers[0].Image
			}
			data := []interface{}{"StatefulSet", item.Namespace, item.Name, *item.Spec.Replicas, image}
			writeRow(sheetName, row, data)
			row++
		}
	}

	daemonsets, err := clientset.AppsV1().DaemonSets(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})
	if err != nil {
		log.Printf("Error getting daemonsets: %v", err)
	} else {
		for _, item := range daemonsets.Items {
			image := "N/A"
			if len(item.Spec.Template.Spec.Containers) > 0 {
				image = item.Spec.Template.Spec.Containers[0].Image
			}
			data := []interface{}{"DaemonSet", item.Namespace, item.Name, "N/A (Per-Node)", image}
			writeRow(sheetName, row, data)
			row++
		}
	}
	log.Printf("Fetched %d workloads", row-2)
}

func fetchStorage(clientset *kubernetes.Clientset) {
	sheetName := "Storage"
	f.NewSheet(sheetName)
	headers := []string{"Kind", "Name", "Capacity", "StorageClass", "Status", "Claim Namespace", "Claim Name"}
	setHeaders(sheetName, headers)
	row := 2

	pvs, err := clientset.CoreV1().PersistentVolumes().List(ctx, metav1.ListOptions{})
	if err != nil {
		log.Printf("Error getting persistent volumes: %v", err)
	} else {
		for _, item := range pvs.Items {
			capacity := item.Spec.Capacity[v1.ResourceStorage]
			claimNS := ""
			claimName := ""
			if item.Spec.ClaimRef != nil {
				claimNS = item.Spec.ClaimRef.Namespace
				claimName = item.Spec.ClaimRef.Name
			}
			data := []interface{}{"PersistentVolume", item.Name, capacity.String(), item.Spec.StorageClassName, item.Status.Phase, claimNS, claimName}
			writeRow(sheetName, row, data)
			row++
		}
	}
	log.Printf("Fetched %d storage volumes", row-2)
}

func fetchSecrets(clientset *kubernetes.Clientset) {
	sheetName := "Secrets"
	f.NewSheet(sheetName)
	headers := []string{"Namespace", "Name", "Type", "Data Keys"}
	setHeaders(sheetName, headers)
	row := 2

	secrets, err := clientset.CoreV1().Secrets(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})
	if err != nil {
		log.Printf("Error getting secrets: %v", err)
	} else {
		for _, item := range secrets.Items {
			if item.Type == v1.SecretTypeServiceAccountToken {
				continue
			}
			var keys []string
			for k := range item.Data {
				keys = append(keys, k)
			}
			data := []interface{}{item.Namespace, item.Name, item.Type, strings.Join(keys, ", ")}
			writeRow(sheetName, row, data)
			row++
		}
	}
	log.Printf("Fetched %d non-default secrets", row-2)
}

func fetchNetwork(clientset *kubernetes.Clientset) {
	sheetName := "NetworkExposure"
	f.NewSheet(sheetName)
	headers := []string{"Kind", "Namespace", "Name", "Type", "External IP / Host", "Ports"}
	setHeaders(sheetName, headers)
	row := 2

	services, err := clientset.CoreV1().Services(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})
	if err != nil {
		log.Printf("Error getting services: %v", err)
	} else {
		for _, item := range services.Items {
			if item.Spec.Type == v1.ServiceTypeLoadBalancer || item.Spec.Type == v1.ServiceTypeNodePort {
				ip := "Pending"
				if len(item.Status.LoadBalancer.Ingress) > 0 {
					if item.Status.LoadBalancer.Ingress[0].IP != "" {
						ip = item.Status.LoadBalancer.Ingress[0].IP
					} else {
						ip = item.Status.LoadBalancer.Ingress[0].Hostname
					}
				}
				var ports []string
				for _, p := range item.Spec.Ports {
					ports = append(ports, fmt.Sprintf("%d:%d", p.Port, p.NodePort))
				}
				data := []interface{}{"Service", item.Namespace, item.Name, item.Spec.Type, ip, strings.Join(ports, ", ")}
				writeRow(sheetName, row, data)
				row++
			}
		}
	}

	ingresses, err := clientset.NetworkingV1().Ingresses(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})
	if err != nil {
		log.Printf("Error getting ingresses: %v", err)
	} else {
		for _, item := range ingresses.Items {
			ip := "Pending"
			if len(item.Status.LoadBalancer.Ingress) > 0 {
				if item.Status.LoadBalancer.Ingress[0].IP != "" {
					ip = item.Status.LoadBalancer.Ingress[0].IP
				} else {
					ip = item.Status.LoadBalancer.Ingress[0].Hostname
				}
			}
			var hosts []string
			for _, rule := range item.Spec.Rules {
				hosts = append(hosts, rule.Host)
			}
			data := []interface{}{"Ingress", item.Namespace, item.Name, "N/A", ip, strings.Join(hosts, ", ")}
			writeRow(sheetName, row, data)
			row++
		}
	}
	log.Printf("Fetched %d network exposure points", row-2)
}

func fetchAccess(clientset *kubernetes.Clientset) {
	sheetName := "AccessControls"
	f.NewSheet(sheetName)
	headers := []string{"Kind", "Namespace", "Name"}
	setHeaders(sheetName, headers)
	row := 2

	sas, err := clientset.CoreV1().ServiceAccounts(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})
	if err != nil {
		log.Printf("Error getting service accounts: %v", err)
	} else {
		for _, item := range sas.Items {
			if item.Name == "default" {
				continue
			}
			data := []interface{}{"ServiceAccount", item.Namespace, item.Name}
			writeRow(sheetName, row, data)
			row++
		}
	}
	log.Printf("Fetched %d non-default ServiceAccounts", row-2)
}

func main() {
	log.Println("Starting K8s asset inventory export...")

	clientset, err := getClientSet()
	if err != nil {
		log.Fatalf("Failed to connect to Kubernetes: %v", err)
	}
	log.Println("Successfully connected to Kubernetes cluster.")

	var kubeconfig string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = filepath.Join(home, ".kube", "config")
	}
	config, err := clientcmd.LoadFromFile(kubeconfig)
	if err != nil {
		log.Fatalf("Failed to load kubeconfig: %v", err)
	}
	currentContext := config.CurrentContext
	clusterName := "default-cluster"
	if ctx, exists := config.Contexts[currentContext]; exists {
		clusterName = ctx.Cluster
	}

	sanitizedClusterName := strings.ReplaceAll(clusterName, ":", "-")
	sanitizedClusterName = strings.ReplaceAll(sanitizedClusterName, "/", "-")

	f = excelize.NewFile()

	style, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
	})
	if err != nil {
		log.Fatalf("Failed to create header style: %v", err)
	}
	headerStyle = &style

	fetchWorkloads(clientset)
	fetchStorage(clientset)
	fetchSecrets(clientset)
	fetchNetwork(clientset)
	fetchAccess(clientset)

	f.DeleteSheet("Sheet1")

	filename := fmt.Sprintf("k8s_asset_inventory_%s.xlsx", sanitizedClusterName)
	if err := f.SaveAs(filename); err != nil {
		log.Fatalf("Failed to save XLSX file: %v", err)
	}

	log.Printf("✅ Successfully exported all assets to %s", filename)
}
