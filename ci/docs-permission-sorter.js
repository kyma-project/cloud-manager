const { argv } = require('process');

const fs = require('fs').promises;

function isSorted(arr) {
  return arr.every((val, index) => index === 0 || val >= arr[index - 1]);
}

function extractPermissionArrays(rawFile, pathArray) {
  const jsonRegex = /```json([\s\S]*?)```/g;
  let match;
  let result = []
  while ((match = jsonRegex.exec(rawFile)) !== null) {
    const jsonContent = JSON.parse(match[1].trim());
    let target = jsonContent
    for (const elem of pathArray) {
      target = target[elem]
    }
    result.push(target)
  }
  return result
}

(async () => {

  const filePath = argv[2];
  const path = argv[3];

  const file = await fs.readFile(filePath, { encoding: 'utf-8' });

  let pathArray = []
  if (path && path != "") {
    pathArray = path.split(".")
  }

  const permissionArrays = extractPermissionArrays(file, pathArray);

  let returnCode = 0;
  permissionArrays.forEach(arr => {
    const hasBeenSorted = isSorted(arr)
    if (hasBeenSorted) {
      return;
    }

    const sorted = [...arr].sort()

    console.error("Expected sorted array, found unsorted.")
    console.error("Found:\n", JSON.stringify(arr, null, 2));
    console.error("Expected:\n", JSON.stringify(sorted, null, 2));

    returnCode = 1;
  })

  process.exit(returnCode);
})()
