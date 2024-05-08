function main() {
    if (!quickgo.environ.v) {
        return Result(1, `QuickGo version not provided in arguments: v=<version>`);
    }

    if (!quickgo.project) {
        return Result(1, `This script must be run in a QuickGo project, or the directory for the project must be specified!`);
    }

    if (!quickgo.project.context.versionMapping) {
        return Result(1, `QuickGo version mapping not provided in '${fs.joinPath(quickgo.projectPath, 'quickgo.yaml')}' project.context`);
    }

    let version = quickgo.environ.v;
    let versionMapping = quickgo.project.context.versionMapping;
    const files = Object.keys(versionMapping);
    for (let i = 0; i < files.length; i++) {
        let file = files[i];
        let regex = versionMapping[file];
        let content = fs.readTextFile(file);
        let match = content.match(regex);
        if (!match) {
            return Result(1, `Version ${version} not found in file ${file} using regex '${regex}'`);
        }

        let oldVersion = match[1];
        console.debug(`Updating version in file ${file} from ${oldVersion} to ${version}`);
        content = content.replace(oldVersion, version);
        fs.writeFile(file, content);
    }    

    return Result(0, `Updated project '${quickgo.project.name}' to version ${version}`);
}