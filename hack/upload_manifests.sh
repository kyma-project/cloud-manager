
REPOSITORY=${1}
MANIFEST_FILEPATH=${2}
EXAMPLE_FILEPATH=${3}
RELEASE_ID=${4}

echo "Uploading manifests"

uploadFile() {
	filePath=${1}
	ghAsset=${2}

	response=$(curl -s -o output.txt -w "%{http_code}" \
		--request POST --data-binary @"$filePath" \
		-H "Authorization: token $GITHUB_TOKEN" \
		-H "Content-Type: text/yaml" \
		$ghAsset)
	if [[ "$response" != "201" ]]; then
		echo "Unable to upload the asset ($filePath): "
		echo "HTTP Status: $response"
		cat output.txt
		exit 1
	else
		echo "$filePath uploaded"
	fi
}

UPLOAD_URL="https://uploads.github.com/repos/${REPOSITORY}/releases/${RELEASE_ID}/assets"

# TODO: figure out finalized file names
uploadFile $MANIFEST_FILEPATH "${UPLOAD_URL}?name=cloud-resources.kyma-project.io_cloudresources.yaml"
uploadFile $EXAMPLE_FILEPATH "${UPLOAD_URL}?name=cloud-resources_v1beta1_cloudresources-default-cr.yaml"