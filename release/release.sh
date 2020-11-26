#!/usr/bin/env bash

ART_ROOT=${WORKDIR}/bazel-bin/release

pushd ${ART_ROOT} || exit 1

# Generate Release.unsigned.unsorted file
{
  go run github.com/v2fly/V2BuildAssist/v2buildutil gen version ${RELEASE_TAG}
  go run github.com/v2fly/V2BuildAssist/v2buildutil gen project "v2fly"
  for zip in $(find -L . -type f -name "*.zip"); do
    go run github.com/v2fly/V2BuildAssist/v2buildutil gen file ${zip}
  done
} >Release.unsigned.unsorted

# Generate Release.unsigned file
go run github.com/v2fly/V2BuildAssist/v2buildutil gen sort < Release.unsigned.unsorted > Release.unsigned
rm -f Release.unsigned.unsorted

# Test if is bleeding edge release
if [[ "$IsBleedingRelease" == true ]]; then
  # If it is a bleeding edge release
  # Prepare JSON data, create a release and get release id
  RELBODY="https://github.com/${COMMENT_TARGETTED_REPO_OWNER}/${COMMENT_TARGETTED_REPO_NAME}/commit/${RELEASE_SHA}"
  JSON_DATA=$(echo "{}" | jq -c ".tag_name=\"${RELEASE_TAG}\"")
  JSON_DATA=$(echo ${JSON_DATA} | jq -c ".name=\"${RELEASE_TAG}\"")
  JSON_DATA=$(echo ${JSON_DATA} | jq -c ".prerelease=${PRERELEASE}")
  JSON_DATA=$(echo ${JSON_DATA} | jq -c ".body=\"${RELBODY}\"")
  RELEASE_DATA=$(curl -X POST --data "${JSON_DATA}" -H "Authorization: token ${PERSONAL_TOKEN}" "https://api.github.com/repos/${UPLOAD_REPO}/releases")
  echo "Bleeding Edge Release data:"
  echo $RELEASE_DATA
  RELEASE_ID=$(echo $RELEASE_DATA | jq ".id")

  # Prepare commit comment message and post it
  echo "Build Finished" > buildcomment
  echo "https://github.com/${UPLOAD_REPO}/releases/tag/${RELEASE_TAG}" >> buildcomment
  go run github.com/v2fly/V2BuildAssist/v2buildutil post commit "${RELEASE_SHA}" < buildcomment
  rm -f buildcomment
else
  # If is a tag release then get the release id
  RELEASE_DATA=$(curl -X GET -H "Authorization: token ${PERSONAL_TOKEN}" "https://api.github.com/repos/${UPLOAD_REPO}/releases/tags/${RELEASE_TAG}")
  echo "Tag Release data:"
  echo $RELEASE_DATA
  RELEASE_ID=$(echo $RELEASE_DATA | jq ".id")
fi

function uploadfile() {
  FILE=$1
  CTYPE=$(file -b --mime-type $FILE)

  curl -H "Authorization: token ${PERSONAL_TOKEN}" -H "Content-Type: ${CTYPE}" --data-binary @$FILE "https://uploads.github.com/repos/${UPLOAD_REPO}/releases/${RELEASE_ID}/assets?name=$(basename $FILE)"
}

function upload() {
  FILE=$1
  DGST=$1.dgst
  openssl dgst -md5 $FILE | sed 's/([^)]*)//g' >>$DGST
  openssl dgst -sha1 $FILE | sed 's/([^)]*)//g' >>$DGST
  openssl dgst -sha256 $FILE | sed 's/([^)]*)//g' >>$DGST
  openssl dgst -sha512 $FILE | sed 's/([^)]*)//g' >>$DGST
  uploadfile $FILE
  uploadfile $DGST
}

# Upload all files to release assets
for asset in $(find -L . -type f -name "*.zip" -or -type f -name "*.unsigned"); do
  upload ${asset}
done
