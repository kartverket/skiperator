# Contributing to Skiperator

## Getting started with development

### Installing dependencies

You're going to want to make sure to have the following things installed:

- [Operator SDK](https://sdk.operatorframework.io/docs/building-operators/golang/installation)
- golang
- kubectl
- [kubectx and kubens](https://github.com/ahmetb/kubectx)
- docker version 17.03+.
- [kind](https://kind.sigs.k8s.io)
- [istioctl](https://istio.io/latest/docs/setup/install/istioctl/)

### Running the operator

Check out the project from git. If you want to get to know the project
hierarchy, have a look at the [kubebuilder documentation](https://book.kubebuilder.io/cronjob-tutorial/basic-project.html).

```
$ git clone git@github.com:kartverket/skiperator-poc.git
```

Start a cluster on docker using `kind`.
```
$ kind create cluster
```
Make sure Kind is the active context
```
$ kubectx kind-kind
```
Install Istio to make sure all CRDs are available
```
$ istioctl install
```
Run `make` to compile the project. If you wish to see what commands are
available, run `make help` for a list of all commands. We're going to install
the CRD into the cluster and then run the operator on your machine to make
development as quick as possible.
```
$ make install run
```

This should bring up the operator. If you see any errors, adress those.

Leaving the process above running, open a new terminal and run the following
commands to apply an Application CR into the cluster. The operator will see the
CR and create all the associated files for that app.
```
$ kubectl create ns skiperator-test
$ kubens skiperator-test
$ kubectl apply -f config/samples/skiperator_v1alpha1_skip.yaml
```

For now it also requires that the `github-auth` secret is placed in the
namespace manually, so create that if it complains about a missing secret.

Now you should have a running app in your namespace. Run the following command
to see all the created resources.
```
$ kubectl get Application,all,networkpolicies,PeerAuthentication,Gateway,VirtualService,Sidecar
```

Now keep developing the app, change source code and restart the operator for
every time you want to recompile. 

## Writing code

Have a look at the following files:
- The main logic is found in the reconcile function in https://github.com/kartverket/skiperator-poc/blob/main/controllers/application_controller.go
- The `Application` Custom Resource API is found in https://github.com/kartverket/skiperator-poc/blob/main/api/v1alpha1/application_types.go
- Examples `Applications` that can be applied to the cluster are found in https://github.com/kartverket/skiperator-poc/tree/main/config/samples

Also look at the following documentation pages:
- https://sdk.operatorframework.io/docs/building-operators/golang/tutorial
- https://sdk.operatorframework.io/docs/building-operators/golang/references/client/
