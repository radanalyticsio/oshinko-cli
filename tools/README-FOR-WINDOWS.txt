#pull the repo

#Log into your account and create a project
#if you already have a project you want to use  then do this command:
# oc project myproject

#do this in the project
oc create sa oshinko
oc policy add-role-to-user admin system:serviceaccount:<Your Project Name>:oshinko


#go into oshinko-rest\tools in the local copy of repo (should already be there)
oc create -f server-ui-template.yaml

#check out this repo
git@github.com:radanalyticsio/oshinko-s2i.git

#cd to oshinko-s2i/pyspark/
oc create -f pysparkbuilddc.json

oc new-app --template oshinko -p OSHINKO_SERVER_IMAGE=radanalyticsio/oshinko-rest -p OSHINKO_CLUSTER_IMAGE=radanalyticsio/openshift-spark -p OSHINKO_WEB_IMAGE=radanalyticsio/oshinko-webui
