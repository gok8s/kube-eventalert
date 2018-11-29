package controller

var UserAlertReasonType = map[string]string{
	"FailedPostStartHook": "容器创建时，执行初始化脚本出错",
	//"FailedPreStopHook":        "容器销毁时，执行脚本出错，容器会被正常销毁，需要修改PreStop shell",
	"UnfinishedPreStopHook":    "容器在销毁前执行用户自定义脚本超过用户设置的TimeOut，容器会被强行删除，请配置正确的Shell脚本",
	"HostPortConflict":         "Host节点端口冲突，请配置正确的端口",
	"BackOff":                  "BackOff",
	"Failed":                   "Failed",
	"CrashLoopBackOff":         "容器启动失败",
	"FailedCreatePodContainer": "创建容器失败",
	//"Unhealthy":                "容器Health接口异常，状态为Unhealthy",
	"Killing": "容器被删除",
}

var AdminAlertReasonType = map[string]string{
	"NodeHasNoDiskPressure":    "Node节点没有足够的磁盘可供分配",
	"NodeHasSufficientMemory":  "Node节点没有足够的内存可供分配",
	"NodeHasSufficientDisk":    "Node节点没有足够的硬盘可供分配",
	"NodeNotReady":             "Node节点不可用",
	"SandboxChanged":           "Network Error Or init image was changed",
	"InvalidDiskCapacity":      "InvalidDiskCapacity ",
	"FreeDiskSpaceFailed":      "FreeDiskSpaceFailed",
	"InsufficientFreeCPU":      "没有足够的CPU",
	"InsufficientFreeMemory":   "没有足够的内存",
	"HostNetworkNotSupported":  "HostNetworkNotSupported",
	"FailedCreatePodContainer": "FailedCreatePodContainer",
	"Failed":                   "Failed", //FailedToPullImage       =  需要过滤FailedPullImages，Failed to pull image
	"NodeNotSchedulable":       "Node不可被调度",
	"KubeletSetupFailed":       "KubeletSetupFailed",
	"FailedAttachVolume":       "FailedAttachVolume",
	"FailedDetachVolume":       "FailedDetachVolume",
	"VolumeResizeFailed":       "VolumeResizeFailed",
	"FileSystemResizeFailed":   "FileSystemResizeFailed",
	"FailedUnMount":            "FailedUnMount",
	"FailedUnmapDevice":        "FailedUnmapDevice",
	"HostPortConflict":         "Host节点端口冲突，请配置正确的端口",
	"NodeSelectorMismatching":  "NodeSelectorMismatching",
	"NilShaper":                "NilShaper",
	"Rebooted":                 "Rebooted",
	"ContainerGCFailed":        "ContainerGCFailed",
	"ImageGCFailed":            "ImageGCFailed",
	"FailedCreatePodSandBox":   "FailedCreatePodSandBox",
	"FailedPodSandBoxStatus":   "FailedPodSandBoxStatus",
	"ErrImageNeverPull":        "ErrImageNeverPull",
	"NetworkNotReady":          "NetworkNotReady",
	"FailedKillPod":            "FailedKillPod",
	"RemovingNode":             "节点被执行下线",
	/*
		以下几种暂时不接受
			"FailedSync":               "FailedSync",
			"FailedMount":              "FailedMount",
			"BackOff":                  "BackOffPullImage",#User级别已经有了
	*/
}

var NormalReasonType = map[string]string{
	"NodeSchedulable":       "NodeSchedulable",
	"Pulling":               "Pulling",
	"Scheduled":             "Scheduled",
	"Pulled":                "Pulled",
	"Started":               "Started",
	"Created":               "Created",
	"SuccessfulMountVolume": "SuccessfulMountVolume",
	"SuccessfulCreate":      "SuccessfulCreate",
	"SuccessfulDelete":      "SuccessfulDelete",
}

var RecoverReasonType = map[string]string{}

var AdminAlertNS = map[string]string{
	"kube-system":            "kube-system-ns",
	"kube-public":            "kube-public-ns",
	"default":                "default-ns",
	"ingress-nginx":          "ingress-nginx",
	"ingress-nginx-decision": "ingress-nginx-decision",
	"ingress-nginx-newapp":   "ingress-nginx-newapp",
}
