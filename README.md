## Use Binary Authorization with Google Cloud Run to allow Cloud Build built images only

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

_Form your workstation, where Google Cloud SDK is installed, autheticate using gcloud:_
```
gcloud auth login
```
_At this moment, your cloud SDK is authenticated with the credentials you provided, next set up the application defaults or the ADC to be used by Google cloud libraries when making API calls, this is critical for the Terraform provider as well._
```
gcloud auth application-default login
```
_Follow the instructions and updated the application defaults, this will set up the right project context and billing context._

_Next, verify the gcloud settings:_
```
gcloud auth list
```
_Make sure that the account listed is the same as you intend to use (owner/editor)._
```
gcloud config list
```
_Check the project and account, if the project is not what you intend to use, set the project via gcloud_:
```
gcloud config set project <project-id>
```
_At this moment, if everything looks good, let's proceed with cloning the git repo locally._
```
git clone https://github.com/rmishgoog/cloud-run-with-bin-auth.git
```
_Next, we will enable a few services and APIs that we need for this tutorial:_
```
gcloud services enable binaryauthorization.googleapis.com  containerregistry.googleapis.com run.googleapis.com compute.googleapis.com --async
```
_It may take a couple of minutes for all these APIs to get enabled, so wait for about 2 minutes here before proceeding._

_Once the services we need have been enabled, we can not look at our current bin auth polciy and mofify it to add the Cloud Build attestor, this will make Cloud Run service for which bin auth is enabled, verify the image and therefore the attestation created by Cloud Build. Cloud Build automatically adds an attestation to the images that it builts and it is stored alongside the image metadata._

_Get the current bin auth policy on the project:_
```
gcloud container binauthz policy export > current_policy_exported.yaml
```
_This file should look simarl to this:_
```
defaultAdmissionRule:
  enforcementMode: ENFORCED_BLOCK_AND_AUDIT_LOG
  evaluationMode: REQUIRE_ATTESTATION
globalPolicyEvaluationMode: ENABLE
name: projects/rmishra-kubernetes-playground/policy
```
_Now, modify the file and add the Cloud Buil attestors to the required attestors, a sample file in the repo is provided, you can use that and make modifications._
```
defaultAdmissionRule:
  enforcementMode: ENFORCED_BLOCK_AND_AUDIT_LOG
  evaluationMode: REQUIRE_ATTESTATION
  requireAttestationsBy:
    - projects/<PROJECT ID>/attestors/built-by-cloud-build
globalPolicyEvaluationMode: ENABLE
name: projects/<PROJECT ID>/policy
```
_Remember to replace the PROJECT ID whit the project you are working in._

_Next, go ahead an import this policy back. Pay attention to the YAML file name, you can copy the exported file, rename and modify there._
```
gcloud container binauthz policy import current_policy.yaml
```
_Check if the Cloud Build attestor is getting listed or not._
```
gcloud container binauthz attestors list
```
_This should produce an output like below (you will see your project name though):_
```
──────────────────────┬───────────────────────────────────────────────────────────────────┬─────────────────┐
│         NAME         │                                NOTE                               │ NUM_PUBLIC_KEYS │
├──────────────────────┼───────────────────────────────────────────────────────────────────┼─────────────────┤
│ built-by-cloud-build │ projects/rmishra-kubernetes-playground/notes/built-by-cloud-build │ 30              │
└──────────────────────┴───────────────────────────────────────────────────────────────────┴─────────────────┘
```
_Now, we will build the application (into a docker image) using Cloud Build. For this tutorial, we will directly build from source, Cloud Build uses CNCNF buildpacks for building container images for a wide variety of runtime environments. Ours is a Go application which is fully supported by Google Cloud Build for direct from source type build._

_Navigate to the application source code directory, here you will also find the Dockerfile to be used by Cloud Build for building your image._
```
cd cloud-run-with-bin-auth
```
```
cd service-source-code/
```
_Execute the Google Cloud Build from the CLI using gcloud SDK, remember to replace the PROJECT ID:_
```
gcloud builds submit --tag=gcr.io/<PROJECT ID>/product-listing-api:binauth
```
_Wait for the build to finish, upon the completion, you will see the image present in the Google Cloud Container Registry, to distinguish the attested image from a non-attested image, we will use easy to understand image tags, for example, the image we just built is an attested image and thus we use the tag binauth._

_Next, let's deploy the Cloud Run service using this image. We are going to use Terraform for provisioning and making changes when we experiment between an attested and an un-attested image, instructions to use Terraform are provided here, you should just make sure that Terraform is installed on the workstation you are executing the tutorial from._

