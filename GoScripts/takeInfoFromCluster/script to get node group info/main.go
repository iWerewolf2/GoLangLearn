package main

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"sort"

	"github.com/xuri/excelize/v2"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// NodeData: row structure for Excel output
type NodeData struct {
	Name            string
	NodeGroupName   string
	InstanceType    string
	OperatingSystem string
	OSImage         string
	KernelVersion   string
	Architecture    string
	ProvisioningSrc string
	CPU             int64
	MemoryGiB       int64
	DiskGiB         int64
}

func main() {
	// Setup kubeconfig and client
	var kubeconfig string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = filepath.Join(home, ".kube", "config")
	} else {
		log.Fatal("Cannot find home directory for kubeconfig")
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatal(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	// Read nodes
	fmt.Println("Reading cluster nodes...")
	nodes, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Processing %d nodes...\n", len(nodes.Items))

	// Extract per-node info (labels, annotations, capacity)
	var allData []NodeData

	for _, node := range nodes.Items {
		labels := node.GetLabels()
		annotations := node.GetAnnotations()
		capacity := node.Status.Capacity

		// --- A. Detect Node Group ---
		group := "N/A"
		// Try commonly used labels for grouping
		if val, ok := labels["eks.amazonaws.com/nodegroup"]; ok {
			group = val
		} else if val, ok := labels["karpenter.sh/nodepool"]; ok {
			group = val
		} else if val, ok := labels["alpha.eksctl.io/nodegroup-name"]; ok {
			group = val
		} else if val, ok := labels["agentpool"]; ok {
			group = val
		}

		// --- B. Detect Instance Type ---
		instType := "N/A"
		if val, ok := labels["node.kubernetes.io/instance-type"]; ok {
			instType = val
		} else if val, ok := labels["beta.kubernetes.io/instance-type"]; ok {
			instType = val
		}

		// --- C. Detect Provisioning Source (Launch Template / ASG) ---
		src := "N/A"
		// 1. Explicit Launch Template Label
		if val, ok := labels["eks.amazonaws.com/launch-template-name"]; ok {
			src = val
		} else if val, ok := labels["eks.amazonaws.com/source-launch-template-name"]; ok {
			src = val
		}
		// 2. Auto Scaling Group Name (Fallback if LT is missing)
		if src == "N/A" {
			if val, ok := labels["aws.amazon.com/autoscalingGroup"]; ok {
				src = val // The ASG name often helps identify the template
			}
		}
		// 3. Karpenter Provisioner
		if src == "N/A" {
			if val, ok := labels["karpenter.sh/provisioner-name"]; ok {
				src = "Karpenter-" + val
			}
		}
		// 4. Check Annotations (Cluster Autoscaler often puts info here)
		if src == "N/A" {
			if val, ok := annotations["k8s.io/cluster-autoscaler/node-template/label/eks.amazonaws.com/nodegroup"]; ok {
				src = "ASG-" + val
			}
		}

		// --- D. Resources (Fixed v1 usage) ---
		const GiB = 1024 * 1024 * 1024
		cpu := capacity[v1.ResourceCPU]
		mem := capacity[v1.ResourceMemory]
		disk := capacity[v1.ResourceEphemeralStorage]

		allData = append(allData, NodeData{
			Name:            node.Name,
			NodeGroupName:   group,
			InstanceType:    instType,
			OperatingSystem: node.Status.NodeInfo.OperatingSystem,
			OSImage:         node.Status.NodeInfo.OSImage,
			KernelVersion:   node.Status.NodeInfo.KernelVersion,
			Architecture:    node.Status.NodeInfo.Architecture,
			ProvisioningSrc: src,
			CPU:             cpu.Value(),
			MemoryGiB:       mem.Value() / int64(GiB),
			DiskGiB:         disk.Value() / int64(GiB),
		})
	}

	// Sort data
	sort.Slice(allData, func(i, j int) bool {
		if allData[i].NodeGroupName != allData[j].NodeGroupName {
			return allData[i].NodeGroupName < allData[j].NodeGroupName
		}
		return allData[i].Name < allData[j].Name
	})

	// Write report to Excel
	f := excelize.NewFile()
	sheet := "Report"
	index, _ := f.NewSheet(sheet)
	f.SetActiveSheet(index)
	f.DeleteSheet("Sheet1") // Remove default sheet

	headers := []string{"Node Group", "Instance Type", "Prov. Source", "Node Name", "OS Image", "CPU", "Mem(GiB)", "Disk(GiB)"}
	f.SetSheetRow(sheet, "A1", &headers)

	// Style: Bold Headers
	style, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})
	f.SetRowStyle(sheet, 1, 1, style)

	lastGroup := ""
	lastType := ""
	lastSrc := ""

	for i, d := range allData {
		row := i + 2

		// Logic: If this row's group is same as last row, leave cell BLANK
		groupCell := d.NodeGroupName
		typeCell := d.InstanceType
		srcCell := d.ProvisioningSrc

		if d.NodeGroupName == lastGroup {
			groupCell = "" // Hide repeated group name

			// Only hide Instance Type if the Group matches AND the type matches
			if d.InstanceType == lastType {
				typeCell = ""
			}
			// Only hide Source if the Group matches AND the source matches
			if d.ProvisioningSrc == lastSrc {
				srcCell = ""
			}
		}

		// Update trackers
		lastGroup = d.NodeGroupName
		lastType = d.InstanceType
		lastSrc = d.ProvisioningSrc

		// Write Row
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), groupCell)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), typeCell)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), srcCell)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), d.Name)
		f.SetCellValue(sheet, fmt.Sprintf("E%d", row), d.OSImage)
		f.SetCellValue(sheet, fmt.Sprintf("F%d", row), d.CPU)
		f.SetCellValue(sheet, fmt.Sprintf("G%d", row), d.MemoryGiB)
		f.SetCellValue(sheet, fmt.Sprintf("H%d", row), d.DiskGiB)
	}

	// Set column widths for readability
	f.SetColWidth(sheet, "A", "C", 20) // Group, Type, Source
	f.SetColWidth(sheet, "D", "E", 30) // Name, OS

	if err := f.SaveAs("Cluster_Report_Grouped.xlsx"); err != nil {
		log.Fatal(err)
	}
	fmt.Println("✅ Done! File saved as 'Cluster_Report_Grouped.xlsx'")
}
