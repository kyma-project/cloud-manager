package cr

/*
The CloudResources reconciler must be tested in isolation since then CloudResource CR is deleted
then the reconciler will delete all cloud-resources.kyma-project.io CRDs. This CRD deletion
conflicts with other tests running.
*/
