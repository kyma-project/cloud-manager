# TODO: make version and repositoryTag be CLI variables
cat <<EOF
name: kyma-project.io/module/cloud-manager
version: 1.4.1
manifest: cloud-resources.kyma-project.io_cloudresources.yaml
defaultCR: cloud-resources_v1beta1_cloudresources.yaml
documentation: https://help.sap.com/docs/btp/sap-business-technology-platform/cloud-manager-module?version=Cloud
repository: https://github.com/kyma-project/cloud-manager.git
repositoryTag: 1.4.1
security: sec-scanners-config-release.yaml
manager:
  name: cloud-manager
  namespace: kyma-system
  group: apps
  version: v1beta1
  kind: Deployment
  icons:
    - name: module-icon
      link: https://raw.githubusercontent.com/kyma-project/kyma/refs/heads/main/docs/assets/logo_icon.svg
EOF
