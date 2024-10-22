# Busola UI Documentation

Kyma manages its modules' user interfaces (UIs) using ConfigMaps. Each ConfigMap has five parts:


* `general` - used to describe the CRD the UI should be looking for as well as a description of the resource in the case none exists on the cluster.
* `list` - used to set which columns to display in the table view
* `detail` - used to display the detailed view of a single resource
* `form` - used to define a GUI form when creating a new resource or editing an existing one. 
* `translations` - used to define elements in different languages (English, German, etc.)

Examples and important notes of each part are below

### General
```yaml
resource:
  kind: GcpNfsVolume
  group: cloud-resources.kyma-project.io
  version: v1beta1
urlPath: gcpnfsvolumes
name: GCP NFS Volumes
scope: namespace
category: Storage
icon: shelf
description: >-
  GcpNfsVolume description here
```
It is imperative that `resource.kind` `resource.group` and `resource.version` matches its `CustomResourceDefinition`. If
there are no matches, Busola doesn't render the UI and path rendering the resource inaccessible in Busola.


### List
```yaml
- source: spec.fileShareName
  name: spec.fileShareName
  sort: true
- source: spec.location
  name: spec.location
  sort: true
- source: spec.tier
  name: spec.tier
  sort: true
- source: status.state
  name: status.state
  sort: true
```

The `source` field is where Busola pulls information from the monitored custom resource (CR).

The `name` field is the human-readable name for the field. This field looks up the value in `Translation` and replaces it with its found value.
If no value is found in `Translation`, it displays as is.

For example, if we have `name: spec.location`, it goes to [translations](#translations), looks up `spec.location` and replaces it with `Location`.

[Official List Documentation](https://github.com/kyma-project/busola/blob/main/docs/extensibility/20-list-columns.md)

[Official List and Detail Widgets](https://github.com/kyma-project/busola/blob/main/docs/extensibility/50-list-and-details-widgets.md)

### Detail
```yaml
body:
    - name: configuration
      widget: Panel
      source: spec
      children:
        - name: spec.fileShareName
          source: fileShareName
          widget: Labels
        - name: spec.capacityGb
          source: capacityGb
          widget: Labels
        - name: spec.location
          source: location
          widget: Labels
        - name: spec.tier
          source: tier
          widget: Labels
    - name: status
      widget: Panel
      source: status
      children:
        - widget: Labels
          source: state
          name: status.state
```

`widget` refers to the component interfaces built into Busola. See link below for official documentation on widgets.

`children` are an array of child widgets. Note, the `source` in each child is relative to its parent.

[Official Detail Documentation](https://github.com/kyma-project/busola/blob/main/docs/extensibility/30-details-summary.md)

[Official List and Detail Widgets](https://github.com/kyma-project/busola/blob/main/docs/extensibility/50-list-and-details-widgets.md)

### Form
```yaml
- path: spec.capacityGb
  simple: true
  name: spec.capacityGb
  required: true
- path: spec.fileShareName
  simple: true
  name: spec.fileShareName
  required: true
- path: spec.location
  simple: true
  name: spec.location
  required: true
- path: spec.tier
  simple: true
  name: spec.tier
  required: true
```

`simple` is a boolean used to display the field in the simple fom. By default, it is `false`

[Official Forms Documentation](https://github.com/kyma-project/busola/blob/main/docs/extensibility/40-form-fields.md)

<a id="translations"></a>
### Translations
```yaml
en:
  spec.tier: Tier
  spec.location: Location
  spec.fileShareName: File Share Name
  spec.capacityGb: Capacity (Gb)
  spec.ipRange: IP Range
  configuration: Configuration
  status.state: State
  status: Status
```

`translations` is optional, but it supports languages formatted for [i18next](https://www.i18next.com/). Translations prettify the `name` field of the aformentioned sections.
They are key-value pairs.


[Official Translations Documentation](https://github.com/kyma-project/busola/blob/main/docs/extensibility/translations-section.md)


# Generating the ConfigMap

The generation of a ConfigMap is handled using `kustomize`. Kustomize  combines general, list, detail, form, and translations
into a single ConfigMap.

Because each CustomResource has its own UI that must be generated, there is a single macro in the `Makefile` to generate all the UIs.

`make build_ui`

For every new UI component, a new `kustomize` command should be added to the Makefile `build_ui` macro.

[Official kustomize documentation](https://kustomize.io/)

# Deploying to your desired provider (AWS, GCP, Azure)

Currently, the pipeline looks at the presence of ConfigMaps files in specific directories to deploy. They are the following:

AWS - `config/dist/skr/crd/bases/providers/aws`

Azure - `config/dist/skr/crd/bases/providers/azure`

GCP - `config/dist/skr/crd/bases/providers/gcp`

Instead of manually copying over the ConfigMap into the specified directory, use the bash script `sync.sh`  

Like the `Makefile`, any new UI component should be added to `sync.sh`.

# Auto-updating ConfigMaps

The following annotation should be added to every ConfigMap:

```yaml
metadata:
  annotations:
    cloud-resources.kyma-project.io/version: v0.0.1
```

Increment the version number every time a change is made to the ConfigMap. SKR Looper will compare the desired and existing version. If there is a difference, it will update the ConfigMap.

# Testing your UI changes without using the pipeline

Because the UI is dictated by ConfigMaps, you can deploy directly to a cluster with kyma installed. Just ensure your ConfigMap goes into the `kyma-system`
namespace. 

You can then `kubectl apply` your ConfigMap.

If a matching CustomResource is found, Busola renders the UI and path.


You can also spin up a local cluster with kyma installed via k3d.

Official documentation and instructions can be found [here](https://kyma-project.io/#/02-get-started/01-quick-install).



# Helpful Links
[Translations](https://github.com/kyma-project/busola/blob/main/docs/extensibility/translations-section.md)