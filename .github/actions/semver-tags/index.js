const semver = require('semver')
const { exec } = require("child_process");
const { mainModule } = require('process');
const core = require('@actions/core');
const github = require('@actions/github');



async function main() {

    try {
        const imageToTag  = core.getInput('image-to-tag');
        const imageVersion = core.getInput('image-version');

        let image = imageToTag.split(":")

        let imageName = image[0]
        let uniqueTag = image[1]

        core.info(`Image name: ${imageName}`);
        core.info(`Image tag: ${uniqueTag}`);

        core.info(`Image to tag: ${imageToTag}`);
        core.info(`Version: ${imageVersion}`);
        
        let allTags = await getAllTags(imageName)
        tagsToSet = getAllVersionTags(imageVersion, allTags)

        core.info(`Tags to set: ${tagsToSet}`);

        for (tag of tagsToSet) {
            core.info(`Setting tag "${tag}" on "${imageName}:${uniqueTag}"`);
            await pushTag(imageName, uniqueTag, tag)
        }
  
        const payload = JSON.stringify(github.context.payload, undefined, 2)
        core.debug(`The event payload: ${payload}`);
      } catch (error) {
        core.error(error.message);
        core.setFailed(error.message);
      }


}



/**
 * Executes a shell command and return it as a Promise.
 * @param cmd {string}
 * @return {Promise<string>}
 */
function execShellCommand(cmd) {
    const exec = require('child_process').exec;
    core.debug(`Running command: ${cmd}` )
    return new Promise((resolve, reject) => {
        exec(cmd, (error, stdout, stderr) => {
            if (error) {
                //console.warn(error);
                core.warning(error);
                reject(error)
            }

            if (stderr) core.debug(stderr)
            if (stdout) core.debug(stdout)

            resolve(stdout? stdout : stderr);
        });
    });
}


async function getAllTags(ref) {
    let output = await execShellCommand(`gcloud alpha artifacts docker tags list ${ref} --format=json --quiet`) 
   
    let tags = JSON.parse(output).map(function (item) {
        return item.tag.split('/').pop();
      });

    return tags
}

async function pushTag(ref, tag, newtag) {
    let cmd = `gcloud alpha artifacts docker tags add ${ref}:${tag} ${ref}:${newtag} --quiet`
    let output = await execShellCommand(cmd) 

    return output
}




function getAllVersionTags(version,allVersions){
    var tags = []

    tags.push(version)

    if (isHighestMinor(version, allVersions)) {
        tags.push( semver.major(version) + "." + semver.minor(version))
    }
    if (isHighestMajor(version, allVersions)) {
        tags.push(semver.major(version))
    }
    if (isHighestVersion(version, allVersions)) {
        tags.push("latest")
    }

    return tags
}

function tagExists(version, allVersions) {
    if(allVersions.indexOf(version) !== -1){
        return true
    } 

    return false
}


function isHighestMajor(version, allVersions) {
    range = semver.major(version) + ".*"
    if (semver.maxSatisfying(allVersions.concat(version), range) == version) {
        return true 
    }
    return false 
}

function isHighestMinor(version, allVersions) {
    range = semver.major(version) + "." + semver.minor(version) + ".*"
    if (semver.maxSatisfying(allVersions.concat(version), range) == version) {
        return true 
    }
    return false 
}

function isHighestVersion(version, allVersions) {
    range =  "*"
    if (semver.maxSatisfying(allVersions.concat(version), range) == version) {
        return true 
    }
    return false 
}



main();