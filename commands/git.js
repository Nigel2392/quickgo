function main() {
    // Example:
    // quickgo exec git m="QuickGo update" origin=master tag
    // quickgo exec git tag=v1.0.0
    // ...

    let errD = os.exec("git add .");
    if (errD.error) {
        return Result(1, `Could not add files to git! ${errD.stdout}`);
    };

    if (quickgo.environ.tag === true) {
        let tagName = os.exec("git tag --sort=committerdate | tail -1").stdout;
        let regex = `v(?:(\d+)\.)?(?:(\d+)\.)?(?:(\d+)\.\d+)`
        let match = tagName.match(regex);
        if (!match) {
            return Result(1, `Could not find a valid tag to increment!`);
        }

        let major = parseInt(match[1]);
        let minor = parseInt(match[2]);
        let patch = parseInt(match[3]);
        patch++;
        let newTag = `v${major}.${minor}.${patch}`;
        console.info(`Incrementing tag from ${tagName} to ${newTag}`);
        quickgo.environ.tag = newTag;
    }

    if (quickgo.environ.m) {
        console.info(`Committing changes with message: '${quickgo.environ.m}'`);
        errD = os.exec(`git commit -m "${quickgo.environ.m}"`);
    } else {
        console.info(`Committing changes with default message: 'QuickGo update'`);
        errD = os.exec(`git commit -m "QuickGo update"`);
    }
    if (errD.error) {
        return Result(1, `Could not commit changes to git! ${errD.stdout}`);
    }

    if (quickgo.environ.tag) {
        console.info(`Tagging commit with tag ${quickgo.environ.tag}`);
        errD = os.exec(`git tag ${quickgo.environ.tag}`);
    }
    if (errD.error) {
        return Result(1, `Could not tag commit with tag ${quickgo.environ.tag}! ${errD.stdout}`);
    }

    let pushStr = `git push`;
    if (quickgo.environ.origin) {
        console.info(`Pushing changes to remote repository`);
        pushStr += ` -u origin ${quickgo.environ.origin}`;
    } else {
        console.info(`Pushing changes to default remote repository`);
        // Try to get the default remote repository, should work for most cases
        let d = os.exec("git symbolic-ref refs/remotes/origin/HEAD --short");
        if (d.error) {
            // Try fallback method
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
    
    console.info(`Executing git command: ${pushStr}`)
    errD = os.exec(pushStr);
    if (errD.error) {
        return Result(1, `Could not push changes to remote repository! ${errD.stdout}`);
    }

    return Result(0, `QuickGo git command executed successfully!`);
}