# notes

# This note will have random though, REX of my journey so i can paste it into an AI to latter create ADR, documentations and other stuff for this project 

**/!\ DONT READ THIS NOTE !**

**/!\ HERE BE DRAGONS, I WROTE THAT THING AT 4am !**

As a platform engineer who discovered Nomad in my work at iAdvize, and most particularly the deployment process in medium size company (50 devs)
by working on our internal deployment tool that turn deployment definition into nomad jobs.
I recently got interested by iac, (cause i love gitops philosophy and the every thing as code philosophy, (Iv been using nix for 3 year now...)), Kubernetes, Service meshing, and most particularly by SRE

i read the "Becoming SRE book from David N Blankman" and now I want to become an SRE. But to become an SRE you must be a senior...
especially if you are targeting big tech company that have big production systems like me .... 🫠

I dont want to get a job of Backend developper for 3 years, then work my way up into DevOps, then platform engineering again (since I already started as an apprentice platform engineer) and finally
maybe get to SRE..

I want SRE as fast as possible.

Soooo in order to work on my objective, i decided that I will have a new fun project, which is "create the most production ready infrastructure possible on AWS" to achieve such a goal
i will :
- use the knowledge i got at iAdvize
- use the knowledge i got by reading SRE/DevOPS books
- unlock the knowledge i dont have by asking random AI models "how to make things better if I was a senior SRE/Platform Engineer/DevOPS

Obviously in this project the term of "Overscoping" dont exist, and we dont talk about "overengineering" but rather about "making thing scale :)"
Obviously this infrastructure i'am creating will serve only 2 purposes: 
- help me learn new things
- help me get a job (maybe 🤡)

Here are things I want to do properly in this project : 

- Respect gitops philosophy
- iac
- kubernetes
- Make things right for a team of 5-10 people
- Release engineering
- Observability
- Use managed service a lot
- Scaling
- SLA/SLO/SLI

Here are the things i didnt focused too much (But would have been great to do if i had time for it):

- Security
- Compliance
- FinOps
- Secrets management
- IDP
- Self service
- Docs

# Though along the way

I will not use terraform in an effort to use as much open source alternative as possible (except managed service from AWS)
backend tfstate on a bucket is a must have and is always the first thing i like to do when i start a new infra project

learning what a VPC and IAM is was fun.

I was blown away by how easy it was to provision a kubernetes cluster with some nodes using AWS (EKS) + openTofu !

I lean that there kind of was 2.5 possible options to do automatic node provisioning :
- Fargate
- Node Group 
    - Node Group with Karpenter

i like the karpenter approach of that (if i understand correctly) could be used to only use spot instances in a immutable way of dealing with compute (remind me of nixos and stuff :D)
In a first iteration i decided to only go with a Simple one instance type managed node group, but i added in bonus a ticket to try karpenter !


When i started to create a CI for terraform I came a cross a problem that was "how to make my ci have the right to get the state and to provision stuff ?" I found in the AWS Docs
that i could use OIDC for it and so i setup it for my github action. this was not too complicated, but I think there is a lot to learn more about it.
(i also decided to use OIDC because i didnt wanted to add my aws credentials in env var of my repos) -> my lead tech senior infra said that it was super cool ! and that they should be doing the same at my company !!!


I also try to make my managed service as much multi AZ as possible (with at least 2 AZ each time) so that I, like in real production grade infrasturcture avoid SPOFs.


Creating a CI was fun, I started with creating a CI for my terraform code 
- tofu init, tofu format, tofu validate, tofu plan upon commiting in a branch, and if that diff contains files from the iac/ folder.
- tofu validate, tofu apply upon merging to master and if taht diff contains files from the iac/ folder.

(i also made it so the tofu apply was pushed in the Pull request as a comment)

Then i created 2 micro services in golang:
- one micro service (api) that would be responsible to serve an API and upon receiving a get request would increment a counter in a postgresql database, and then return a resposne with the new value of the counter.
- another micro service (web-server) that would serve a static html file with a button that will send a get request to its own golang backend code API and then this golang backend code will call the (api) micro service using a get request.

Once this was done I created a CI for these golang code that : 
- lint & format golang code for each micro services using golangci lint and gofmt AND lint the dockerfile using the standard hadolint (i used external github actions for each of these steps)
- then build the docker image, tag it, push it to ECR

