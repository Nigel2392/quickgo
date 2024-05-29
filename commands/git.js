function main() {
    // Example:
    // quickgo exec git m="QuickGo update" origin=master tag
    // quickgo exec git tag=v1.0.0
    // ...

    // Make git aware of all changes.
    let errD = os.exec("git add .");
    if (errD.error) {
        return Fail(`Could not add files to git! ${errD.stdout}`);
    };

    // Get the latest tag, increment the patch version and set it as the new tag.
    // Only if the tag flag is set.
    if (quickgo.environ.tag === true || quickgo.environ.tag === "true") {
        let tagName = os.exec("git tag --sort=committerdate | tail -1").stdout;
        let regex = `v(?:(\d+)\.)?(?:(\d+)\.)?(?:(\d+)\.\d+)`
        let match = tagName.match(regex);
        if (!match) {
            return Fail(`Could not find a valid tag to increment!`);
        }

        let major = parseInt(match[1]);
        let minor = parseInt(match[2]);
        let patch = parseInt(match[3]);
        patch++;
        let newTag = `v${major}.${minor}.${patch}`;
        console.info(`Incrementing tag from ${tagName} to ${newTag}`);
        quickgo.environ.tag = newTag;
    }

    // Commit changes with the message provided by the user or the default message.
    if (quickgo.environ.m) {
        console.info(`Committing changes with message: '${quickgo.environ.m}'`);
        errD = os.exec(`git commit -m "${quickgo.environ.m}"`);
    } else {
        console.info(`Committing changes with default message: 'QuickGo update'`);
        errD = os.exec(`git commit -m "QuickGo update"`);
    }
    if (errD.error) {
        return Fail(`Could not commit changes to git! ${errD.stdout}`);
    }

    // Tag the commit with the tag provided by the user or the latest tag.
    if (quickgo.environ.tag) {
        console.info(`Tagging commit with tag ${quickgo.environ.tag}`);
        errD = os.exec(`git tag ${quickgo.environ.tag}`);
    }
    if (errD.error) {
        return Fail(`Could not tag commit with tag ${quickgo.environ.tag}! ${errD.stdout}`);
    }

    // Push changes to the remote repository.
    // If the origin flag is set, use it as the remote repository.
    // If the tag flag is set, push tags to the remote repository.
    // By default runs: git push -u origin <default-origin>
    let pushStr = `git push`;
    if (quickgo.environ.origin) {
        console.info(`Pushing changes to remote repository`);
        pushStr += ` -u origin ${quickgo.environ.origin}`;
    } else {
        console.info(`Pushing changes to default remote repository`);
        // Try to get the default remote repository, should work for most cases
        let d = os.exec("git symbolic-ref refs/remotes/origin/HEAD --short");
        if (d.error) {
            // Log a warning, let git handle the default origin
            console.warn(`Could not determine default remote repository, trying fallback method ${d.stdout}`);
        } else {
            // Trim the origin
            let remote = d.stdout.trim();
            remote = remote.replace("origin/", "");
            pushStr += ` -u origin ${remote}`;
        }
    }
    if (quickgo.environ.tag) {
        console.info(`Pushing tags to remote repository`);
        pushStr += ` --tags`;
    }
    
    // Actually execute the push command.
    console.info(`Executing git command: ${pushStr}`)
    errD = os.exec(pushStr);
    if (errD.error) {
        return Fail(`Could not push changes to remote repository! ${errD.stdout}`);
    }

    return Success(`QuickGo git command executed successfully!`);
}