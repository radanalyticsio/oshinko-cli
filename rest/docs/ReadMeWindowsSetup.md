Instructions for Deploying Oshinko on Windows
=====================
The steps are relatively simple. You need to have a user account on OpenShift.  

1. Go ahead and log in.
2. Now create a project, which will also log you into the project.

    $ oc new project myproject

3. If you already have a project you want to use go ahead and use that instead

    $ oc project myproject

4. Then create a service account

   $ oc create sa oshinko

   $ oc policy add-role-to-user admin system:serviceaccount:[Your Project Name]:oshinko

5. go into oshinko-rest\tools in the local copy of repo (should already be there)

    $ oc create -f server-ui-template.yaml

6. Move up two directories and then check out this repo

   $ git clone git@github.com:radanalyticsio/oshinko-s2i.git

7. Now in the new directory cd to oshinko-s2i/pyspark/ and do the following commands

    $ oc create -f pysparkbuilddc.json

    $ oc new-app --template oshinko

8. Now if you go into the web interface, go ahead and click on the URL for the Oshinko Web-UI. This brings you to the Oshinko web interface. Click deploy, pick a name for the cluster, and then set the number of workers you want in addition to the master.
