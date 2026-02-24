package apiserver

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"KubePot/core/pool"
	"KubePot/core/rpc/client"
	"KubePot/utils/is"
	"KubePot/utils/log"
)

// 服务运行状态标志
var serverRunning bool

// Start 启动 Kubelet 蜜罐服务
func Start(addr string) {
	// 检查服务是否已经在运行
	if serverRunning {
		fmt.Printf("Apiserver 服务已经在运行，跳过启动\n")
		return
	}

	// 建立socket，监听端口
	netListen, err := net.Listen("tcp", addr)
	if err != nil {
		log.Pr("apiserver", "127.0.0.1", "apiserver 监听失败", err)
		return
	}
	defer netListen.Close()

	// 设置服务运行状态为true
	serverRunning = true
	defer func() {
		serverRunning = false
	}()

	log.Pr("apiserver", addr, "蜜罐服务已启动")

	// 创建连接池
	wg, poolX := pool.New(10)
	defer poolX.Release()

	// 循环接受连接
	for {
		wg.Add(1)
		poolX.Submit(func() {
			// 稍微延迟，避免过快接受连接
			time.Sleep(time.Second * 1)

			conn, err := netListen.Accept()
			if err != nil {
				log.Pr("Apiserver", "127.0.0.1", "apiserver 连接失败", err)
				wg.Done()
				return
			}

			// 获取客户端IP
			clientIP := strings.Split(conn.RemoteAddr().String(), ":")[0]
			log.Pr("Apiserver", clientIP, "已经连接")

			// 上报连接事件
			var attackID string

			// 处理连接
			go handleConnection(conn, attackID, clientIP)
			wg.Done()
		})
	}
}

// handleConnection 处理客户端连接
func handleConnection(conn net.Conn, attackID string, clientIP string) {
	defer conn.Close()

	// 创建缓冲区
	buffer := make([]byte, 4096)

	// 循环读取客户端请求
	for {
		// 读取请求数据
		n, err := conn.Read(buffer)
		if err != nil {
			if err != io.EOF {
				log.Pr("apiserver", clientIP, "读取数据失败", err)
			}
			break
		}

		// 解析请求路径
		requestData := string(buffer[:n])
		path := extractPath(requestData)
		//queryParams := extractQueryParams(requestData)

		// 记录请求
		log.Pr("Apiserver", clientIP, "请求路径", path)
		info := path
		if is.Rpc() {
			go client.ReportResult("Apiserver", "Apiserver 蜜罐", conn.RemoteAddr().String(), info, attackID)
		} else {
			//go report.ReportKubelet("Kubelet", "本机", conn.RemoteAddr().String(), path)
		}

		// 获取响应数据
		responseData := getResponseData(path)

		if responseData != "ok" {
			// 将响应数据转换为 JSON
			jsonData, err := json.MarshalIndent(responseData, "", "  ")
			if err != nil {
				log.Pr("Kubelet", clientIP, "JSON 编码失败", err)
				conn.Write([]byte("HTTP/1.1 500 Internal Server Error\r\nContent-Type: text/plain\r\n\r\nInternal Server Error"))
				continue
			}

			// 构建 HTTP 响应
			response := fmt.Sprintf(
				"HTTP/1.1 200 OK\r\nContent-Type: application/json\r\nContent-Length: %d\r\n\r\n%s",
				len(jsonData), jsonData,
			)

			// 发送响应
			conn.Write([]byte(response))
		} else {
			// 构建 HTTP 响应
			response := fmt.Sprintf(
				"HTTP/1.1 200 OK\r\nContent-Type: text/plain; \r\nContent-Length: %d\r\n\r\n%s",
				len(responseData.(string)), responseData.(string),
			)
			// 发送响应
			conn.Write([]byte(response))
		}

	}
}

// extractPath 从请求中提取路径（去除参数）
func extractPath(request string) string {
	lines := strings.Split(request, "\r\n")
	if len(lines) > 0 {
		parts := strings.Split(lines[0], " ")
		if len(parts) > 1 {
			// 提取路径并去除查询参数（如 ?limit=500）
			path := parts[1]
			queryStart := strings.Index(path, "?")
			if queryStart > 0 {
				return path[:queryStart]
			}
			return path
		}
	}
	return "/"
}

// // extractPath 从请求中提取路径
// func extractPath(request string) string {
// 	lines := strings.Split(request, "\r\n")
// 	if len(lines) > 0 {
// 		parts := strings.Split(lines[0], " ")
// 		if len(parts) > 1 {
// 			return parts[1]
// 		}
// 	}
// 	return "/"
// }

