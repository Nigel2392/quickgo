function main() {
    if (!quickgo.environ.push) {
        return Result(1, `QuickGo git command not provided in arguments: quickgo exec git push`);
    }

    os.exec(`git add .`);

    //let tagName = quickgo.environ.tag;
    //if (!tagName) {
    //    tagName = os.exec("git tag --sort=committerdate | tail -1");
    //}

    if (quickgo.environ.m) {
        console.debug(`Committing changes with message ${quickgo.environ.m}`);
        os.exec(`git commit -m "${quickgo.environ.message}"`);
    } else {
        os.exec(`git commit -m "QuickGo update"`);
    }

    if (quickgo.environ.tag) {
        console.debug(`Tagging commit with tag ${quickgo.environ.tag}`);
        os.exec(`git tag ${quickgo.environ.tag}`);
    }

    return Result(0, `QuickGo git command executed successfully!`);
}