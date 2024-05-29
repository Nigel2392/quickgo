function main() {
    // Example:
    // quickgo exec version v=1.0.0
    //
    // Must be run in a QuickGo project, or the directory for the project must be specified!
    // The project must have a 'quickgo.yaml' file with a 'versionMapping' context property.
    // The 'versionMapping' property must be an object with file paths as keys and regex strings as values.
    // The regex strings must have a capture group to match the current version.
    // The script will update the version in the files using the regex strings.

    
    if (!quickgo.environ.v) {
        return Fail(`QuickGo version not provided in arguments: v=<version>`);
    }

    if (!quickgo.project) {
        return Fail(`This script must be run in a QuickGo project, or the directory for the project must be specified!`);
    }

    if (!quickgo.project.context.versionMapping) {
        return Fail(`QuickGo version mapping not provided in '${fs.joinPath(quickgo.projectPath, 'quickgo.yaml')}' project.context`);
    }

    let version = quickgo.environ.v;
    let versionMapping = quickgo.project.context.versionMapping;

    // Mapping of file paths to regex strings to match and replace the current version.
    const files = Object.keys(versionMapping);
    for (let i = 0; i < files.length; i++) {
        let file = files[i];
        let regex = versionMapping[file];
        let compiled = new RegExp(regex, 'g');

        // Read the file content.
        let content = fs.readTextFile(file);

        // Check if there is a match for the version in the file.
        let match = compiled.exec(content);
        if (!match) {
            return Fail(`Version ${version} not found in file ${file} using regex '${regex}'`);
        }

        // Replace the old version with the new version.
        let oldVersion = match[1];
        console.debug(`Updating version in file ${file} from ${oldVersion} to ${version}`);
        content = content.replace(oldVersion, version);

        // Write the updated content back to the file.
        fs.writeFile(file, content);
    }    

    return Success(`Updated project '${quickgo.project.name}' to version ${version}`);
}