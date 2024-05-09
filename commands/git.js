function main() {
    os.exec(`git add .`);

    //let tagName = quickgo.environ.tag;
    //if (!tagName) {
    //    tagName = os.exec("git tag --sort=committerdate | tail -1");
    //}

    if (quickgo.environ.m) {
        console.info(`Committing changes with message: '${quickgo.environ.m}'`);
        os.exec(`git commit -m "${quickgo.environ.m}"`);
    } else {
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
    os.exec(pushStr);

    return Result(0, `QuickGo git command executed successfully!`);
}