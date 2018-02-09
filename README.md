## GJPush
This is an experimental tool to upload builds to Game Jolt from the command line.
It supports resuming an upload and creating builds/releases on the fly!  
![](https://i.imgur.com/3kfk7Wf.gif)

### Usage
1. Download from https://github.com/gamejolt/joltron/releases
2. Run from cmd / terminal:
    ```
    gjpush file
    ```
    where `file` is the path to the build you want to upload

### Options
GJPush will prompt you for additionanl info it needs, but you can automate it by passing it in through _options_:
```
-t, --token=TOKEN        Your service API authentication token
-g, --game=GAME-ID       The game ID
-p, --package=PACKAGE    The package ID
-r, --release=VERSION    The release version to attach the build file to
-b, --browser            Upload a browser build. By default uploads a desktop build.
```

1. __Token__ is your "password" to the tool. At the moment only testers are given one, but when the tool is launched publicly, you could get a token from your dashboard.
2. The __game ID__ and __package ID__ are available in the url of the manage game package page, for example:
    ![like so](https://i.imgur.com/HcePzxN.png)
3. The __release__ is [semver](https://semver.org/), looks like 1.2.3

Once all required options are specified, the upload can happen in a single command:
![](https://i.imgur.com/r9kteuT.gif)

### Example
Pushing version 2.0.1 as a desktop build for a game with ID 1 and package ID 2:
```
gjpush -t my-token -g 1 -p 2 -r 2.0.1 game.exe
```

To push a browser build, simply add the `-b` options:
```
gjpush -t my-token -g 1 -p 2 -r 2.0.1 -b game_html.zip
```

### Want to help test?
Awesome! Send me an email at yariv@gamejolt.com with:
1. A link to your profile
2. A link to a game page
