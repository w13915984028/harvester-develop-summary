
## manifest upgrade

### embedded rancher upgrade

### harvester managedchart upgrade

### addons upgrade

### others



## node upgrade


### before node upgrade


upgrade controller put labels on upgrade.

```
StateUpgradingNodes          = "UpgradingNodes"

		toUpdate := upgrade.DeepCopy()
		singleNodeName := upgrade.Status.SingleNode
		if singleNodeName != "" {
			logrus.Info("Start single node upgrade job")
			if _, err = h.jobClient.Create(applyNodeJob(upgrade, info, singleNodeName, upgradeJobTypeSingleNodeUpgrade)); err != nil && !apierrors.IsAlreadyExists(err) {
				setUpgradeCompletedCondition(toUpdate, StateFailed, corev1.ConditionFalse, err.Error(), "")
				return h.upgradeClient.Update(toUpdate)
			}
		} else {

			...
			// go with RKE2 pre-drain/post-drain hooks
			logrus.Infof("Start upgrading Kubernetes runtime to %s", info.Release.Kubernetes)
			if err := h.upgradeKubernetes(info.Release.Kubernetes); err != nil {
				setUpgradeCompletedCondition(toUpdate, StateFailed, corev1.ConditionFalse, err.Error(), "")
				return h.upgradeClient.Update(toUpdate)
			}
		}

		toUpdate.Labels[upgradeStateLabel] = StateUpgradingNodes
		harvesterv1.NodesUpgraded.CreateUnknownIfNotExists(toUpdate)
		return h.upgradeClient.Update(toUpdate)
```

And via `upgradeKubernetes`, it turns help to Rancher/RKE2 to continue the upgrade.


```
func (h *upgradeHandler) upgradeKubernetes(kubernetesVersion string) error {
	cluster, err := h.clusterCache.Get("fleet-local", "local")
	if err != nil {
		return err
	}

	toUpdate := cluster.DeepCopy()
	toUpdate.Spec.KubernetesVersion = kubernetesVersion

	if toUpdate.Spec.RKEConfig == nil {
		toUpdate.Spec.RKEConfig = &provisioningv1.RKEConfig{}
	}

	toUpdate.Spec.RKEConfig.ProvisionGeneration++
	toUpdate.Spec.RKEConfig.UpgradeStrategy.ControlPlaneConcurrency = "1"
	toUpdate.Spec.RKEConfig.UpgradeStrategy.WorkerConcurrency = "1"
	toUpdate.Spec.RKEConfig.UpgradeStrategy.ControlPlaneDrainOptions.DeleteEmptyDirData = rke2DrainNodes
	toUpdate.Spec.RKEConfig.UpgradeStrategy.ControlPlaneDrainOptions.Enabled = rke2DrainNodes
	toUpdate.Spec.RKEConfig.UpgradeStrategy.ControlPlaneDrainOptions.Force = rke2DrainNodes
	toUpdate.Spec.RKEConfig.UpgradeStrategy.ControlPlaneDrainOptions.IgnoreDaemonSets = &rke2DrainNodes
	toUpdate.Spec.RKEConfig.UpgradeStrategy.WorkerDrainOptions.DeleteEmptyDirData = rke2DrainNodes
	toUpdate.Spec.RKEConfig.UpgradeStrategy.WorkerDrainOptions.Enabled = rke2DrainNodes
	toUpdate.Spec.RKEConfig.UpgradeStrategy.WorkerDrainOptions.Force = rke2DrainNodes
	toUpdate.Spec.RKEConfig.UpgradeStrategy.WorkerDrainOptions.IgnoreDaemonSets = &rke2DrainNodes

	updateDrainHooks(&toUpdate.Spec.RKEConfig.UpgradeStrategy.ControlPlaneDrainOptions.PreDrainHooks, preDrainAnnotation)
	updateDrainHooks(&toUpdate.Spec.RKEConfig.UpgradeStrategy.ControlPlaneDrainOptions.PostDrainHooks, postDrainAnnotation)
	updateDrainHooks(&toUpdate.Spec.RKEConfig.UpgradeStrategy.WorkerDrainOptions.PreDrainHooks, preDrainAnnotation)
	updateDrainHooks(&toUpdate.Spec.RKEConfig.UpgradeStrategy.WorkerDrainOptions.PostDrainHooks, postDrainAnnotation)

	_, err = h.clusterClient.Update(toUpdate)
	return err
}

		
```

