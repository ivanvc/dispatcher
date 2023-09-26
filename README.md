# dispatcher
Dispatcher is a simple Kubernetes job enqueuer. It uses job templates stored
in the cluster, and exposes an HTTP service to enqueue them.

It uses a job template
definition stored in a JobTemplate CRD, and it then executing it is done through
creating a JobExecution. Invoking a JobExecution can be done through its HTTP
API. Therefore, executing jobs doesn't require clients to know the actual job
template, but just the job's name.

## Description
The first step to use Dispatcher, is to have a JobTemplate (CRD) defined in
the cluster. The JobTemplate CRD defines the spec from
`batchv1.JobTemplateSpec`, so the syntax is the same as defining a CronJob or a
Job.

Given a JobTemplate like with the following definition:

```yaml
apiVersion: dispatcher.ivan.vc/v1alpha1
kind: JobTemplate
metadata:
  name: jobtemplate-sample
spec:
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: my-container
            image: alpine:latest
            command: ["echo", "$PAYLOAD"]
            env:
            - name: PAYLOAD
              value: "{{ .Payload }}"
          restartPolicy: Never
      backoffLimit: 0
```

It can be executed by manually creating a JobExecution (CRD), or by calling the
HTTP API endpoint. Although the former is possible, is the least desired way to
execute a Job.

Manually executing the previously defined JobTemplate, can be done by creating
the following JobExecution:

```yaml
apiVersion: dispatcher.ivan.vc/v1alpha1
kind: JobExecution
metadata:
  name: jobexecution-sample
spec:
  jobTemplateName: jobtemplate-sample
  payload: my test payload
```

The dispatcher controller will create a Job using the `jobexecution-sample`
template, and will feed `Payload` with `"my test payload"`. Therefore, the
output of the job would be just a simple echo of this payload.

The best way to execute the JobTemplate would be using the HTTP API endpoint, it
could be done by calling:

```bash
curl http://dispatcher-manager/execute/[namespace]/jobexecution-sample -X PUT -d
'my test payload'
```

Replace `[namespace]` with the actual namespace where you created the
JobTemplate. You can also configure a default namespace (by default it is
`"default"`), and omit the namespace from the URL path.

By running this, it will create a JobExecution, that will create a Job with the
payload that it received from the HTTP request body.

## Getting Started
You’ll need a Kubernetes cluster to run against. You can use
[KIND](https://sigs.k8s.io/kind) to get a local cluster for testing, or run
against a remote cluster. **Note:** Your controller will automatically use the
current context in your kubeconfig file (i.e. whatever cluster `kubectl
cluster-info` shows).

### Running on the cluster
1. Install Instances of Custom Resources:

```sh
kubectl apply -f config/samples/
```

2. Build and push your image to the location specified by `IMG`:
	
```sh
make docker-build docker-push IMG=<some-registry>/dispatcher:tag
```
	
3. Deploy the controller to the cluster with the image specified by `IMG`:

```sh
make deploy IMG=<some-registry>/dispatcher:tag
```

### Uninstall CRDs
To delete the CRDs from the cluster:

```sh
make uninstall
```

### Undeploy controller
UnDeploy the controller to the cluster:

```sh
make undeploy
```

## Contributing
Pull Requests are welcome for new features. If you have any issues or ideas to
discuss, feel free to open an issue.

### How it works
This project aims to follow the Kubernetes [Operator
pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/)

It uses
[Controllers](https://kubernetes.io/docs/concepts/architecture/controller/)
which provide a reconcile function responsible for synchronizing resources
untile the desired state is reached on the cluster.

### Test It Out
1. Install the CRDs into the cluster:

```sh
make install
```

2. Run your controller (this will run in the foreground, so switch to a new terminal if you want to leave it running):

```sh
make run
```

**NOTE:** You can also run this in one step by running: `make install run`

### Modifying the API definitions
If you are editing the API definitions, generate the manifests such as CRs or CRDs using:

```sh
make manifests
```

**NOTE:** Run `make --help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## License

Copyright 2022 Ivan Valdes.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

