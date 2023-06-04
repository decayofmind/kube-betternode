# kube-better-node

![meme](https://github.com/decayofmind/kube-better-node/assets/4447814/3d64b91d-9107-4ea6-b9b6-66169cbb0597)

This small program does the only thing [kubernetes-sigs/descheduler](https://github.com/kubernetes-sigs/descheduler) can't do. It will evict Pods which could be potentially schedulled to a better Node, based on `preferredDuringSchedulingIgnoredDuringExecution` `nodeAffinity` terms.

### ⚠️ Disclaimer ⚠️

* I have no intention to replicate and follow all Descheduler policies and rules, but feel free to create a PR if something is missing.
* I do not need functionality provided by this program anymore, however I still see it usefull for exploring the Golang world.
* Remember, it' a workaround and there's no guarantee that a Pod will be schedulled to a Node you expect it to be.

## Why?

There're at least two rotten unmerged PRs in the Descheduler repository:

* [#130](https://github.com/kubernetes-sigs/descheduler/pull/130) (`FindBetterPreferredNode` and `CalcPodPriorityScore` are inspired by it. Thanks [@tsu1980](https://github.com/tsu1980))
* [#129](https://github.com/kubernetes-sigs/descheduler/pull/129)

In fact, this functionallity will be never implemented in Descheduler. 

As [@asnkh](https://github.com/asnkh) said in [#211](https://github.com/kubernetes-sigs/descheduler/issues/211#issuecomment-602026583):

> ...unless descheduler can make the same decision as kube-scheduler, it can cause this kind of ineffective pod evictions. I have no good idea to overcome this difficulty. Copying all scheduling policies in kube-scheduler to descheduler is not realistic.

## Use case

However, there're real world use-cases, where such dumb eviction can be usefull. 

For example, when your project is small and there's a need to use some special Spot instance node type (such as Tesla `p2.xlarge`).
This type of instance can be taken back from you by AWS, but if your project can afford some performance degradation for a short period of time (untill there's a new node of `p2.xlarge` given), `kube-scheduler` will place Pods on some other Node available at the moment.

But when finally new Tesla `p2.xlarge` node will apper and join the cluster, there's nothing to schedule your Pods back to it. 

Here **kube-better-node** can be usefull. 

## Usage

```sh
❯ kube-better-node -h
Usage of kube-better-node:
  -dry-run
    	Dry run
  -tolerance int
    	Ignore certain weight difference
  -v value
    	number for the log level verbosity
```

## Installation

```
helm repo add decayofmind https://decayofmind.github.io/charts/
helm install kube-better-node decayofmind/kube-better-node
```

## Links

* https://github.com/kubernetes-sigs/cluster-capacity
