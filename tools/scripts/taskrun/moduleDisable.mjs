#!/usr/bin/env zx

let out


out = await $`kubectl get -n kyma-system kyma default -o json | jq '.spec.modules | map(.name == "cloud-manager") | index(true)'`
const idx = `${out.stdout}`.trim()

console.log("CM Index: ", idx);


out = await $`kubectl patch -n kyma-system kyma default --type=json -p="[{'op': 'remove', 'path': '/spec/modules/${idx}'}]"`
console.log(out.stdout);

