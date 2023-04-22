# Suite of Tools for Kubernetes and Cloud Infrastructure

This project provides a suite of tools to perform tasks on Kubernetes or cloud infrastructure in AWS or GCP. These tools are designed to run more complex tasks than one-liner bash scripts, and are implemented in Golang which is easier to write, maintain, and test than a bash script.
Writing these tools in Golang allows for easier packaging and portability, e.g. it can be packaged in a container image and deployed anywhere.

## Tools included in this project

`k8s` - run various tasks on kubernetes cluster(s)


## Getting started

To get started with this project, you'll need to have Golang installed on your machine. You'll also need access to a Kubernetes cluster or cloud infrastructure in AWS or GCP.

Once you have these prerequisites, you can clone this repository and build the tools using the go build command. You can then run the tools using the generated binary.

Container image currently is not provided. To build this project run:
```bash
$ make install
```

## Detailed description

### `k8s`

It is always better to get required insight using metrics, however in some cases the query to get specific results is too complex or the result when plotted on a graph is not providing insight you are after.
`spread-by-zone` tool lists spread by zone for each Deployment and Statefulset. When run with `output=text` option the result provided can look like:
```
ns-foo      my-deployment         ***                **
ns-foo      another-deployment    **       **        **
```
where `*` represents a pod in a specific zone, my-deployment has 3 pods in zone a, none in zone b and 2 in zone c.

To get a comma-delimited output for further parsing in a spreadsheet app run with `csv` output format

`./toolbox k8s spread-by-zone output=csv`


## License

This project is licensed under the Apache 2.0 License. See the LICENSE file for more information.
