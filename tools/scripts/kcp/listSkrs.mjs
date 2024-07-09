#!/usr/bin/env zx

const argv = minimist(process.argv.slice(2), {
    alias: {
        a: "all",
        o: "output",
    },
    default: {
        all: false,
        output: "",
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
        kymaState: kyma.status ? kyma.status.state : "-",
        id: `${kyma.metadata.name}`,
        shootName: kyma.metadata.labels["kyma-project.io/shoot-name"],
        cmFound: false,
        cmState: "-",
        cmChannel: "-",
    }
    if (kyma.status && Array.isArray(kyma.status.modules)) {
        for (let m of kyma.status.modules) {
            if (m.name === "cloud-manager") {
                item.cmFound = true
                item.cmState = m.state
                item.cmChannel = m.channel
            }
        }
    }

    if (!argv.all && !item.cmFound) {
        continue
    }

    data.push(item)
}

if (argv.output === "json") {
    console.log(JSON.stringify(data, undefined, "  "))
    process.exit();
}

if (argv.output === "yaml") {
    console.log(YAML.stringify(data, undefined, "  "))
    process.exit();
}

if (argv.output === "wide") {
    console.log(`${"GLOBAL ACCOUNT".padStart(38)} ${"SUBACCOUNT".padStart(38)} ${"ID".padStart(38)}  ${"SHOOT".padStart(13)} ${"STATE".padStart(12)} ${"CM-STATE".padStart(13)} ${"CM-CHANNEL".padStart(15)}`)
} else {
    console.log(`${"ID".padStart(38)}  ${"SHOOT".padStart(13)} ${"STATE".padStart(12)} ${"CM-STATE".padStart(13)} ${"CM-CHANNEL".padStart(15)}`)
}
for (let item of data) {
    if (argv.output === "wide") {
        console.log(`${item.globalAccountId.padStart(38)} ${item.subaccountId.padStart(38)} ${item.id.padStart(38)}  ${item.shoot.padStart(13)} ${item.kymaState.padStart(12)}  ${item.cmState.padStart(13)} ${item.cmChannel.padStart(15)}`)
    } else {
        console.log(`${item.id.padStart(38)}  ${item.shoot.padStart(13)} ${item.kymaState.padStart(12)}  ${item.cmState.padStart(13)} ${item.cmChannel.padStart(15)}`)
    }
}
