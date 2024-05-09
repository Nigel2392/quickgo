function main() {
    os.exec(`git add .`);

    if (quickgo.environ.tag === true) {
        let tagName = os.exec("git tag --sort=committerdate | tail -1");
        let regex = `(?:(\d+)\.)?(?:(\d+)\.)?(?:(\d+)\.\d+)`
        let match = tagName.match(regex);
        if (!match) {
            return Result(1, `Could not find a valid tag to increment!`);
        }

        let major = parseInt(match[1]);
        let minor = parseInt(match[2]);
        let patch = parseInt(match[3]);
        patch++;
        let newTag = `${major}.${minor}.${patch}`;
        console.info(`Incrementing tag from ${tagName} to ${newTag}`);
        quickgo.environ.tag = newTag;
    }

    if (quickgo.environ.m) {
        console.info(`Committing changes with message: '${quickgo.environ.m}'`);
        os.exec(`git commit -m "${quickgo.environ.m}"`);
    } else {
        console.info(`Committing changes with default message: 'QuickGo update'`);
        os.exec(`git commit -m "QuickGo update"`);
    }

    if (quickgo.environ.tag) {
        console.info(`Tagging commit with tag ${quickgo.environ.tag}`);
        os.exec(`git tag ${quickgo.environ.tag}`);
    }

    let pushStr = `git push`;
    if (quickgo.environ.origin) {
        console.info(`Pushing changes to remote repository`);
        pushStr += ` -u origin ${quickgo.environ.origin}`;
    }
    if (quickgo.environ.tag) {
        console.info(`Pushing tags to remote repository`);
        pushStr += ` --tags`;
    }
    
    console.info(`Executing git command: ${pushStr}`)
    os.exec(pushStr);

    return Result(0, `QuickGo git command executed successfully!`);
}