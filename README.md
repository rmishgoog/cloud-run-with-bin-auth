## Using Binary Authorization with Google Cloud Run

_Google Cloud Run provides a serverless platform to run stateless containers at scale, being a serverless offering from the fastest growing public cloud on planet, Cloud Run needs no infrastructure provisioning or management from the users, you simply bring your containers (no matter what programming stack or container tool like docker or podman, as long as it can build an OCI compliant image) and supply it to Cloud Run. Very flexible, right? Not just that but it keeps your application portable across the platforms, for example you can decide to move the same container into a GKE cluster or simply run it as a container on  VM with no orchestration layer on top._

_While containers and Kubernetes have made life easier for organizations, there are also challenges that come along, one of them being the security, on a very high level, security of your Kubernetes clusters for example can be broken down into 4 C's, also called as the 4 C of cloud native security, as described here by CNCF in Kubernetes' official documentation:_

https://kubernetes.io/docs/concepts/security/overview/

_One of those Cs is the "Container", with Cloud-Cluster-Code (innermost) being the other three. So, what does it mean by securing the container? There are many sides to it, starting from base images, what packages they include, whether the container runs as a priviliged one or uses non-privilged users, other security constructs such as seccomp, apparmor as applied by your runtime or sanboxing the container with something like Kata containers, gVisor or Firecracker. That's a topic for another disucssion, here we are talking about a "serverless" container platform, which is what Cloud Run is, by the very definition of being serverless, Google has taken care of securing the Cloud infrastructure (including physical security), the cluser (yes! there's a Kubernetes cluster behind the scenes, Cloud Run is built on k-native!), so you are left with securing your container (some aspects, runtime aspects are covered by Google, again!) and the code. What you build and pack inside the container matters as much as anything else with respect to security of your cloud native applications._

_That's where software supply chain security becomes important, how do you ensure that your builds can be trusted, done by only authorized system, tested, scanned and have passed your CI/CD pipelines before making it to production? That's where Google Cloud's binary authorization plays a vital role, refer to the Binary Authorization or Bin Auth (as we are going to refer to it in this tutorial) documentation:_

https://cloud.google.com/binary-authorization/docs

_In short, Binary Authorization works mainly with three key constructs, the policy, the attestors and the attestations, attestations are like 'signed certificates' proving that a certain build artifact like your docker image has successfully passed certain validations or has been built using a "trusted" system and can be admitted into your Kubernetes cluster. Both GKE and Cloud Run support Bin Auth which can be enabled by few clicks of the button and you will see it in action. In this tutorial, we will focus on Cloud Run and demonstrate a simple use case where our Cloud Run service will only accept images built using Google Cloud Build. Any attempt to bypass Google Cloud Build and pushing images directly to the registry will result in service failing to provision a new revision (which is what we want!)._

_Let's see some Google Cloud Binary Auth in action:_

_First thing first, let's make sure you have the required tools/software set up:_
1. _Google cloud SDK_
2. _Terraform_
3. _Docker or any other OCI image spec compatible container tool like podman_
4. _A Google Cloud account, though principle of least privileges is the recommended approach, here we will use an account with owner/editor role_