e.g.

```
    labels:
      harvesterhci.io/latestUpgrade: "true"
      harvesterhci.io/upgradeState: UpgradingNodes
    name: hvst-upgrade-bcq9d
    namespace: harvester-system
```

logs:
```
time="2025-06-03T14:20:25Z" level=info msg="handle upgrade harvester-system/hvst-upgrade-bcq9d with labels map[harvesterhci.io/latestUpgrade:true harvesterhci.io/upgradeState:UpgradingSystemServices]"
time="2025-06-03T14:20:25Z" level=info msg="Start upgrading Kubernetes runtime to v1.31.4+rke2r1"
time="2025-06-03T14:20:26Z" level=info msg="handle upgrade harvester-system/hvst-upgrade-bcq9d with labels map[harvesterhci.io/latestUpgrade:true harvesterhci.io/upgradeState:UpgradingNodes]"
time="2025-06-03T14:20:26Z" level=info msg="handle upgrade harvester-system/hvst-upgrade-bcq9d with labels map[harvesterhci.io/latestUpgrade:true harvesterhci.io/upgradeState:UpgradingNodes]"
time="2025-06-03T14:20:56Z" level=info msg="node nogal01 is in upgrade pre-draining state, try to detach unused volumes"
time="2025-06-03T14:20:56Z" level=info msg="requeue the node nogal01 while waiting for these volumes can be detached: pvc-34a5c936-7d28-4525-bcb1-c28e042a45f5, pvc-6266503e-98e5-400f-856c-e71ac7c76492, pvc-84d1ec83-4511-4836-9cd9-4f65aee053c1, pvc-ce8fe5f1-1fa6-429e-89fc-aa32fa72b1c5, pvc-514e23a2-663d-4be1-81a9-12f3e1d36427, pvc-a68f737a-0778-4c05-ada0-6692ede8f654, pvc-df9ebf54-79a2-4545-b5b6-4407580c7285, pvc-f7331231-35c1-4d4a-bb6c-3e8ac12b5d8e, pvc-00b4411a-33fa-4dee-b2c7-22454753bf10, pvc-247dcf0c-6512-4c03-b1c0-df7541d3207e, pvc-26c17e94-c43b-4f57-98b5-062c2b735cad, pvc-503f72c1-8ea1-4a0c-82f9-04e9489b57f9, pvc-2faf86dc-2f90-46ff-9b49-7b2475782a33, pvc-a8c757d0-ccb2-42fe-8cc8-e05d967a03b9, pvc-bb615477-2595-4362-a527-2f4583db633d, pvc-e0cb345c-f76d-498a-9267-527597e0aebb, pvc-e45b2f82-3a26-4930-aa3f-0dc6b37bbdaa, pvc-1606fdc2-f55b-440c-97e7-ec4bc1936be4, pvc-9a701b0d-d60e-4eee-9456-f26665575058, pvc-d8bbc607-5e6e-4e0f-9ef9-f1c3f03a9ab5, pvc-fe5c3c0a-6f8a-4c59-95bd-c3644d2023e4, pvc-40a7a1b9-7717-4802-b034-96fd17e8120a, pvc-8539707a-3033-4617-aca8-f95543edd760, pvc-dab39ef0-f3d7-432a-8dc1-2fabe617f86a, pvc-4330ec0d-981d-434c-8ce5-19fa6f7bc44d, pvc-a306f6a9-4841-421b-b741-daa3e58c4c06, pvc-f3a40688-38d7-4240-8d86-d9e77753396f, pvc-2fc1f76a-362b-4bc5-8def-e28613151073, pvc-34822d7f-44a7-40f8-9d79-3c5c8024ddb1, pvc-6675f68f-7af3-4867-b268-caec661dd3f2, pvc-681eccd4-4a0a-48a4-9b3f-ef1db36c0808, pvc-6abfee90-7ee1-4b94-83b7-87bdb2118860, pvc-a1578290-2f93-406e-aeed-5cdb4fc34a58"
```