_Switch to the terraform directory:_
```
cd ../cloud-run-provisioning/
```
_Create a file terraform.tfvars and provide a value for the following variables:_
```
project   = "<PROJECT ID>"
developer = "user:<ACCOUNT EMAIL>"
```
_Provide values for the PROJECT ID and ACCOUNT EMAIL, PROJECT ID should be the project configured in gcloud SDK for the context, use the same account you authenticated gcloud with for the ACCOUNT EMAIL, for this tutorial the account should have Editor/Owner role on the project, as mentioned earlier in a more production type scenario, principle of least privileges must apply and thus you may end up using a service account instead with sufficient permissions to deploy Cloud Run service._

_If you have used a different tag for the container image then make sure you update the main.tf._
```
image = "gcr.io/${var.project}/product-listing-api:binauth"
```
_The annotation which indicates to Cloud Run that a binary authorization policy must be enforced before allowing the container image:_
```
"run.googleapis.com/binary-authorization" : "default"
```
_Here we are using the default policy that is available in the project, the policy was updated above and imported after adding Cloud Build's attestation as a required attestation._

_Now let's deploy the Cloud Run service:_
```
terraform init
```
_Check the plan generated by Terraform._
```
terraform plan
```
_Finally, deploy the service alongside other required Google Cloud resources._
```
terraform apply -auto-approve
```
_Check for any errors, and if none, your Cloud Run service is provisioned successfully, you can check that out in the Cloud Console. In the terraform code that we used, we had granted the user account used in this tutorial the run.invoker role, so you can also validate by simply issuing the curl and passing in the identity token for this account to see if the service is responding (note service URLs will vary)._
```
curl -H "Authorization: Bearer $(gcloud auth print-identity-token)" https://cloudrun-binauth-demo-service-7ljpzipopq-uc.a.run.app/cars
```

_Alright, so deployment was successful. Now, let's try to by pass Cloud Build and deploy a locally build container image and see how Binary Authorization will prevent this image from being deployed and render the new Cloud Run revision in error, not serving any traffic and completely unusable._
```
cd ../service-source-code/
```
```
docker build -t gcr.io/<project id>/product-listing-api:nobinauth .
```
```
docker push gcr.io/<project id>/product-listing-api:nobinauth
```
_Note, remember to replace the "project id" with your project, additionally, on this workstation we have autneticated docker with the gcloud credentials helper, for more information on how it works, refer to the below documentation._
  
https://cloud.google.com/container-registry/docs/advanced-authentication#gcloud-helper

_Now, let's update the terraform code to deploy the new image which was not built by Cloud Build but rather a developer with access to push images into GCR, built it locally and eventually pushed it, this can be a risky proposition for the enterprises as you maynot want to trust any random images being pushed and deployed to Production, they may contain known vulnerabilities, bugs or security flaws which can compromise your production environment, moreover they literally bypass the whole CI sub-system, for example through configured steps in Cloud Build, before an image is built and pushed, you may want to run some code scans, unit tests or integration tests, this way you can be more assured that an image pushed to GCR after CI sub-system has executed can be trusted better._

```
cd ../cloud-run-provisioning/
```
_In the main.tf, make this change:_
```
containers {
        image = "gcr.io/${var.project}/product-listing-api:nobinauth"
      }
```
```
terraform plan
```
```
Plan: 0 to add, 1 to change, 0 to destroy.
```
_Alright, so just the service to change here, let's deploy it:_
```
terraform apply -auto-approve
```
_And we are informed as expected that image was found with no signed attestation which can be verified by Cloud Build attestors:_
```
Error: resource is in failed state "Ready:False", message: Container image 'gcr.io/rmishra-kubernetes-playground/product-listing-api@sha256:81a7a088debb1b38198a3e00900d26f7e1fdeea6f1869e94a120185b8443cf71' is not authorized by policy. Image gcr.io/rmishra-kubernetes-playground/product-listing-api@sha256:81a7a088debb1b38198a3e00900d26f7e1fdeea6f1869e94a120185b8443cf71 denied by attestor projects/rmishra-kubernetes-playground/attestors/built-by-cloud-build: No attestations found that were valid and signed by a key trusted by the attestor
```
_Now, revert back to the image with attestaion, in the main.tf file:_
```
image = "gcr.io/${var.project}/product-listing-api:binauth"
```
```
terraform apply -auto-approve
```
_As we can see, the service's new revision with an attested image gets deployed successfully and works as expected._

_Finally, clean up with Terraform._
```
terraform destroy -auto-approve
```
_P.S.: Opinions are personal and do not refelect my employers views, source code provided can be used/re-purposed as needed but comes with no support for any production or non-production usage._