// getResponseData 根据路径获取响应数据
func getResponseData(path string) interface{} {
	// // 初始化 namespacePodData 结构体
	// var namespacePodData = map[string]interface{}{
	// 	"default": map[string]interface{}{
	// 		"kind":       "PodList",
	// 		"apiVersion": "v1",
	// 		"metadata": map[string]interface{}{
	// 			"resourceVersion": "20529619",
	// 		},
	// 		"items": []map[string]interface{}{
	// 			{
	// 				"metadata": map[string]interface{}{
	// 					"name":              "tomcat01-7dccfcbff8-qdgns",
	// 					"generateName":      "tomcat01-7dccfcbff8-",
	// 					"namespace":         "default",
	// 					"uid":               "a186dd48-61ea-4b2b-bb4b-c1a818d8daca",
	// 					"resourceVersion":   "20529092",
	// 					"creationTimestamp": "2025-06-28T12:05:19Z",
	// 					"labels": map[string]string{
	// 						"app":               "tomcat01",
	// 						"pod-template-hash": "7dccfcbff8",
	// 					},
	// 					"ownerReferences": []map[string]interface{}{
	// 						{
	// 							"apiVersion":         "apps/v1",
	// 							"kind":               "ReplicaSet",
	// 							"name":               "tomcat01-7dccfcbff8",
	// 							"uid":                "0f48166a-ed7f-4b64-bf15-02ac3b1aa683",
	// 							"controller":         true,
	// 							"blockOwnerDeletion": true,
	// 						},
	// 					},
	// 					"managedFields": []map[string]interface{}{
	// 						{
	// 							"manager":    "kube-controller-manager",
	// 							"operation":  "Update",
	// 							"apiVersion": "v1",
	// 							"time":       "2025-06-28T12:05:19Z",
	// 							"fieldsType": "FieldsV1",
	// 							"fieldsV1":   map[string]interface{}{ /* 省略，内容较多 */ },
	// 						},
	// 						{
	// 							"manager":     "kubelet",
	// 							"operation":   "Update",
	// 							"apiVersion":  "v1",
	// 							"time":        "2025-06-28T12:05:21Z",
	// 							"fieldsType":  "FieldsV1",
	// 							"fieldsV1":    map[string]interface{}{ /* 省略，内容较多 */ },
	// 							"subresource": "status",
	// 						},
	// 					},
	// 				},
	// 				"spec": map[string]interface{}{
	// 					"volumes": []map[string]interface{}{
	// 						{
	// 							"name": "kube-api-access-vd4n5",
	// 							"projected": map[string]interface{}{
	// 								"sources": []map[string]interface{}{
	// 									{
	// 										"serviceAccountToken": map[string]interface{}{
	// 											"expirationSeconds": 3607,
	// 											"path":              "token",
	// 										},
	// 									},
	// 									{
	// 										"configMap": map[string]interface{}{
	// 											"name": "kube-root-ca.crt",
	// 											"items": []map[string]interface{}{
	// 												{
	// 													"key":  "ca.crt",
	// 													"path": "ca.crt",
	// 												},
	// 											},
	// 										},
	// 									},
	// 									{
	// 										"downwardAPI": map[string]interface{}{
	// 											"items": []map[string]interface{}{
	// 												{
	// 													"path": "namespace",
	// 													"fieldRef": map[string]interface{}{
	// 														"apiVersion": "v1",
	// 														"fieldPath":  "metadata.namespace",
	// 													},
	// 												},
	// 											},
	// 										},
	// 									},
	// 								},
	// 								"defaultMode": 420,
	// 							},
	// 						},
	// 					},
	// 					"containers": []map[string]interface{}{
	// 						{
	// 							"name":  "tomcat01",
	// 							"image": "tomcat:latest",
	// 							"ports": []map[string]interface{}{
	// 								{
	// 									"containerPort": 8080,
	// 									"protocol":      "TCP",
	// 								},
	// 							},
	// 							"resources": map[string]interface{}{},
	// 							"volumeMounts": []map[string]interface{}{
	// 								{
	// 									"name":      "kube-api-access-vd4n5",
	// 									"readOnly":  true,
	// 									"mountPath": "/var/run/secrets/kubernetes.io/serviceaccount",
	// 								},
	// 							},
	// 							"terminationMessagePath":   "/dev/termination-log",
	// 							"terminationMessagePolicy": "File",
	// 							"imagePullPolicy":          "Never",
	// 							"securityContext": map[string]interface{}{
	// 								"privileged": true,
	// 							},
	// 						},
	// 					},
	// 					"restartPolicy":                 "Always",
	// 					"terminationGracePeriodSeconds": 30,
	// 					"dnsPolicy":                     "ClusterFirst",
	// 					"serviceAccountName":            "lisi",
	// 					"serviceAccount":                "lisi",
	// 					"nodeName":                      "node2",
	// 					"securityContext":               map[string]interface{}{},
	// 					"schedulerName":                 "default-scheduler",
	// 					"tolerations": []map[string]interface{}{
	// 						{
	// 							"key":               "node.kubernetes.io/not-ready",
	// 							"operator":          "Exists",
	// 							"effect":            "NoExecute",
	// 							"tolerationSeconds": 300,
	// 						},
	// 						{
	// 							"key":               "node.kubernetes.io/unreachable",
	// 							"operator":          "Exists",
	// 							"effect":            "NoExecute",
	// 							"tolerationSeconds": 300,
	// 						},
	// 					},
	// 					"priority":           0,
	// 					"enableServiceLinks": true,
	// 					"preemptionPolicy":   "PreemptLowerPriority",
	// 				},
	// 				"status": map[string]interface{}{
	// 					"phase": "Running",
	// 					"conditions": []map[string]interface{}{
	// 						{
	// 							"type":               "Initialized",
	// 							"status":             "True",
	// 							"lastProbeTime":      nil,
	// 							"lastTransitionTime": "2025-06-28T12:05:19Z",
	// 						},
	// 						{
	// 							"type":               "Ready",
	// 							"status":             "True",
	// 							"lastProbeTime":      nil,
	// 							"lastTransitionTime": "2025-06-28T12:05:21Z",
	// 						},
	// 						{
	// 							"type":               "ContainersReady",
	// 							"status":             "True",
	// 							"lastProbeTime":      nil,
	// 							"lastTransitionTime": "2025-06-28T12:05:21Z",
	// 						},
	// 						{
	// 							"type":               "PodScheduled",
	// 							"status":             "True",
	// 							"lastProbeTime":      nil,
	// 							"lastTransitionTime": "2025-06-28T12:05:19Z",
	// 						},
	// 					},
	// 					"hostIP": "10.211.55.7",
	// 					"podIP":  "10.244.1.108",
	// 					"podIPs": []map[string]interface{}{
	// 						{
	// 							"ip": "10.244.1.108",
	// 						},
	// 					},
	// 					"startTime": "2025-06-28T12:05:19Z",
	// 					"containerStatuses": []map[string]interface{}{
	// 						{
	// 							"name": "tomcat01",
	// 							"state": map[string]interface{}{
	// 								"running": map[string]interface{}{
	// 									"startedAt": "2025-06-28T12:05:21Z",
	// 								},
	// 							},
	// 							"lastState":    map[string]interface{}{},
	// 							"ready":        true,
	// 							"restartCount": 0,
	// 							"image":        "tomcat:latest",
	// 							"imageID":      "docker-pullable://tomcat@sha256:1374a565d5122fdb42807f3a5f2d4fcc245a5e15420ff5bb5123afedc8ef769d",
	// 							"containerID":  "docker://8ae66327906ed63219d6885c556893ed42956eb4b50d8c8486a5c858344304b6",
	// 							"started":      true,
	// 						},
	// 					},
	// 					"qosClass": "BestEffort",
	// 				},
	// 			},
	// 		},
	// 	},
	// }
	// 初始化 namespacePodData 结构体（适配 kubectl 表格显示）
	var namespacePodData = map[string]interface{}{
		"default": map[string]interface{}{
			"kind":       "Table",
			"apiVersion": "meta.k8s.io/v1",
			"metadata": map[string]interface{}{
				"resourceVersion": "20529619",
			},
			"columnDefinitions": []map[string]interface{}{
				{"name": "NAME", "type": "string", "format": "name"},
				{"name": "READY", "type": "string"},
				{"name": "STATUS", "type": "string"},
				{"name": "RESTARTS", "type": "string"},
				{"name": "AGE", "type": "string"},
			},
			"rows": []map[string]interface{}{
				{
					"cells": []interface{}{
						"tomcat01-7dccfcbff8-qdgns",
						"1/1",
						"Running",
						"0",
						"117m",
					},
					"object": map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"name":              "tomcat01-7dccfcbff8-qdgns",
							"namespace":         "default",
							"creationTimestamp": "2025-06-28T12:05:19Z",
						},
						"status": map[string]interface{}{
							"phase": "Running",
							"containerStatuses": []map[string]interface{}{
								{
									"name":  "tomcat01",
									"ready": true,
									"state": map[string]interface{}{
										"running": map[string]interface{}{},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	switch {
	// 处理命名空间Pod列表请求
	case strings.HasPrefix(path, "/api/v1/namespaces/") && strings.HasSuffix(path, "/pods"):
		{
			// 提取命名空间名称
			namespace := strings.TrimPrefix(path, "/api/v1/namespaces/")
			namespace = strings.TrimSuffix(namespace, "/pods")

			// 检查命名空间是否存在
			if data, exists := namespacePodData[namespace]; exists {
				return data
			}

			// 为不存在的命名空间返回空Pod列表
			return map[string]interface{}{
				"kind":       "PodList",
				"apiVersion": "v1",
				"metadata": map[string]interface{}{
					"resourceVersion": "20523975",
				},
				"items": []map[string]interface{}{},
			}
		}
	case path == "/api/v1/pods?limit=500":
		{
			return map[string]interface{}{
				"kind":       "PodList",
				"apiVersion": "v1",
				"metadata": map[string]interface{}{
					"selfLink":        "/api/v1/pods",
					"resourceVersion": "123456",
					"continue":        "",
				},
				"items": []map[string]interface{}{
					{
						"metadata": map[string]interface{}{
							"name":              "nginx-78d6c9b7d4-5j6k7",
							"generateName":      "nginx-78d6c9b7d4-",
							"namespace":         "default",
							"selfLink":          "/api/v1/namespaces/default/pods/nginx-78d6c9b7d4-5j6k7",
							"uid":               "6d9f4e3c-1e2f-11ea-8d71-0242ac130003",
							"resourceVersion":   "12345",
							"creationTimestamp": "2025-06-26T08:30:00Z",
							"labels": map[string]string{
								"app":  "nginx",
								"tier": "web",
							},
							"annotations": map[string]string{
								"kubernetes.io/limit-ranger": "LimitRanger plugin set: cpu request for container nginx",
							},
							"ownerReferences": []map[string]interface{}{
								{
									"apiVersion":         "apps/v1",
									"kind":               "ReplicaSet",
									"name":               "nginx-78d6c9b7d4",
									"uid":                "6d9d9e3c-1e2f-11ea-8d71-0242ac130003",
									"controller":         true,
									"blockOwnerDeletion": true,
								},
							},
						},
						"spec": map[string]interface{}{
							"volumes": []map[string]interface{}{
								{
									"name": "default-token-5q7c9",
									"secret": map[string]interface{}{
										"secretName":  "default-token-5q7c9",
										"defaultMode": 420,
									},
								},
							},
							"containers": []map[string]interface{}{
								{
									"name":  "nginx",
									"image": "nginx:1.23",
									"ports": []map[string]interface{}{
										{
											"containerPort": 80,
											"protocol":      "TCP",
										},
									},
									"resources": map[string]interface{}{
										"requests": map[string]interface{}{
											"cpu": "100m",
										},
									},
									"volumeMounts": []map[string]interface{}{
										{
											"name":      "default-token-5q7c9",
											"readOnly":  true,
											"mountPath": "/var/run/secrets/kubernetes.io/serviceaccount",
										},
									},
									"terminationMessagePath":   "/dev/termination-log",
									"terminationMessagePolicy": "File",
									"imagePullPolicy":          "Always",
								},
							},
							"restartPolicy":                 "Always",
							"terminationGracePeriodSeconds": 30,
							"dnsPolicy":                     "ClusterFirst",
							"serviceAccountName":            "default",
							"serviceAccount":                "default",
							"nodeName":                      "node1",
							"securityContext":               map[string]interface{}{},
							"schedulerName":                 "default-scheduler",
							"tolerations": []map[string]interface{}{
								{
									"key":               "node.kubernetes.io/not-ready",
									"operator":          "Exists",
									"effect":            "NoExecute",
									"tolerationSeconds": 300,
								},
								{
									"key":               "node.kubernetes.io/unreachable",
									"operator":          "Exists",
									"effect":            "NoExecute",
									"tolerationSeconds": 300,
								},
							},
						},
						"status": map[string]interface{}{
							"phase": "Running",
							"conditions": []map[string]interface{}{
								{
									"type":               "Initialized",
									"status":             "True",
									"lastProbeTime":      nil,
									"lastTransitionTime": "2025-06-26T08:30:00Z",
								},
								{
									"type":               "Ready",
									"status":             "True",
									"lastProbeTime":      nil,
									"lastTransitionTime": "2025-06-26T08:32:00Z",
								},
								{
									"type":               "ContainersReady",
									"status":             "True",
									"lastProbeTime":      nil,
									"lastTransitionTime": "2025-06-26T08:32:00Z",
								},
								{
									"type":               "PodScheduled",
									"status":             "True",
									"lastProbeTime":      nil,
									"lastTransitionTime": "2025-06-26T08:30:00Z",
								},
							},
							"hostIP": "192.168.1.101",
							"podIP":  "10.244.1.5",
							"podIPs": []map[string]interface{}{
								{
									"ip": "10.244.1.5",
								},
							},
							"startTime": "2025-06-26T08:30:00Z",
							"containerStatuses": []map[string]interface{}{
								{
									"name": "nginx",
									"state": map[string]interface{}{
										"running": map[string]interface{}{
											"startedAt": "2025-06-26T08:32:00Z",
										},
									},
									"lastState":    map[string]interface{}{},
									"ready":        true,
									"restartCount": 0,
									"image":        "nginx:1.23",
									"imageID":      "docker-pullable://nginx@sha256:1234567890abcdef1234567890abcdef",
									"containerID":  "docker://1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
									"started":      true,
								},
							},
							"qosClass": "Burstable",
						},
					},
					// 其他 Pod 类似结构，此处省略...
				},
			}
		}
	// 处理 /api/v1 路径
	case path == "/api/v1":
		{
			return map[string]interface{}{
				"kind":         "APIResourceList",
				"groupVersion": "v1",
				"resources": []map[string]interface{}{
					{
						"name":         "bindings",
						"singularName": "",
						"namespaced":   true,
						"kind":         "Binding",
						"verbs":        []string{"create"},
					},
					{
						"name":         "componentstatuses",
						"singularName": "",
						"namespaced":   false,
						"kind":         "ComponentStatus",
						"verbs":        []string{"get", "list"},
						"shortNames":   []string{"cs"},
					},
					{
						"name":         "configmaps",
						"singularName": "",
						"namespaced":   true,
						"kind":         "ConfigMap",
						"verbs": []string{
							"create", "delete", "deletecollection", "get", "list", "patch", "update", "watch",
						},
						"shortNames":         []string{"cm"},
						"storageVersionHash": "qFsyl6wFWjQ=",
					},
					{
						"name":         "endpoints",
						"singularName": "",
						"namespaced":   true,
						"kind":         "Endpoints",
						"verbs": []string{
							"create", "delete", "deletecollection", "get", "list", "patch", "update", "watch",
						},
						"shortNames":         []string{"ep"},
						"storageVersionHash": "fWeeMqaN/OA=",
					},
					{
						"name":         "events",
						"singularName": "",
						"namespaced":   true,
						"kind":         "Event",
						"verbs": []string{
							"create", "delete", "deletecollection", "get", "list", "patch", "update", "watch",
						},
						"shortNames":         []string{"ev"},
						"storageVersionHash": "r2yiGXH7wu8=",
					},
					{
						"name":         "limitranges",
						"singularName": "",
						"namespaced":   true,
						"kind":         "LimitRange",
						"verbs": []string{
							"create", "delete", "deletecollection", "get", "list", "patch", "update", "watch",
						},
						"shortNames":         []string{"limits"},
						"storageVersionHash": "EBKMFVe6cwo=",
					},
					{
						"name":         "namespaces",
						"singularName": "",
						"namespaced":   false,
						"kind":         "Namespace",
						"verbs": []string{
							"create", "delete", "get", "list", "patch", "update", "watch",
						},
						"shortNames":         []string{"ns"},
						"storageVersionHash": "Q3oi5N2YM8M=",
					},
					{
						"name":         "namespaces/finalize",
						"singularName": "",
						"namespaced":   false,
						"kind":         "Namespace",
						"verbs":        []string{"update"},
					},
					{
						"name":         "namespaces/status",
						"singularName": "",
						"namespaced":   false,
						"kind":         "Namespace",
						"verbs":        []string{"get", "patch", "update"},
					},
					{
						"name":         "nodes",
						"singularName": "",
						"namespaced":   false,
						"kind":         "Node",
						"verbs": []string{
							"create", "delete", "deletecollection", "get", "list", "patch", "update", "watch",
						},
						"shortNames":         []string{"no"},
						"storageVersionHash": "XwShjMxG9Fs=",
					},
					{
						"name":         "nodes/proxy",
						"singularName": "",
						"namespaced":   false,
						"kind":         "NodeProxyOptions",
						"verbs":        []string{"create", "delete", "get", "patch", "update"},
					},
					{
						"name":         "nodes/status",
						"singularName": "",
						"namespaced":   false,
						"kind":         "Node",
						"verbs":        []string{"get", "patch", "update"},
					},
					{
						"name":         "persistentvolumeclaims",
						"singularName": "",
						"namespaced":   true,
						"kind":         "PersistentVolumeClaim",
						"verbs": []string{
							"create", "delete", "deletecollection", "get", "list", "patch", "update", "watch",
						},
						"shortNames":         []string{"pvc"},
						"storageVersionHash": "QWTyNDq0dC4=",
					},
					{
						"name":         "persistentvolumeclaims/status",
						"singularName": "",
						"namespaced":   true,
						"kind":         "PersistentVolumeClaim",
						"verbs":        []string{"get", "patch", "update"},
					},
					{
						"name":         "persistentvolumes",
						"singularName": "",
						"namespaced":   false,
						"kind":         "PersistentVolume",
						"verbs": []string{
							"create", "delete", "deletecollection", "get", "list", "patch", "update", "watch",
						},
						"shortNames":         []string{"pv"},
						"storageVersionHash": "HN/zwEC+JgM=",
					},
					{
						"name":         "persistentvolumes/status",
						"singularName": "",
						"namespaced":   false,
						"kind":         "PersistentVolume",
						"verbs":        []string{"get", "patch", "update"},
					},
					{
						"name":         "pods",
						"singularName": "",
						"namespaced":   true,
						"kind":         "Pod",
						"verbs": []string{
							"create", "delete", "deletecollection", "get", "list", "patch", "update", "watch",
						},
						"shortNames":         []string{"po"},
						"categories":         []string{"all"},
						"storageVersionHash": "xPOwRZ+Yhw8=",
					},
					{
						"name":         "pods/attach",
						"singularName": "",
						"namespaced":   true,
						"kind":         "PodAttachOptions",
						"verbs":        []string{"create", "get"},
					},
					{
						"name":         "pods/binding",
						"singularName": "",
						"namespaced":   true,
						"kind":         "Binding",
						"verbs":        []string{"create"},
					},
					{
						"name":         "pods/ephemeralcontainers",
						"singularName": "",
						"namespaced":   true,
						"kind":         "Pod",
						"verbs":        []string{"get", "patch", "update"},
					},
					{
						"name":         "pods/eviction",
						"singularName": "",
						"namespaced":   true,
						"group":        "policy",
						"version":      "v1",
						"kind":         "Eviction",
						"verbs":        []string{"create"},
					},
					{
						"name":         "pods/exec",
						"singularName": "",
						"namespaced":   true,
						"kind":         "PodExecOptions",
						"verbs":        []string{"create", "get"},
					},
					{
						"name":         "pods/log",
						"singularName": "",
						"namespaced":   true,
						"kind":         "Pod",
						"verbs":        []string{"get"},
					},
					{
						"name":         "pods/portforward",
						"singularName": "",
						"namespaced":   true,
						"kind":         "PodPortForwardOptions",
						"verbs":        []string{"create", "get"},
					},
					{
						"name":         "pods/proxy",
						"singularName": "",
						"namespaced":   true,
						"kind":         "PodProxyOptions",
						"verbs":        []string{"create", "delete", "get", "patch", "update"},
					},
					{
						"name":         "pods/status",
						"singularName": "",
						"namespaced":   true,
						"kind":         "Pod",
						"verbs":        []string{"get", "patch", "update"},
					},
					{
						"name":         "podtemplates",
						"singularName": "",
						"namespaced":   true,
						"kind":         "PodTemplate",
						"verbs": []string{
							"create", "delete", "deletecollection", "get", "list", "patch", "update", "watch",
						},
						"storageVersionHash": "LIXB2x4IFpk=",
					},
					{
						"name":         "replicationcontrollers",
						"singularName": "",
						"namespaced":   true,
						"kind":         "ReplicationController",
						"verbs": []string{
							"create", "delete", "deletecollection", "get", "list", "patch", "update", "watch",
						},
						"shortNames":         []string{"rc"},
						"categories":         []string{"all"},
						"storageVersionHash": "Jond2If31h0=",
					},
					{
						"name":         "replicationcontrollers/scale",
						"singularName": "",
						"namespaced":   true,
						"group":        "autoscaling",
						"version":      "v1",
						"kind":         "Scale",
						"verbs":        []string{"get", "patch", "update"},
					},
					{
						"name":         "replicationcontrollers/status",
						"singularName": "",
						"namespaced":   true,
						"kind":         "ReplicationController",
						"verbs":        []string{"get", "patch", "update"},
					},
					{
						"name":         "resourcequotas",
						"singularName": "",
						"namespaced":   true,
						"kind":         "ResourceQuota",
						"verbs": []string{
							"create", "delete", "deletecollection", "get", "list", "patch", "update", "watch",
						},
						"shortNames":         []string{"quota"},
						"storageVersionHash": "8uhSgffRX6w=",
					},
					{
						"name":         "resourcequotas/status",
						"singularName": "",
						"namespaced":   true,
						"kind":         "ResourceQuota",
						"verbs":        []string{"get", "patch", "update"},
					},
					{
						"name":         "secrets",
						"singularName": "",
						"namespaced":   true,
						"kind":         "Secret",
						"verbs": []string{
							"create", "delete", "deletecollection", "get", "list", "patch", "update", "watch",
						},
						"storageVersionHash": "S6u1pOWzb84=",
					},
					{
						"name":         "serviceaccounts",
						"singularName": "",
						"namespaced":   true,
						"kind":         "ServiceAccount",
						"verbs": []string{
							"create", "delete", "deletecollection", "get", "list", "patch", "update", "watch",
						},
						"shortNames":         []string{"sa"},
						"storageVersionHash": "pbx9ZvyFpBE=",
					},
					{
						"name":         "serviceaccounts/token",
						"singularName": "",
						"namespaced":   true,
						"group":        "authentication.k8s.io",
						"version":      "v1",
						"kind":         "TokenRequest",
						"verbs":        []string{"create"},
					},
					{
						"name":         "services",
						"singularName": "",
						"namespaced":   true,
						"kind":         "Service",
						"verbs": []string{
							"create", "delete", "deletecollection", "get", "list", "patch", "update", "watch",
						},
						"shortNames":         []string{"svc"},
						"categories":         []string{"all"},
						"storageVersionHash": "0/CO1lhkEBI=",
					},
					{
						"name":         "services/proxy",
						"singularName": "",
						"namespaced":   true,
						"kind":         "ServiceProxyOptions",
						"verbs":        []string{"create", "delete", "get", "patch", "update"},
					},
					{
						"name":         "services/status",
						"singularName": "",
						"namespaced":   true,
						"kind":         "Service",
						"verbs":        []string{"get", "patch", "update"},
					},
				},
			}
		}

	case path == "/apis/apiextensions.k8s.io":
		{
			return map[string]interface{}{
				"kind":       "APIGroup",
				"apiVersion": "v1",
				"name":       "apiextensions.k8s.io",
				"versions": []map[string]interface{}{
					{
						"groupVersion": "apiextensions.k8s.io/v1",
						"version":      "v1",
					},
				},
				"preferredVersion": map[string]string{
					"groupVersion": "apiextensions.k8s.io/v1",
					"version":      "v1",
				},
			}
		}
	case path == "/apis/admissionregistration.k8s.io/v1":
		{
			return map[string]interface{}{
				"kind":         "APIResourceList",
				"apiVersion":   "v1",
				"groupVersion": "admissionregistration.k8s.io/v1",
				"resources": []map[string]interface{}{
					{
						"name":         "mutatingwebhookconfigurations",
						"singularName": "",
						"namespaced":   false,
						"kind":         "MutatingWebhookConfiguration",
						"verbs": []string{
							"create", "delete", "deletecollection", "get", "list", "patch", "update", "watch",
						},
						"categories":         []string{"api-extensions"},
						"storageVersionHash": "Sqi0GUgDaX0=",
					},
					{
						"name":         "validatingwebhookconfigurations",
						"singularName": "",
						"namespaced":   false,
						"kind":         "ValidatingWebhookConfiguration",
						"verbs": []string{
							"create", "delete", "deletecollection", "get", "list", "patch", "update", "watch",
						},
						"categories":         []string{"api-extensions"},
						"storageVersionHash": "B0wHjQmsGNk=",
					},
				},
			}
		}

	case path == "/api":
		{
			return map[string]interface{}{
				"kind":     "APIVersions",
				"versions": []string{"v1"},
				"serverAddressByClientCIDRs": []map[string]string{
					{"clientCIDR": "0.0.0.0/0", "serverAddress": "127.0.0.1:6443"},
				},
			}
		}

	case path == "/.well-known/openid-configuration":
		{
			return map[string]interface{}{
				"issuer": "https://kubernetes.default.svc.cluster.local",
				//"jwks_uri":                              "https://10.211.55.6:6443/openid/v1/jwks",
				"response_types_supported":              []string{"id_token"},
				"subject_types_supported":               []string{"public"},
				"id_token_signing_alg_values_supported": []string{"RS256"},
			}
		}
	case path == "/apis/apps/v1":
		{
			return map[string]interface{}{
				"kind":       "APIGroup",
				"apiVersion": "v1",
				"name":       "apiextensions.k8s.io",
				"versions": []map[string]interface{}{
					{
						"groupVersion": "apiextensions.k8s.io/v1",
						"version":      "v1",
					},
				},
				"preferredVersion": map[string]string{
					"groupVersion": "apiextensions.k8s.io/v1",
					"version":      "v1",
				},
			}
		}
	case path == "/apis/apps":
		{
			return map[string]interface{}{
				"kind":       "APIGroup",
				"apiVersion": "v1",
				"name":       "apps",
				"versions": []map[string]interface{}{
					{
						"groupVersion": "apps/v1",
						"version":      "v1",
					},
				},
				"preferredVersion": map[string]string{
					"groupVersion": "apps/v1",
					"version":      "v1",
				},
			}
		}
	case path == "/apis/apiregistration.k8s.io/v1":
		{
			return map[string]interface{}{
				"kind":         "APIResourceList",
				"apiVersion":   "v1",
				"groupVersion": "apiregistration.k8s.io/v1",
				"resources": []map[string]interface{}{
					{
						"name":         "apiservices",
						"singularName": "",
						"namespaced":   false,
						"kind":         "APIService",
						"verbs": []string{
							"create", "delete", "deletecollection", "get", "list", "patch", "update", "watch",
						},
						"categories":         []string{"api-extensions"},
						"storageVersionHash": "InPBPD7+PqM=",
					},
					{
						"name":         "apiservices/status",
						"singularName": "",
						"namespaced":   false,
						"kind":         "APIService",
						"verbs":        []string{"get", "patch", "update"},
					},
				},
			}
		}
	// 新增：处理 /apis/certificates.k8s.io/v1 路径
	case path == "/apis/certificates.k8s.io/v1":
		{
			return map[string]interface{}{
				"kind":         "APIResourceList",
				"apiVersion":   "v1",
				"groupVersion": "certificates.k8s.io/v1",
				"resources": []map[string]interface{}{
					{
						"name":         "certificatesigningrequests",
						"singularName": "",
						"namespaced":   false,
						"kind":         "CertificateSigningRequest",
						"verbs": []string{
							"create", "delete", "deletecollection", "get", "list", "patch", "update", "watch",
						},
						"shortNames":         []string{"csr"},
						"storageVersionHash": "95fRKMXA+00=",
					},
					{
						"name":         "certificatesigningrequests/approval",
						"singularName": "",
						"namespaced":   false,
						"kind":         "CertificateSigningRequest",
						"verbs":        []string{"get", "patch", "update"},
					},
					{
						"name":         "certificatesigningrequests/status",
						"singularName": "",
						"namespaced":   false,
						"kind":         "CertificateSigningRequest",
						"verbs":        []string{"get", "patch", "update"},
					},
				},
			}
		}
	case path == "/apis/certificates.k8s.io":
		{
			return map[string]interface{}{
				"kind":       "APIGroup",
				"apiVersion": "v1",
				"name":       "certificates.k8s.io",
				"versions": []map[string]interface{}{
					{
						"groupVersion": "certificates.k8s.io/v1",
						"version":      "v1",
					},
				},
				"preferredVersion": map[string]string{
					"groupVersion": "certificates.k8s.io/v1",
					"version":      "v1",
				},
			}
		}
	case path == "/apis/batch/v1beta1":
		{
			return map[string]interface{}{
				"kind":         "APIResourceList",
				"apiVersion":   "v1",
				"groupVersion": "batch/v1beta1",
				"resources": []map[string]interface{}{
					{
						"name":         "cronjobs",
						"singularName": "",
						"namespaced":   true,
						"kind":         "CronJob",
						"verbs": []string{
							"create", "delete", "deletecollection", "get", "list", "patch", "update", "watch",
						},
						"shortNames":         []string{"cj"},
						"categories":         []string{"all"},
						"storageVersionHash": "sd5LIXh4Fjs=",
					},
					{
						"name":         "cronjobs/status",
						"singularName": "",
						"namespaced":   true,
						"kind":         "CronJob",
						"verbs":        []string{"get", "patch", "update"},
					},
				},
			}
		}
	case path == "/apis/batch/v1":
		{
			return map[string]interface{}{
				"kind":         "APIResourceList",
				"apiVersion":   "v1",
				"groupVersion": "batch/v1",
				"resources": []map[string]interface{}{
					{
						"name":         "cronjobs",
						"singularName": "",
						"namespaced":   true,
						"kind":         "CronJob",
						"verbs": []string{
							"create", "delete", "deletecollection", "get", "list", "patch", "update", "watch",
						},
						"shortNames":         []string{"cj"},
						"categories":         []string{"all"},
						"storageVersionHash": "sd5LIXh4Fjs=",
					},
					{
						"name":         "cronjobs/status",
						"singularName": "",
						"namespaced":   true,
						"kind":         "CronJob",
						"verbs":        []string{"get", "patch", "update"},
					},
					{
						"name":         "jobs",
						"singularName": "",
						"namespaced":   true,
						"kind":         "Job",
						"verbs": []string{
							"create", "delete", "deletecollection", "get", "list", "patch", "update", "watch",
						},
						"categories":         []string{"all"},
						"storageVersionHash": "mudhfqk/qZY=",
					},
					{
						"name":         "jobs/status",
						"singularName": "",
						"namespaced":   true,
						"kind":         "Job",
						"verbs":        []string{"get", "patch", "update"},
					},
				},
			}
		}

	case path == "/":
		return map[string]interface{}{
			"paths": []string{
				"/.well-known/openid-configuration",
				"/api",
				"/api/v1",
				"/apis",
				"/apis/",
				"/apis/admissionregistration.k8s.io",
				"/apis/admissionregistration.k8s.io/v1",
				"/apis/apiextensions.k8s.io",
				"/apis/apiextensions.k8s.io/v1",
				"/apis/apiregistration.k8s.io",
				"/apis/apiregistration.k8s.io/v1",
				"/apis/apps",
				"/apis/apps/v1",
				"/apis/authentication.k8s.io",
				"/apis/authentication.k8s.io/v1",
				"/apis/authorization.k8s.io",
				"/apis/authorization.k8s.io/v1",
				"/apis/autoscaling",
				"/apis/autoscaling/v1",
				"/apis/autoscaling/v2",
				"/apis/autoscaling/v2beta1",
				"/apis/autoscaling/v2beta2",
				"/apis/batch",
				"/apis/batch/v1",
				"/apis/batch/v1beta1",
				"/apis/certificates.k8s.io",
				"/apis/certificates.k8s.io/v1",
				"/apis/coordination.k8s.io",
				"/apis/coordination.k8s.io/v1",
				"/apis/discovery.k8s.io",
				"/apis/discovery.k8s.io/v1",
				"/apis/discovery.k8s.io/v1beta1",
				"/apis/events.k8s.io",
				"/apis/events.k8s.io/v1",
				"/apis/events.k8s.io/v1beta1",
				"/apis/flowcontrol.apiserver.k8s.io",
				"/apis/flowcontrol.apiserver.k8s.io/v1beta1",
				"/apis/flowcontrol.apiserver.k8s.io/v1beta2",
				"/apis/networking.k8s.io",
				"/apis/networking.k8s.io/v1",
				"/apis/node.k8s.io",
				"/apis/node.k8s.io/v1",
				"/apis/node.k8s.io/v1beta1",
				"/apis/policy",
				"/apis/policy/v1",
				"/apis/policy/v1beta1",
				"/apis/rbac.authorization.k8s.io",
				"/apis/rbac.authorization.k8s.io/v1",
				"/apis/scheduling.k8s.io",
				"/apis/scheduling.k8s.io/v1",
				"/apis/storage.k8s.io",
				"/apis/storage.k8s.io/v1",
				"/apis/storage.k8s.io/v1beta1",
				"/healthz",
				"/healthz/autoregister-completion",
				"/healthz/etcd",
				"/healthz/log",
				"/healthz/ping",
				"/healthz/poststarthook/aggregator-reload-proxy-client-cert",
				"/healthz/poststarthook/apiservice-openapi-controller",
				"/healthz/poststarthook/apiservice-registration-controller",
				"/healthz/poststarthook/apiservice-status-available-controller",
				"/healthz/poststarthook/bootstrap-controller",
				"/healthz/poststarthook/crd-informer-synced",
				"/healthz/poststarthook/generic-apiserver-start-informers",
				"/healthz/poststarthook/kube-apiserver-autoregistration",
				"/healthz/poststarthook/priority-and-fairness-config-consumer",
				"/healthz/poststarthook/priority-and-fairness-config-producer",
				"/healthz/poststarthook/priority-and-fairness-filter",
				"/healthz/poststarthook/rbac/bootstrap-roles",
				"/healthz/poststarthook/scheduling/bootstrap-system-priority-classes",
				"/healthz/poststarthook/start-apiextensions-controllers",
				"/healthz/poststarthook/start-apiextensions-informers",
				"/healthz/poststarthook/start-cluster-authentication-info-controller",
				"/healthz/poststarthook/start-kube-aggregator-informers",
				"/healthz/poststarthook/start-kube-apiserver-admission-initializer",
				"/livez",
				"/livez/autoregister-completion",
				"/livez/etcd",
				"/livez/log",
				"/livez/ping",
				"/livez/poststarthook/aggregator-reload-proxy-client-cert",
				"/livez/poststarthook/apiservice-openapi-controller",
				"/livez/poststarthook/apiservice-registration-controller",
				"/livez/poststarthook/apiservice-status-available-controller",
				"/livez/poststarthook/bootstrap-controller",
				"/livez/poststarthook/crd-informer-synced",
				"/livez/poststarthook/generic-apiserver-start-informers",
				"/livez/poststarthook/kube-apiserver-autoregistration",
				"/livez/poststarthook/priority-and-fairness-config-consumer",
				"/livez/poststarthook/priority-and-fairness-config-producer",
				"/livez/poststarthook/priority-and-fairness-filter",
				"/livez/poststarthook/rbac/bootstrap-roles",
				"/livez/poststarthook/scheduling/bootstrap-system-priority-classes",
				"/livez/poststarthook/start-apiextensions-controllers",
				"/livez/poststarthook/start-apiextensions-informers",
				"/livez/poststarthook/start-cluster-authentication-info-controller",
				"/livez/poststarthook/start-kube-aggregator-informers",
				"/livez/poststarthook/start-kube-apiserver-admission-initializer",
				"/logs",
				"/metrics",
				"/openapi/v2",
				"/openid/v1/jwks",
				"/readyz",
				"/readyz/autoregister-completion",
				"/readyz/etcd",
				"/readyz/informer-sync",
				"/readyz/log",
				"/readyz/ping",
				"/readyz/poststarthook/aggregator-reload-proxy-client-cert",
				"/readyz/poststarthook/apiservice-openapi-controller",
				"/readyz/poststarthook/apiservice-registration-controller",
				"/readyz/poststarthook/apiservice-status-available-controller",
				"/readyz/poststarthook/bootstrap-controller",
				"/readyz/poststarthook/crd-informer-synced",
				"/readyz/poststarthook/generic-apiserver-start-informers",
				"/readyz/poststarthook/kube-apiserver-autoregistration",
				"/readyz/poststarthook/priority-and-fairness-config-consumer",
				"/readyz/poststarthook/priority-and-fairness-config-producer",
				"/readyz/poststarthook/priority-and-fairness-filter",
				"/readyz/poststarthook/rbac/bootstrap-roles",
				"/readyz/poststarthook/scheduling/bootstrap-system-priority-classes",
				"/readyz/poststarthook/start-apiextensions-controllers",
				"/readyz/poststarthook/start-apiextensions-informers",
				"/readyz/poststarthook/start-cluster-authentication-info-controller",
				"/readyz/poststarthook/start-kube-aggregator-informers",
				"/readyz/poststarthook/start-kube-apiserver-admission-initializer",
				"/readyz/shutdown",
				"/version",
			},
		}
	case path == "/healthz":
		{
			return "ok"
		}
	case path == "/healthz/autoregister-completion":
		{
			return "ok"
		}
	case path == "/healthz/poststarthook/aggregator-reload-proxy-client-cert":
		{
			return "ok"
		}
	case path == "/healthz/poststarthook/apiservice-openapi-controller":
		{
			return "ok"
		}
	case path == "/healthz/poststarthook/apiservice-openapi-controller":
		{
			return "ok"
		}
	case path == "/healthz/poststarthook/apiservice-registration-controller":
		{
			return "ok"
		}
	case path == "/healthz/poststarthook/apiservice-registration-controller":
		{
			return "ok"
		}
	case path == "/healthz/poststarthook/apiservice-status-available-controller":
		{
			return "ok"
		}
	case path == "/healthz/poststarthook/bootstrap-controller":
		{
			return "ok"
		}
	case path == "/healthz/poststarthook/crd-informer-synced":
		{
			return "ok"
		}
	case path == "/healthz/poststarthook/generic-apiserver-start-informers":
		{
			return "ok"
		}
	case path == "/healthz/poststarthook/kube-apiserver-autoregistration":
		{
			return "ok"
		}
	case path == "/healthz/poststarthook/priority-and-fairness-config-consumer":
		{
			return "ok"
		}
	case path == "/healthz/poststarthook/priority-and-fairness-config-producer":
		{
			return "ok"
		}
	case path == "/healthz/poststarthook/priority-and-fairness-filter":
		{
			return "ok"
		}
	case path == "/healthz/poststarthook/rbac/bootstrap-roles":
		{
			return "ok"
		}
	case path == "/healthz/poststarthook/scheduling/bootstrap-system-priority-classes":
		{
			return "ok"
		}
	case path == "/healthz/poststarthook/start-apiextensions-controllers":
		{
			return "ok"
		}
	case path == "/healthz/poststarthook/start-apiextensions-informers":
		{
			return "ok"
		}
	case path == "/healthz/poststarthook/start-cluster-authentication-info-controller":
		{
			return "ok"
		}
	case path == "/healthz/poststarthook/start-cluster-authentication-info-controller":
		{
			return "ok"
		}
	case path == "/healthz/poststarthook/start-kube-aggregator-informers":
		{
			return "ok"
		}
	case path == "/healthz/poststarthook/start-kube-apiserver-admission-initializer":
		{
			return "ok"
		}
	case path == "/livez":
		{
			return "ok"
		}
	case path == "/livez/autoregister-completion":
		{
			return "ok"
		}
	case path == "/livez/etcd":
		{
			return "ok"
		}
	case path == "/livez/log":
		{
			return "ok"
		}
	case path == "/livez/ping":
		{
			return "ok"
		}
	case path == "/livez/poststarthook/aggregator-reload-proxy-client-cert":
		{
			return "ok"
		}
	case path == "/livez/poststarthook/apiservice-openapi-controller":
		{
			return "ok"
		}
	case path == "/livez/poststarthook/apiservice-registration-controller":
		{
			return "ok"
		}
	case path == "/livez/poststarthook/apiservice-status-available-controller":
		{
			return "ok"
		}
	case path == "/livez/poststarthook/bootstrap-controller":
		{
			return "ok"
		}
	case path == "/livez/poststarthook/crd-informer-synced":
		{
			return "ok"
		}
	case path == "/livez/poststarthook/generic-apiserver-start-informers":
		{
			return "ok"
		}
	case path == "/livez/poststarthook/kube-apiserver-autoregistration":
		{
			return "ok"
		}
	case path == "/livez/poststarthook/priority-and-fairness-config-consumer":
		{
			return "ok"
		}
	case path == "/livez/poststarthook/priority-and-fairness-config-producer":
		{
			return "ok"
		}
	case path == "/livez/poststarthook/priority-and-fairness-filter":
		{
			return "ok"
		}
	case path == "/livez/poststarthook/rbac/bootstrap-roles":
		{
			return "ok"
		}
	case path == "/livez/poststarthook/scheduling/bootstrap-system-priority-classes":
		{
			return "ok"
		}
	case path == "/livez/poststarthook/start-apiextensions-controllers":
		{
			return "ok"
		}
	case path == "/livez/poststarthook/start-apiextensions-informers":
		{
			return "ok"
		}
	case path == "/livez/poststarthook/start-cluster-authentication-info-controller":
		{
			return "ok"
		}
	case path == "/livez/poststarthook/start-kube-aggregator-informers":
		{
			return "ok"
		}
	case path == "/livez/poststarthook/start-kube-apiserver-admission-initializer":
		{
			return "ok"
		}
	case path == "/logs":
		{
			return "ok"
		}
	case path == "/metrics":
		{
			return "ok"
		}
	case path == "/openapi/v2":
		{
			return "ok"
		}
	case path == "/openid/v1/jwks":
		{
			return "ok"
		}
	case path == "/readyz":
		{
			return "ok"
		}
	case path == "/readyz/autoregister-completion":
		{
			return "ok"
		}
	case path == "/readyz/etcd":
		{
			return "ok"
		}
	case path == "/readyz/informer-sync":
		{
			return "ok"
		}
	case path == "/readyz/log":
		{
			return "ok"
		}
	case path == "/readyz/ping":
		{
			return "ok"
		}
	case path == "/readyz/poststarthook/aggregator-reload-proxy-client-cert":
		{
			return "ok"
		}
	case path == "/readyz/poststarthook/apiservice-openapi-controller":
		{
			return "ok"
		}
	case path == "/readyz/poststarthook/apiservice-registration-controller":
		{
			return "ok"
		}
	case path == "/readyz/poststarthook/apiservice-status-available-controller":
		{
			return "ok"
		}
	case path == "/readyz/poststarthook/bootstrap-controller":
		{
			return "ok"
		}
	case path == "/readyz/poststarthook/crd-informer-synced":
		{
			return "ok"
		}
	case path == "/readyz/poststarthook/generic-apiserver-start-informers":
		{
			return "ok"
		}
	case path == "/readyz/poststarthook/kube-apiserver-autoregistration":
		{
			return "ok"
		}
	case path == "/readyz/poststarthook/priority-and-fairness-config-consumer":
		{
			return "ok"
		}
	case path == "/readyz/poststarthook/priority-and-fairness-config-producer":
		{
			return "ok"
		}
	case path == "/readyz/poststarthook/priority-and-fairness-filter":
		{
			return "ok"
		}
	case path == "/readyz/poststarthook/rbac/bootstrap-roles":
		{
			return "ok"
		}
	case path == "/readyz/poststarthook/scheduling/bootstrap-system-priority-classes":
		{
			return "ok"
		}
	case path == "/readyz/poststarthook/rbac/bootstrap-roles":
		{
			return "ok"
		}
	case path == "/readyz/poststarthook/scheduling/bootstrap-system-priority-classes":
		{
			return "ok"
		}
	case path == "/readyz/poststarthook/start-apiextensions-controllers":
		{
			return "ok"
		}
	case path == "/readyz/poststarthook/start-apiextensions-informers":
		{
			return "ok"
		}
	case path == "/readyz/poststarthook/start-cluster-authentication-info-controller":
		{
			return "ok"
		}
	case path == "/readyz/poststarthook/start-kube-aggregator-informers":
		{
			return "ok"
		}
	case path == "/readyz/poststarthook/start-kube-apiserver-admission-initializer":
		{
			return "ok"
		}
	case path == "/readyz/shutdown":
		{
			return "ok"
		}
	case path == "/version":
		{
			return "ok"
		}
	case path == "/healthz/etcd":
		{
			return "ok"
		}
	case path == "/healthz/log":
		{
			return "ok"
		}
	case path == "/healthz/log":
		{
			return "ok"
		}
	case path == "/healthz/ping":
		{
			return "ok"
		}
	case path == "healthz/poststarthook/aggregator-reload-proxy-client-cert":
		{
			return "ok"
		}
	case path == "/apis/storage.k8s.io/v1beta1":
		{
			return map[string]interface{}{
				"kind":         "APIResourceList",
				"apiVersion":   "v1",
				"groupVersion": "storage.k8s.io/v1beta1",
				"resources": []map[string]interface{}{
					{
						"name":         "csistoragecapacities",
						"singularName": "",
						"namespaced":   true,
						"kind":         "CSIStorageCapacity",
						"verbs": []string{
							"create", "delete", "deletecollection", "get", "list", "patch", "update", "watch",
						},
						"storageVersionHash": "4as6MA/kOg0=",
					},
				},
			}
		}
	case path == "/apis/":
		return map[string]interface{}{
			"kind":       "APIGroupList",
			"apiVersion": "v1",
			"groups": []map[string]interface{}{
				{
					"name": "apiregistration.k8s.io",
					"versions": []map[string]interface{}{
						{
							"groupVersion": "apiregistration.k8s.io/v1",
							"version":      "v1",
						},
					},
					"preferredVersion": map[string]string{
						"groupVersion": "apiregistration.k8s.io/v1",
						"version":      "v1",
					},
				},
				{
					"name": "apps",
					"versions": []map[string]interface{}{
						{
							"groupVersion": "apps/v1",
							"version":      "v1",
						},
					},
					"preferredVersion": map[string]string{
						"groupVersion": "apps/v1",
						"version":      "v1",
					},
				},
				{
					"name": "events.k8s.io",
					"versions": []map[string]interface{}{
						{
							"groupVersion": "events.k8s.io/v1",
							"version":      "v1",
						},
						{
							"groupVersion": "events.k8s.io/v1beta1",
							"version":      "v1beta1",
						},
					},
					"preferredVersion": map[string]string{
						"groupVersion": "events.k8s.io/v1",
						"version":      "v1",
					},
				},
				{
					"name": "authentication.k8s.io",
					"versions": []map[string]interface{}{
						{
							"groupVersion": "authentication.k8s.io/v1",
							"version":      "v1",
						},
					},
					"preferredVersion": map[string]string{
						"groupVersion": "authentication.k8s.io/v1",
						"version":      "v1",
					},
				},
				{
					"name": "authorization.k8s.io",
					"versions": []map[string]interface{}{
						{
							"groupVersion": "authorization.k8s.io/v1",
							"version":      "v1",
						},
					},
					"preferredVersion": map[string]string{
						"groupVersion": "authorization.k8s.io/v1",
						"version":      "v1",
					},
				},
				{
					"name": "autoscaling",
					"versions": []map[string]interface{}{
						{
							"groupVersion": "autoscaling/v2",
							"version":      "v2",
						},
						{
							"groupVersion": "autoscaling/v1",
							"version":      "v1",
						},
						{
							"groupVersion": "autoscaling/v2beta1",
							"version":      "v2beta1",
						},
						{
							"groupVersion": "autoscaling/v2beta2",
							"version":      "v2beta2",
						},
					},
					"preferredVersion": map[string]string{
						"groupVersion": "autoscaling/v2",
						"version":      "v2",
					},
				},
				{
					"name": "batch",
					"versions": []map[string]interface{}{
						{
							"groupVersion": "batch/v1",
							"version":      "v1",
						},
						{
							"groupVersion": "batch/v1beta1",
							"version":      "v1beta1",
						},
					},
					"preferredVersion": map[string]string{
						"groupVersion": "batch/v1",
						"version":      "v1",
					},
				},
				{
					"name": "certificates.k8s.io",
					"versions": []map[string]interface{}{
						{
							"groupVersion": "certificates.k8s.io/v1",
							"version":      "v1",
						},
					},
					"preferredVersion": map[string]string{
						"groupVersion": "certificates.k8s.io/v1",
						"version":      "v1",
					},
				},
				{
					"name": "networking.k8s.io",
					"versions": []map[string]interface{}{
						{
							"groupVersion": "networking.k8s.io/v1",
							"version":      "v1",
						},
					},
					"preferredVersion": map[string]string{
						"groupVersion": "networking.k8s.io/v1",
						"version":      "v1",
					},
				},
				{
					"name": "policy",
					"versions": []map[string]interface{}{
						{
							"groupVersion": "policy/v1",
							"version":      "v1",
						},
						{
							"groupVersion": "policy/v1beta1",
							"version":      "v1beta1",
						},
					},
					"preferredVersion": map[string]string{
						"groupVersion": "policy/v1",
						"version":      "v1",
					},
				},
				{
					"name": "rbac.authorization.k8s.io",
					"versions": []map[string]interface{}{
						{
							"groupVersion": "rbac.authorization.k8s.io/v1",
							"version":      "v1",
						},
					},
					"preferredVersion": map[string]string{
						"groupVersion": "rbac.authorization.k8s.io/v1",
						"version":      "v1",
					},
				},
				{
					"name": "storage.k8s.io",
					"versions": []map[string]interface{}{
						{
							"groupVersion": "storage.k8s.io/v1",
							"version":      "v1",
						},
						{
							"groupVersion": "storage.k8s.io/v1beta1",
							"version":      "v1beta1",
						},
					},
					"preferredVersion": map[string]string{
						"groupVersion": "storage.k8s.io/v1",
						"version":      "v1",
					},
				},
				{
					"name": "admissionregistration.k8s.io",
					"versions": []map[string]interface{}{
						{
							"groupVersion": "admissionregistration.k8s.io/v1",
							"version":      "v1",
						},
					},
					"preferredVersion": map[string]string{
						"groupVersion": "admissionregistration.k8s.io/v1",
						"version":      "v1",
					},
				},
				{
					"name": "apiextensions.k8s.io",
					"versions": []map[string]interface{}{
						{
							"groupVersion": "apiextensions.k8s.io/v1",
							"version":      "v1",
						},
					},
					"preferredVersion": map[string]string{
						"groupVersion": "apiextensions.k8s.io/v1",
						"version":      "v1",
					},
				},
				{
					"name": "scheduling.k8s.io",
					"versions": []map[string]interface{}{
						{
							"groupVersion": "scheduling.k8s.io/v1",
							"version":      "v1",
						},
					},
					"preferredVersion": map[string]string{
						"groupVersion": "scheduling.k8s.io/v1",
						"version":      "v1",
					},
				},
				{
					"name": "coordination.k8s.io",
					"versions": []map[string]interface{}{
						{
							"groupVersion": "coordination.k8s.io/v1",
							"version":      "v1",
						},
					},
					"preferredVersion": map[string]string{
						"groupVersion": "coordination.k8s.io/v1",
						"version":      "v1",
					},
				},
				{
					"name": "node.k8s.io",
					"versions": []map[string]interface{}{
						{
							"groupVersion": "node.k8s.io/v1",
							"version":      "v1",
						},
						{
							"groupVersion": "node.k8s.io/v1beta1",
							"version":      "v1beta1",
						},
					},
					"preferredVersion": map[string]string{
						"groupVersion": "node.k8s.io/v1",
						"version":      "v1",
					},
				},
				{
					"name": "discovery.k8s.io",
					"versions": []map[string]interface{}{
						{
							"groupVersion": "discovery.k8s.io/v1",
							"version":      "v1",
						},
						{
							"groupVersion": "discovery.k8s.io/v1beta1",
							"version":      "v1beta1",
						},
					},
					"preferredVersion": map[string]string{
						"groupVersion": "discovery.k8s.io/v1",
						"version":      "v1",
					},
				},
				{
					"name": "flowcontrol.apiserver.k8s.io",
					"versions": []map[string]interface{}{
						{
							"groupVersion": "flowcontrol.apiserver.k8s.io/v1beta2",
							"version":      "v1beta2",
						},
						{
							"groupVersion": "flowcontrol.apiserver.k8s.io/v1beta1",
							"version":      "v1beta1",
						},
					},
					"preferredVersion": map[string]string{
						"groupVersion": "flowcontrol.apiserver.k8s.io/v1beta2",
						"version":      "v1beta2",
					},
				},
			},
		}

	// 新增：处理 /apis/storage.k8s.io/v1 路径
	case path == "/apis/storage.k8s.io/v1":
		{
			return map[string]interface{}{
				"kind":         "APIResourceList",
				"apiVersion":   "v1",
				"groupVersion": "storage.k8s.io/v1",
				"resources": []map[string]interface{}{
					{
						"name":         "csidrivers",
						"singularName": "",
						"namespaced":   false,
						"kind":         "CSIDriver",
						"verbs": []string{
							"create", "delete", "deletecollection", "get", "list", "patch", "update", "watch",
						},
						"storageVersionHash": "hL6j/rwBV5w=",
					},
					{
						"name":         "csinodes",
						"singularName": "",
						"namespaced":   false,
						"kind":         "CSINode",
						"verbs": []string{
							"create", "delete", "deletecollection", "get", "list", "patch", "update", "watch",
						},
						"storageVersionHash": "Pe62DkZtjuo=",
					},
					{
						"name":         "storageclasses",
						"singularName": "",
						"namespaced":   false,
						"kind":         "StorageClass",
						"verbs": []string{
							"create", "delete", "deletecollection", "get", "list", "patch", "update", "watch",
						},
						"shortNames":         []string{"sc"},
						"storageVersionHash": "K+m6uJwbjGY=",
					},
					{
						"name":         "volumeattachments",
						"singularName": "",
						"namespaced":   false,
						"kind":         "VolumeAttachment",
						"verbs": []string{
							"create", "delete", "deletecollection", "get", "list", "patch", "update", "watch",
						},
						"storageVersionHash": "tJx/ezt6UDU=",
					},
					{
						"name":         "volumeattachments/status",
						"singularName": "",
						"namespaced":   false,
						"kind":         "VolumeAttachment",
						"verbs":        []string{"get", "patch", "update"},
					},
				},
			}
		}
	case path == "/apis/storage.k8s.io":
		{
			return map[string]interface{}{
				"kind":       "APIGroup",
				"apiVersion": "v1",
				"name":       "storage.k8s.io",
				"versions": []map[string]interface{}{
					{
						"groupVersion": "storage.k8s.io/v1",
						"version":      "v1",
					},
					{
						"groupVersion": "storage.k8s.io/v1beta1",
						"version":      "v1beta1",
					},
				},
				"preferredVersion": map[string]string{
					"groupVersion": "storage.k8s.io/v1",
					"version":      "v1",
				},
			}
		}
	case path == "/apis/scheduling.k8s.io/v1":
		{
			return map[string]interface{}{
				"kind":         "APIResourceList",
				"apiVersion":   "v1",
				"groupVersion": "scheduling.k8s.io/v1",
				"resources": []map[string]interface{}{
					{
						"name":         "priorityclasses",
						"singularName": "",
						"namespaced":   false,
						"kind":         "PriorityClass",
						"verbs": []string{
							"create", "delete", "deletecollection", "get", "list", "patch", "update", "watch",
						},
						"shortNames":         []string{"pc"},
						"storageVersionHash": "1QwjyaZjj3Y=",
					},
				},
			}
		}
	case path == "/apis/scheduling.k8s.io":
		{
			return map[string]interface{}{
				"kind":       "APIGroup",
				"apiVersion": "v1",
				"name":       "scheduling.k8s.io",
				"versions": []map[string]interface{}{
					{
						"groupVersion": "scheduling.k8s.io/v1",
						"version":      "v1",
					},
				},
				"preferredVersion": map[string]string{
					"groupVersion": "scheduling.k8s.io/v1",
					"version":      "v1",
				},
			}
		}

	case path == "/apis/rbac.authorization.k8s.io/v1":
		{
			return map[string]interface{}{
				"kind":         "APIResourceList",
				"apiVersion":   "v1",
				"groupVersion": "rbac.authorization.k8s.io/v1",
				"resources": []map[string]interface{}{
					{
						"name":         "clusterrolebindings",
						"singularName": "",
						"namespaced":   false,
						"kind":         "ClusterRoleBinding",
						"verbs": []string{
							"create", "delete", "deletecollection", "get", "list", "patch", "update", "watch",
						},
						"storageVersionHash": "48tpQ8gZHFc=",
					},
					{
						"name":         "clusterroles",
						"singularName": "",
						"namespaced":   false,
						"kind":         "ClusterRole",
						"verbs": []string{
							"create", "delete", "deletecollection", "get", "list", "patch", "update", "watch",
						},
						"storageVersionHash": "bYE5ZWDrJ44=",
					},
					{
						"name":         "rolebindings",
						"singularName": "",
						"namespaced":   true,
						"kind":         "RoleBinding",
						"verbs": []string{
							"create", "delete", "deletecollection", "get", "list", "patch", "update", "watch",
						},
						"storageVersionHash": "eGsCzGH6b1g=",
					},
					{
						"name":         "roles",
						"singularName": "",
						"namespaced":   true,
						"kind":         "Role",
						"verbs": []string{
							"create", "delete", "deletecollection", "get", "list", "patch", "update", "watch",
						},
						"storageVersionHash": "7FuwZcIIItM=",
					},
				},
			}
		}
	case path == "/apis/rbac.authorization.k8s.io":
		{
			return map[string]interface{}{
				"kind":       "APIGroup",
				"apiVersion": "v1",
				"name":       "rbac.authorization.k8s.io",
				"versions": []map[string]interface{}{
					{
						"groupVersion": "rbac.authorization.k8s.io/v1",
						"version":      "v1",
					},
				},
				"preferredVersion": map[string]string{
					"groupVersion": "rbac.authorization.k8s.io/v1",
					"version":      "v1",
				},
			}
		}

	case path == "/apis/policy/v1beta1":
		{
			return map[string]interface{}{
				"kind":         "APIResourceList",
				"apiVersion":   "v1",
				"groupVersion": "policy/v1beta1",
				"resources": []map[string]interface{}{
					{
						"name":         "poddisruptionbudgets",
						"singularName": "",
						"namespaced":   true,
						"kind":         "PodDisruptionBudget",
						"verbs": []string{
							"create", "delete", "deletecollection", "get", "list", "patch", "update", "watch",
						},
						"shortNames":         []string{"pdb"},
						"storageVersionHash": "6BGBu0kpHtk=",
					},
					{
						"name":         "poddisruptionbudgets/status",
						"singularName": "",
						"namespaced":   true,
						"kind":         "PodDisruptionBudget",
						"verbs":        []string{"get", "patch", "update"},
					},
					{
						"name":         "podsecuritypolicies",
						"singularName": "",
						"namespaced":   false,
						"kind":         "PodSecurityPolicy",
						"verbs": []string{
							"create", "delete", "deletecollection", "get", "list", "patch", "update", "watch",
						},
						"shortNames":         []string{"psp"},
						"storageVersionHash": "khBLobUXkqA=",
					},
				},
			}
		}
	case path == "/apis/policy/v1":
		{
			return map[string]interface{}{
				"kind":         "APIResourceList",
				"apiVersion":   "v1",
				"groupVersion": "policy/v1",
				"resources": []map[string]interface{}{
					{
						"name":         "poddisruptionbudgets",
						"singularName": "",
						"namespaced":   true,
						"kind":         "PodDisruptionBudget",
						"verbs": []string{
							"create", "delete", "deletecollection", "get", "list", "patch", "update", "watch",
						},
						"shortNames":         []string{"pdb"},
						"storageVersionHash": "6BGBu0kpHtk=",
					},
					{
						"name":         "poddisruptionbudgets/status",
						"singularName": "",
						"namespaced":   true,
						"kind":         "PodDisruptionBudget",
						"verbs":        []string{"get", "patch", "update"},
					},
				},
			}
		}
	case path == "/apis/node.k8s.io/v1beta1":
		{
			return map[string]interface{}{
				"kind":         "APIResourceList",
				"apiVersion":   "v1",
				"groupVersion": "node.k8s.io/v1beta1",
				"resources": []map[string]interface{}{
					{
						"name":         "runtimeclasses",
						"singularName": "",
						"namespaced":   false,
						"kind":         "RuntimeClass",
						"verbs": []string{
							"create", "delete", "deletecollection", "get", "list", "patch", "update", "watch",
						},
						"storageVersionHash": "WQTu1GL3T2Q=",
					},
				},
			}
		}
	// 新增：处理 /apis/policy 路径
	case path == "/apis/policy":
		{
			return map[string]interface{}{
				"kind":       "APIGroup",
				"apiVersion": "v1",
				"name":       "policy",
				"versions": []map[string]interface{}{
					{
						"groupVersion": "policy/v1",
						"version":      "v1",
					},
					{
						"groupVersion": "policy/v1beta1",
						"version":      "v1beta1",
					},
				},
				"preferredVersion": map[string]string{
					"groupVersion": "policy/v1",
					"version":      "v1",
				},
			}
		}
	case path == "/apis/node.k8s.io/v1":
		{
			return map[string]interface{}{
				"kind":         "APIResourceList",
				"apiVersion":   "v1",
				"groupVersion": "node.k8s.io/v1",
				"resources": []map[string]interface{}{
					{
						"name":         "runtimeclasses",
						"singularName": "",
						"namespaced":   false,
						"kind":         "RuntimeClass",
						"verbs": []string{
							"create", "delete", "deletecollection", "get", "list", "patch", "update", "watch",
						},
						"storageVersionHash": "WQTu1GL3T2Q=",
					},
				},
			}
		}
	case path == "/apis/node.k8s.io":
		{
			return map[string]interface{}{
				"kind":       "APIGroup",
				"apiVersion": "v1",
				"name":       "node.k8s.io",
				"versions": []map[string]interface{}{
					{
						"groupVersion": "node.k8s.io/v1",
						"version":      "v1",
					},
					{
						"groupVersion": "node.k8s.io/v1beta1",
						"version":      "v1beta1",
					},
				},
				"preferredVersion": map[string]string{
					"groupVersion": "node.k8s.io/v1",
					"version":      "v1",
				},
			}
		}
	case path == "/apis/networking.k8s.io/v1":
		{
			return map[string]interface{}{
				"kind":         "APIResourceList",
				"apiVersion":   "v1",
				"groupVersion": "networking.k8s.io/v1",
				"resources": []map[string]interface{}{
					{
						"name":         "ingressclasses",
						"singularName": "",
						"namespaced":   false,
						"kind":         "IngressClass",
						"verbs": []string{
							"create", "delete", "deletecollection", "get", "list", "patch", "update", "watch",
						},
						"storageVersionHash": "l/iqIbDgFyQ=",
					},
					{
						"name":         "ingresses",
						"singularName": "",
						"namespaced":   true,
						"kind":         "Ingress",
						"verbs": []string{
							"create", "delete", "deletecollection", "get", "list", "patch", "update", "watch",
						},
						"shortNames":         []string{"ing"},
						"storageVersionHash": "39NQlfNR+bo=",
					},
					{
						"name":         "ingresses/status",
						"singularName": "",
						"namespaced":   true,
						"kind":         "Ingress",
						"verbs":        []string{"get", "patch", "update"},
					},
					{
						"name":         "networkpolicies",
						"singularName": "",
						"namespaced":   true,
						"kind":         "NetworkPolicy",
						"verbs": []string{
							"create", "delete", "deletecollection", "get", "list", "patch", "update", "watch",
						},
						"shortNames":         []string{"netpol"},
						"storageVersionHash": "YpfwF18m1G8=",
					},
				},
			}
		}
	case path == "/apis/networking.k8s.io":
		{
			return map[string]interface{}{
				"kind":       "APIGroup",
				"apiVersion": "v1",
				"name":       "networking.k8s.io",
				"versions": []map[string]interface{}{
					{
						"groupVersion": "networking.k8s.io/v1",
						"version":      "v1",
					},
				},
				"preferredVersion": map[string]string{
					"groupVersion": "networking.k8s.io/v1",
					"version":      "v1",
				},
			}
		}
	case path == "/apis/flowcontrol.apiserver.k8s.io/v1beta2":
		{
			return map[string]interface{}{
				"kind":         "APIResourceList",
				"apiVersion":   "v1",
				"groupVersion": "flowcontrol.apiserver.k8s.io/v1beta2", // 修正为v1beta2
				"resources": []map[string]interface{}{
					{
						"name":         "flowschemas",
						"singularName": "",
						"namespaced":   false,
						"kind":         "FlowSchema",
						"verbs": []string{
							"create", "delete", "deletecollection", "get", "list", "patch", "update", "watch",
						},
						"storageVersionHash": "9bSnTLYweJ0=",
					},
					{
						"name":         "flowschemas/status",
						"singularName": "",
						"namespaced":   false,
						"kind":         "FlowSchema",
						"verbs":        []string{"get", "patch", "update"},
					},
					{
						"name":         "prioritylevelconfigurations",
						"singularName": "",
						"namespaced":   false,
						"kind":         "PriorityLevelConfiguration",
						"verbs": []string{
							"create", "delete", "deletecollection", "get", "list", "patch", "update", "watch",
						},
						"storageVersionHash": "BFVwf8eYnsw=",
					},
					{
						"name":         "prioritylevelconfigurations/status",
						"singularName": "",
						"namespaced":   false,
						"kind":         "PriorityLevelConfiguration",
						"verbs":        []string{"get", "patch", "update"},
					},
				},
			}
		}

	case path == "/apis/flowcontrol.apiserver.k8s.io/v1beta1":
		{
			return map[string]interface{}{
				"kind":         "APIResourceList",
				"apiVersion":   "v1",
				"groupVersion": "flowcontrol.apiserver.k8s.io/v1beta1",
				"resources": []map[string]interface{}{
					{
						"name":         "flowschemas",
						"singularName": "",
						"namespaced":   false,
						"kind":         "FlowSchema",
						"verbs": []string{
							"create", "delete", "deletecollection", "get", "list", "patch", "update", "watch",
						},
						"storageVersionHash": "9bSnTLYweJ0=",
					},
					{
						"name":         "flowschemas/status",
						"singularName": "",
						"namespaced":   false,
						"kind":         "FlowSchema",
						"verbs":        []string{"get", "patch", "update"},
					},
					{
						"name":         "prioritylevelconfigurations",
						"singularName": "",
						"namespaced":   false,
						"kind":         "PriorityLevelConfiguration",
						"verbs": []string{
							"create", "delete", "deletecollection", "get", "list", "patch", "update", "watch",
						},
						"storageVersionHash": "BFVwf8eYnsw=",
					},
					{
						"name":         "prioritylevelconfigurations/status",
						"singularName": "",
						"namespaced":   false,
						"kind":         "PriorityLevelConfiguration",
						"verbs":        []string{"get", "patch", "update"},
					},
				},
			}
		}
	case path == "/apis/flowcontrol.apiserver.k8s.io":
		{
			return map[string]interface{}{
				"kind":       "APIGroup",
				"apiVersion": "v1",
				"name":       "flowcontrol.apiserver.k8s.io",
				"versions": []map[string]interface{}{
					{
						"groupVersion": "flowcontrol.apiserver.k8s.io/v1beta2",
						"version":      "v1beta2",
					},
					{
						"groupVersion": "flowcontrol.apiserver.k8s.io/v1beta1",
						"version":      "v1beta1",
					},
				},
				"preferredVersion": map[string]string{
					"groupVersion": "flowcontrol.apiserver.k8s.io/v1beta2",
					"version":      "v1beta2",
				},
			}
		}
	case path == "/apis/events.k8s.io/v1beta1":
		{
			return map[string]interface{}{
				"kind":         "APIResourceList",
				"apiVersion":   "v1",
				"groupVersion": "events.k8s.io/v1beta1",
				"resources": []map[string]interface{}{
					{
						"name":         "events",
						"singularName": "",
						"namespaced":   true,
						"kind":         "Event",
						"verbs": []string{
							"create", "delete", "deletecollection", "get", "list", "patch", "update", "watch",
						},
						"shortNames":         []string{"ev"},
						"storageVersionHash": "r2yiGXH7wu8=",
					},
				},
			}
		}
	case path == "/apis/events.k8s.io/v1":
		{
			return map[string]interface{}{
				"kind":         "APIResourceList",
				"apiVersion":   "v1",
				"groupVersion": "events.k8s.io/v1",
				"resources": []map[string]interface{}{
					{
						"name":         "events",
						"singularName": "",
						"namespaced":   true,
						"kind":         "Event",
						"verbs": []string{
							"create", "delete", "deletecollection", "get", "list", "patch", "update", "watch",
						},
						"shortNames":         []string{"ev"},
						"storageVersionHash": "r2yiGXH7wu8=",
					},
				},
			}
		}
	case path == "/apis/events.k8s.io":
		{
			return map[string]interface{}{
				"kind":       "APIGroup",
				"apiVersion": "v1",
				"name":       "events.k8s.io",
				"versions": []map[string]interface{}{
					{
						"groupVersion": "events.k8s.io/v1",
						"version":      "v1",
					},
					{
						"groupVersion": "events.k8s.io/v1beta1",
						"version":      "v1beta1",
					},
				},
				"preferredVersion": map[string]string{
					"groupVersion": "events.k8s.io/v1",
					"version":      "v1",
				},
			}
		}
	// 新增：处理 /apis/discovery.k8s.io/v1beta1 路径
	case path == "/apis/discovery.k8s.io/v1beta1":
		{
			return map[string]interface{}{
				"kind":         "APIResourceList",
				"apiVersion":   "v1",
				"groupVersion": "discovery.k8s.io/v1beta1",
				"resources": []map[string]interface{}{
					{
						"name":         "endpointslices",
						"singularName": "",
						"namespaced":   true,
						"kind":         "EndpointSlice",
						"verbs": []string{
							"create", "delete", "deletecollection", "get", "list", "patch", "update", "watch",
						},
						"storageVersionHash": "Nx3SIv6I0mE=",
					},
				},
			}
		}

	// 新增：处理 /apis/discovery.k8s.io/v1 路径
	case path == "/apis/discovery.k8s.io/v1":
		{
			return map[string]interface{}{
				"kind":         "APIResourceList",
				"apiVersion":   "v1",
				"groupVersion": "discovery.k8s.io/v1",
				"resources": []map[string]interface{}{
					{
						"name":         "endpointslices",
						"singularName": "",
						"namespaced":   true,
						"kind":         "EndpointSlice",
						"verbs": []string{
							"create", "delete", "deletecollection", "get", "list", "patch", "update", "watch",
						},
						"storageVersionHash": "Nx3SIv6I0mE=",
					},
				},
			}
		}
	// 新增：处理 /apis/discovery.k8s.io 路径
	case path == "/apis/discovery.k8s.io":
		{
			return map[string]interface{}{
				"kind":       "APIGroup",
				"apiVersion": "v1",
				"name":       "discovery.k8s.io",
				"versions": []map[string]interface{}{
					{
						"groupVersion": "discovery.k8s.io/v1",
						"version":      "v1",
					},
					{
						"groupVersion": "discovery.k8s.io/v1beta1",
						"version":      "v1beta1",
					},
				},
				"preferredVersion": map[string]string{
					"groupVersion": "discovery.k8s.io/v1",
					"version":      "v1",
				},
			}
		}
	case path == "/apis/coordination.k8s.io/v1":
		{
			return map[string]interface{}{
				"kind":         "APIResourceList",
				"apiVersion":   "v1",
				"groupVersion": "coordination.k8s.io/v1",
				"resources": []map[string]interface{}{
					{
						"name":         "leases",
						"singularName": "",
						"namespaced":   true,
						"kind":         "Lease",
						"verbs": []string{
							"create", "delete", "deletecollection", "get", "list", "patch", "update", "watch",
						},
						"storageVersionHash": "gqkMMb/YqFM=",
					},
				},
			}
		}
	// 新增：处理 /apis/coordination.k8s.io 路径
	case path == "/apis/coordination.k8s.io":
		{
			return map[string]interface{}{
				"kind":       "APIGroup",
				"apiVersion": "v1",
				"name":       "coordination.k8s.io",
				"versions": []map[string]interface{}{
					{
						"groupVersion": "coordination.k8s.io/v1",
						"version":      "v1",
					},
				},
				"preferredVersion": map[string]string{
					"groupVersion": "coordination.k8s.io/v1",
					"version":      "v1",
				},
			}
		}
	// 新增：处理 /apis/authentication.k8s.io 路径
	case path == "/apis/authentication.k8s.io":
		{
			return map[string]interface{}{
				"kind":       "APIGroup",
				"apiVersion": "v1",
				"name":       "authentication.k8s.io",
				"versions": []map[string]interface{}{
					{
						"groupVersion": "authentication.k8s.io/v1",
						"version":      "v1",
					},
				},
				"preferredVersion": map[string]string{
					"groupVersion": "authentication.k8s.io/v1",
					"version":      "v1",
				},
			}
		}
	// 新增：处理 /apis/autoscaling 路径
	case path == "/apis/autoscaling":
		{
			return map[string]interface{}{
				"kind":       "APIGroup",
				"apiVersion": "v1",
				"name":       "autoscaling",
				"versions": []map[string]interface{}{
					{
						"groupVersion": "autoscaling/v2",
						"version":      "v2",
					},
					{
						"groupVersion": "autoscaling/v1",
						"version":      "v1",
					},
					{
						"groupVersion": "autoscaling/v2beta1",
						"version":      "v2beta1",
					},
					{
						"groupVersion": "autoscaling/v2beta2",
						"version":      "v2beta2",
					},
				},
				"preferredVersion": map[string]string{
					"groupVersion": "autoscaling/v2",
					"version":      "v2",
				},
			}
		}
	case path == "/apis/autoscaling/v2beta2":
		{
			return map[string]interface{}{
				"kind":         "APIResourceList",
				"apiVersion":   "v1",
				"groupVersion": "autoscaling/v2beta2",
				"resources": []map[string]interface{}{
					{
						"name":         "horizontalpodautoscalers",
						"singularName": "",
						"namespaced":   true,
						"kind":         "HorizontalPodAutoscaler",
						"verbs": []string{
							"create", "delete", "deletecollection", "get", "list", "patch", "update", "watch",
						},
						"shortNames":         []string{"hpa"},
						"categories":         []string{"all"},
						"storageVersionHash": "oQlkt7f5j/A=",
					},
					{
						"name":         "horizontalpodautoscalers/status",
						"singularName": "",
						"namespaced":   true,
						"kind":         "HorizontalPodAutoscaler",
						"verbs":        []string{"get", "patch", "update"},
					},
				},
			}
		}
	case path == "/apis/batch":
		{
			return map[string]interface{}{
				"kind":       "APIGroup",
				"apiVersion": "v1",
				"name":       "batch",
				"versions": []map[string]interface{}{
					{
						"groupVersion": "batch/v1",
						"version":      "v1",
					},
					{
						"groupVersion": "batch/v1beta1",
						"version":      "v1beta1",
					},
				},
				"preferredVersion": map[string]string{
					"groupVersion": "batch/v1",
					"version":      "v1",
				},
			}
		}
	case path == "/apis/autoscaling/v2beta1":
		{
			return map[string]interface{}{
				"kind":         "APIResourceList",
				"apiVersion":   "v1",
				"groupVersion": "autoscaling/v2beta1",
				"resources": []map[string]interface{}{
					{
						"name":         "horizontalpodautoscalers",
						"singularName": "",
						"namespaced":   true,
						"kind":         "HorizontalPodAutoscaler",
						"verbs": []string{
							"create", "delete", "deletecollection", "get", "list", "patch", "update", "watch",
						},
						"shortNames":         []string{"hpa"},
						"categories":         []string{"all"},
						"storageVersionHash": "oQlkt7f5j/A=",
					},
					{
						"name":         "horizontalpodautoscalers/status",
						"singularName": "",
						"namespaced":   true,
						"kind":         "HorizontalPodAutoscaler",
						"verbs":        []string{"get", "patch", "update"},
					},
				},
			}
		}
	case path == "/apis/autoscaling/v2":
		{
			return map[string]interface{}{
				"kind":         "APIResourceList",
				"apiVersion":   "v1",
				"groupVersion": "autoscaling/v2",
				"resources": []map[string]interface{}{
					{
						"name":         "horizontalpodautoscalers",
						"singularName": "",
						"namespaced":   true,
						"kind":         "HorizontalPodAutoscaler",
						"verbs": []string{
							"create", "delete", "deletecollection", "get", "list", "patch", "update", "watch",
						},
						"shortNames":         []string{"hpa"},
						"categories":         []string{"all"},
						"storageVersionHash": "oQlkt7f5j/A=",
					},
					{
						"name":         "horizontalpodautoscalers/status",
						"singularName": "",
						"namespaced":   true,
						"kind":         "HorizontalPodAutoscaler",
						"verbs":        []string{"get", "patch", "update"},
					},
				},
			}
		}

	case path == "/apis/autoscaling/v1":
		{
			return map[string]interface{}{
				"kind":         "APIResourceList",
				"apiVersion":   "v1",
				"groupVersion": "autoscaling/v1",
				"resources": []map[string]interface{}{
					{
						"name":         "horizontalpodautoscalers",
						"singularName": "",
						"namespaced":   true,
						"kind":         "HorizontalPodAutoscaler",
						"verbs": []string{
							"create", "delete", "deletecollection", "get", "list", "patch", "update", "watch",
						},
						"shortNames":         []string{"hpa"},
						"categories":         []string{"all"},
						"storageVersionHash": "oQlkt7f5j/A=",
					},
					{
						"name":         "horizontalpodautoscalers/status",
						"singularName": "",
						"namespaced":   true,
						"kind":         "HorizontalPodAutoscaler",
						"verbs":        []string{"get", "patch", "update"},
					},
				},
			}
		}

	case path == "/apis/authorization.k8s.io/v1":
		{
			return map[string]interface{}{
				"kind":         "APIResourceList",
				"apiVersion":   "v1",
				"groupVersion": "authorization.k8s.io/v1",
				"resources": []map[string]interface{}{
					{
						"name":         "localsubjectaccessreviews",
						"singularName": "",
						"namespaced":   true,
						"kind":         "LocalSubjectAccessReview",
						"verbs":        []string{"create"},
					},
					{
						"name":         "selfsubjectaccessreviews",
						"singularName": "",
						"namespaced":   false,
						"kind":         "SelfSubjectAccessReview",
						"verbs":        []string{"create"},
					},
					{
						"name":         "selfsubjectrulesreviews",
						"singularName": "",
						"namespaced":   false,
						"kind":         "SelfSubjectRulesReview",
						"verbs":        []string{"create"},
					},
					{
						"name":         "subjectaccessreviews",
						"singularName": "",
						"namespaced":   false,
						"kind":         "SubjectAccessReview",
						"verbs":        []string{"create"},
					},
				},
			}
		}
	case path == "/apis/authorization.k8s.io":
		{
			return map[string]interface{}{
				"kind":       "APIGroup",
				"apiVersion": "v1",
				"name":       "authorization.k8s.io",
				"versions": []map[string]interface{}{
					{
						"groupVersion": "authorization.k8s.io/v1",
						"version":      "v1",
					},
				},
				"preferredVersion": map[string]string{
					"groupVersion": "authorization.k8s.io/v1",
					"version":      "v1",
				},
			}
		}
	// 新增：处理 /apis/authentication.k8s.io/v1 路径
	case path == "/apis/authentication.k8s.io/v1":
		{
			return map[string]interface{}{
				"kind":         "APIResourceList",
				"apiVersion":   "v1",
				"groupVersion": "authentication.k8s.io/v1",
				"resources": []map[string]interface{}{
					{
						"name":         "tokenreviews",
						"singularName": "",
						"namespaced":   false,
						"kind":         "TokenReview",
						"verbs":        []string{"create"},
					},
				},
			}
		}
	case path == "/api":
		return map[string]interface{}{
			"kind":     "APIVersions",
			"versions": []string{"v1"},
			"serverAddressByClientCIDRs": []map[string]string{
				{"clientCIDR": "0.0.0.0/0", "serverAddress": "127.0.0.1"},
			},
		}
	case path == "/api/v1":
		return map[string]interface{}{
			"kind":         "APIResourceList",
			"groupVersion": "v1",
			"resources": []map[string]interface{}{
				{
					"name":         "pods",
					"singularName": "pod",
					"namespaced":   true,
					"kind":         "Pod",
					"verbs":        []string{"get", "list", "watch", "create", "update", "patch", "delete"},
					"shortNames":   []string{"po"},
				},
				{
					"name":         "nodes",
					"singularName": "node",
					"namespaced":   false,
					"kind":         "Node",
					"verbs":        []string{"get", "list", "watch"},
					"shortNames":   []string{"no"},
				},
				{
					"name":         "services",
					"singularName": "service",
					"namespaced":   true,
					"kind":         "Service",
					"verbs":        []string{"get", "list", "watch", "create", "update", "patch", "delete"},
					"shortNames":   []string{"svc"},
				},
			},
		}
	case strings.HasPrefix(path, "/api/v1/pods"):
		return map[string]interface{}{
			"kind":       "PodList",
			"apiVersion": "v1",
			"metadata": map[string]interface{}{
				"selfLink":        "/api/v1/pods",
				"resourceVersion": "12345",
			},
			"items": []map[string]interface{}{
				{
					"metadata": map[string]interface{}{
						"name":              "nginx-deployment-76bf4969df-2bsk9",
						"namespace":         "default",
						"uid":               "a1b2c3d4-e5f6-4a5b-9c8d-7e6f5a4b3c2d",
						"creationTimestamp": time.Now().Format(time.RFC3339),
					},
					"spec": map[string]interface{}{
						"containers": []map[string]interface{}{
							{
								"name":  "nginx",
								"image": "nginx:1.14.2",
							},
						},
					},
					"status": map[string]interface{}{
						"phase": "Running",
					},
				},
			},
		}
	case strings.HasPrefix(path, "/api/v1/nodes"):
		return map[string]interface{}{
			"kind":       "NodeList",
			"apiVersion": "v1",
			"metadata": map[string]interface{}{
				"selfLink":        "/api/v1/nodes",
				"resourceVersion": "67890",
			},
			"items": []map[string]interface{}{
				{
					"metadata": map[string]interface{}{
						"name":              "node1.example.com",
						"uid":               "b2c3d4e5-f6a7-5b6c-0d1e-8f9a0b1c2d3e",
						"creationTimestamp": time.Now().Format(time.RFC3339),
					},
					"status": map[string]interface{}{
						"nodeInfo": map[string]interface{}{
							"kubeletVersion":  "v1.21.0",
							"operatingSystem": "linux",
						},
					},
				},
			},
		}
	case strings.HasPrefix(path, "/healthz"):
		return "ok"
	default:
		return map[string]interface{}{
			"kind":       "Status",
			"apiVersion": "v1",
			"metadata":   map[string]interface{}{},
			"status":     "Failure",
			"message":    "not found",
			"reason":     "NotFound",
			"code":       404,
		}
	}
}