### job controller

The job controller checks job and update the upgrade.harvesterhci status and related upgrade secret.


job: harvester-system or cattle-system

with labels:

	upgradePlanLabel                = "upgrade.cattle.io/plan"
	upgradeNodeLabel                = "upgrade.cattle.io/node"

example:

```
    labels:
      harvesterhci.io/node: nogal01
      harvesterhci.io/upgrade: hvst-upgrade-bcq9d
      harvesterhci.io/upgradeComponent: node
      harvesterhci.io/upgradeJobType: pre-drain
    name: hvst-upgrade-bcq9d-pre-drain-nogal01
    namespace: harvester-system
```

### state-machine	

#### pre-drain

1. via job-controller

1. move node status to `nodeStatePreDrained` if successful

```

upgradeJobTypePreDrain          = "pre-drain"


			if jobType == upgradeJobTypePreDrain && nodeState == nodeStatePreDraining {
				logrus.Debugf("Pre-drain job %s is done.", job.Name)
				setNodeUpgradeStatus(toUpdate, nodeName, nodeStatePreDrained, "", "")                  
				preDrained = true
			}


rancherPlanSecretNamespace    = "fleet-local"
rancherPlanSecretMachineLabel = "rke.cattle.io/machine-name"
rancherPlanSecretType         = "rke.cattle.io/machine-plan"


	secrets, err := h.secretClient.List(rancherPlanSecretNamespace, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", rancherPlanSecretMachineLabel, machineName),
		FieldSelector: fmt.Sprintf("type=%s", rancherPlanSecretType),
	})


preDrainAnnotation  = "harvesterhci.io/pre-hook"
rke2PreDrainAnnotation  = "rke.cattle.io/pre-drain"

	if preDrained {
		toUpdate := secret.DeepCopy()
		toUpdate.Annotations[preDrainAnnotation] = secret.Annotations[rke2PreDrainAnnotation]
		if _, err := h.secretClient.Update(toUpdate); err != nil {                                      // from secret to trigger post-drain
			return nil, err
		}
	}


```

##### pre-drain tasks

```
command_pre_drain() {
  recover_rancher_system_agent

  wait_longhorn_engines

  # Shut down non-live migratable VMs
  upgrade-helper vm-live-migrate-detector "$HARVESTER_UPGRADE_NODE_NAME" --shutdown

  # Live migrate VMs
  kubectl taint node $HARVESTER_UPGRADE_NODE_NAME --overwrite kubevirt.io/drain=draining:NoSchedule

  # Wait for VM migrated
  wait_vms_out_or_shutdown

  # KubeVirt's pdb might cause drain fail
  wait_evacuation_pdb_gone

  remove_rke2_canal_config
  disable_rke2_charts
}
```

##### checkpoints

kubectl get cluster -n fleet-local local -oyaml
...

```
  - lastTransitionTime: "2025-06-03T14:20:26Z"
    message: 'draining bootstrap node(s) custom-6abd433f0f30: draining node'   /////////////  
    reason: Waiting
    status: Unknown
    type: ControlPlaneReady
```


#### pos-drain