I also create a require-label workflow that would fail if the PR was not tagged with either "Major, minor, or patch" so that these tag would be used for incrementing the release version
each time we rebase onto master (so we also needed a workflow to create that release on github)

And finally, when commiting we push a image to ECR with the tag being the branch name and if rebasing into master, we would be using the release tag generated (just like how it is done at iadvize)

I terraformed the ECR, made i MUTABLE (because each commit on a branch should overwrite the previous tag pushed for that branch)
and I also hard coded the repository names for each of my micro services in my erc.tf file

I was unsure about this and was questioning myself if i should have made it though github action (so that new registry could have been created on the go, without having to modify the iac code) but decided to do it all though iac since im destroy everything when i stop working on my project to not get gigantesque AWS bills

also added some policy rules to my eks to auto clean untagged images and only keep the 50 more recent images
(AND I USED A for_each for the repository name yay! to avoid redundancy)


I also had a LONGGGGGG time figuring out the right way to write a dockerfile for my apps to be deployed on my EKS managed nodes that was amazon linux arm64

i decided to use chainguard open source base image both for the golang builder and for the static runner to make my image secure and as thin as possible (so save some space on ECR)



Writing deployment deifintion using yaml vanilla syntax of kubernetes wasnt too complicated for my micro-services
for each micro services i have a deployment.yaml and a service.yaml

for the moment the deployment yaml have hardcoded branch image tag for the docker image, but i plan to use argocd or helm to fix this problem (though i dont know yet how it will work 🫠)


for the web-server micro-service i also have a service that is of type Load Balancer because i wanted public traffic to be able to see the index.html from my computer.



I had a lot lot of problems with creating my RDS and making it reachable for my (api) micro service

- First i had to understand how to provision as code a RDS and what was the difference with Aurora RDS and standard RDS
- Then i need to create subnet and security group
- (note that i didnt multi AZed my RDS)
- Then i needed to pass the db_url to my (api) micro service env var, but i didnt wanted to keep the  original db_url because each time im destroying + reapply my whole infrastructure this url changes
so i created a private Route53 dns that is "postgresql.justalternate-eks-cluster.internal" and then passed this as env var.

BUT THIS WAS NOT WORKING my (api) service could not reach my RDS (timeout) despite it being in the same VPC and all!!

so i created a private subnet, changed some settings in the iac, but nothing was working.
by debugging and prompting an LLM, i finally understood that when I provision my EKS, AWS create a default security group, that i needed to aslo add to my ingress of the security group of my RDS !!!!
this was a very big trap that took me hours to understand !!!

(note that i didnt setup any password and used this very bad default config :
  db_name           = "my_postgres"
  engine            = "postgres" # We will not use Aurora so we can create our own backup and restoration strategie
  engine_version    = "17.6"
  instance_class    = "db.t4g.micro"
  username          = "username"
  password          = "password")

(security is not a priority for this project..., i know thats bad...)


And now finnaly I have my 2 apps that can communicate between each other on my EKS builded by my CI pushed to my ECR and with a RDS that is working !

Now im currently setting up healthchecks, statup probe, readiness, liveness on each micro services !

After talking with my lead tech infra at iavize, i realized it was possible to use spot in a managed node group thanks to "launch template", so i added this step in the bonus section of my todo list.

Talking to my lead tech infra also made me realized that AMP/AMG isnt the best options in terms of pricing and that most company because of that, prefer to host their own observability stack themselves
and most particularly for prometheus and grafana

From this point of view there is multiple way to get this monitoring stack up and running

1) the statup way: put every thing in our EKS but make the monitoring stack on a different node group and its own namespace
2) the scale up way: create a 2nd cluster to host the monitoring stack
3) the giant tech enterprise way: use seperate cluster and a separate account

To keep thing no too complicated i decided to go with option 1) (this is also what my company does btw)

While setting up prometheus, grafana and loki, i decided to use the very good helm chart that is provided by the kube-prometheus-stack from the prometheus community
I also needed to add an EKS addon for block storage (ebs csi), so i updated my iac.

in a first iteration i created a bash script to execute that would use helm to install the monitoring stack in my cluster.
but this was tedious since im constantly applying and destroying my whole infrastructure
so i decided to move the Observability stack install into terraaform by using the helm provider.

In order to allow grafana to access cloudwatch source, i use oidc, which was basically the same config i used for the CI but with additional IAM role for grafana


