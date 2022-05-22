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