```


			} else if jobType == upgradeJobTypePostDrain && nodeState == nodeStatePostDraining {
				logrus.Debugf("Post-drain job %s is done.", job.Name)
				if repoInfo.Release.OS == node.Status.NodeInfo.OSImage {
					setNodeUpgradeStatus(toUpdate, nodeName, StateSucceeded, "", "")
					postDrained = true
				} else {
					setNodeUpgradeStatus(toUpdate, nodeName, nodeStateWaitingReboot, "", "")
					if err := h.setNodeWaitRebootLabel(node, repoInfo); err != nil {
						return nil, err
					}
					// postDrain ack will be handled in node controller
				}


postDrainAnnotation = "harvesterhci.io/post-hook"
rke2PostDrainAnnotation = "rke.cattle.io/post-drain"


	if postDrained {
		toUpdate := secret.DeepCopy()
		toUpdate.Annotations[postDrainAnnotation] = secret.Annotations[rke2PostDrainAnnotation]
		if _, err := h.secretClient.Update(toUpdate); err != nil {
			return nil, err
		}
	}

```

### node-controller


```

harvesterNodePendingOSImage = "harvesterhci.io/pendingOSImage"

func (h *nodeHandler) OnChanged(_ string, node *corev1.Node) (*corev1.Node, error) {
	if node == nil || node.DeletionTimestamp != nil || node.Annotations == nil {
		return node, nil
	}

	expectedVersion, ok := node.Annotations[harvesterNodePendingOSImage]
	if !ok {
		return node, nil
	}
```

### secret-controller

Follow above, use annotation to check if continue or stop

```

	if secret.Annotations[rke2PreDrainAnnotation] == secret.Annotations[preDrainAnnotation] && secret.Annotations[rke2PostDrainAnnotation] == secret.Annotations[postDrainAnnotation] {
		return secret, nil
	}


	switch upgrade.Status.NodeStatuses[nodeName].State {
	case nodeStateImagesPreloaded:
		if secret.Annotations[rke2PreDrainAnnotation] != secret.Annotations[preDrainAnnotation] {
			if err := checkEligibleToDrain(upgrade, nodeName); err != nil {
				return nil, err
			}
			logrus.Debugf("Create pre-drain job on %s", nodeName)
			if err := h.createHookJob(upgrade, nodeName, upgradeJobTypePreDrain, nodeStatePreDraining); err != nil {  // pre-drain job
				return nil, err
			}
		}
	case nodeStatePreDrained:
		if secret.Annotations[rke2PostDrainAnnotation] != secret.Annotations[postDrainAnnotation] {
			if err := checkEligibleToDrain(upgrade, nodeName); err != nil {
				return nil, err
			}
			logrus.Debugf("Create post-drain job on %s", nodeName)
			if err := h.createHookJob(upgrade, nodeName, upgradeJobTypePostDrain, nodeStatePostDraining); err != nil {
				return nil, err
			}
		}
	}

```

## Misc

### Enable Rancher debug


```
enable Rancher debug:

$ kubectl edit deployment -n cattle-system rancher

add below line

      containers:
      - args:
        - --no-cacerts
        - --http-listen-port=80
        - --https-listen-port=443
        - --add-local=true
        - --debug=true  // newly added

```

## upgrade repo

The repo is on a running VM, which has a http server for downloading the new ISO related files

### test repo

The http endpoint for upgrade related pods to use is:

```
http://upgrade-repo-hvst-upgrade-g62sf.harvester-system/harvester-iso/harvester-release.yaml
```

To test it on Harvester host, check the upgrade repo's carrier pod's IP, then access via IP directly.

```
curl -sSfL http://http://10.52.3.48/harvester-iso/harvester-release.yaml -o /tmp/harvester-release.yaml

```

### manually stop/start

When upgrade repo vm is in abnormal status.

```
run following commands on Harvester host with sudo -i, where it can access the embedded kubeconfig file

$ virtctl stop upgrade-repo-hvst-upgrade-xyzab -n harvester-system
$ virtctl start upgrade-repo-hvst-upgrade-xyzab -n harvester-system
```

