#!/usr/bin/env zx

const argv = minimist(process.argv.slice(2), {
    alias: {
        a: 'all',
    },
    default: {
        all: false,
    },
    boolean: [
        "all",
    ],
})

let out = await $`kubectl get -n kcp-system kyma -o json`
let list = JSON.parse(out.stdout)

let data = [];

for (let kyma of list.items) {
    let item = {
        name: kyma.metadata.name,
        shoot: kyma.metadata.labels["kyma-project.io/shoot-name"],
        globalAccountId: kyma.metadata.labels["kyma-project.io/global-account-id"],
        subaccountId: kyma.metadata.labels["kyma-project.io/subaccount-id"],
        brokerPlanName: kyma.metadata.labels["kyma-project.io/broker-plan-name"],
    }
    let state = kyma.status.state
    let id = kyma.metadata.name
    let shoot = kyma.metadata.labels["kyma-project.io/shoot-name"]
    let cmFound = false
    let cmState = "-"
    let cmChannel = "-"
    if (Array.isArray(kyma.status.modules)) {
        for (let m of kyma.status.modules) {
            if (m.name === "cloud-manager") {
                cmFound = true
                cmState = m.state
                cmChannel = m.channel
            }
        }
    }

    if (!argv.all && !cmFound) {
        continue
    }

    data.push({id, shoot, state, cmState, cmChannel})
}

console.log(`${"ID".padStart(38)}  ${"SHOOT".padStart(13)} ${"STATE".padStart(12)} ${"CM-STATE".padStart(13)} ${"CM-CHANNEL".padStart(15)}`)
for (let item of data) {
    let id = item.id
    let shoot = item.shoot
    console.log(`${item.id.padStart(38)}  ${item.shoot.padStart(13)} ${item.state.padStart(12)}  ${item.cmState.padStart(13)} ${item.cmChannel.padStart(15)}`)
}
